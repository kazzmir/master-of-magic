package combat

import (
    "context"
    "fmt"
    "testing"
    "testing/synctest"
    "time"
)

func TestRateLimiter(test *testing.T) {
    for _, maxEvents := range []int{1, 3, 300} {
        test.Run(fmt.Sprintf("%d_packets", maxEvents), func(test *testing.T) {
            synctest.Test(test, func(test *testing.T) {
                limiter := RateLimiter{
                    WindowSize: time.Second,
                    MaxEvents: maxEvents,
                }
                sendWindow := func() {
                    test.Helper()
                    start := time.Now()
                    for packet := 1; packet <= maxEvents; packet++ {
                        limiter.AddAndLimit(context.Background())
                        if elapsed := time.Since(start); elapsed != 0 {
                            test.Fatalf("Packet %d of %d was rate limited for %v", packet, maxEvents, elapsed)
                        }
                    }

                    // Packet N+1 must wait for the remainder of this window.
                    done := make(chan struct{})
                    go func() {
                        limiter.AddAndLimit(context.Background())
                        close(done)
                    }()
                    synctest.Wait()
                    select {
                        case <-done:
                            test.Fatal("Packet N+1 was not rate limited")
                        default:
                    }

                    time.Sleep(limiter.WindowSize - time.Nanosecond)
                    synctest.Wait()
                    select {
                        case <-done:
                            test.Fatal("Packet N+1 was released before the window ended")
                        default:
                    }

                    time.Sleep(time.Nanosecond)
                    synctest.Wait()
                    select {
                        case <-done:
                        default:
                            test.Fatal("Packet N+1 was not released when the window ended")
                    }
                }

                sendWindow()
                // Move past the boundary and verify a fresh N-packet allowance.
                time.Sleep(time.Nanosecond)
                sendWindow()
            })
        })
    }
}
