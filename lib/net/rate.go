package net

import (
    "context"
    "time"
)

type RateLimiter struct {
    WindowSize time.Duration
    MaxEvents int
    count int
    windowStart time.Time
}

func (limiter *RateLimiter) AddAndLimit(quit context.Context) {
    limiter.count += 1
    if time.Since(limiter.windowStart) > limiter.WindowSize {
        limiter.windowStart = time.Now()
        limiter.count = 1
    }

    if limiter.count > limiter.MaxEvents {
        select {
            case <-quit.Done():
            // sleep for the duration of the window
            case <-time.After(limiter.WindowSize - time.Since(limiter.windowStart)):
        }
    }
}

