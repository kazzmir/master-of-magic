package main

import (
	"errors"
	"flag"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	if got := formatTime(time.Hour + 2*time.Minute + 3*time.Second + 400*time.Millisecond); got != "01:02:03.4" {
		t.Fatalf("formatted time = %q", got)
	}
	if got := formatTime(0); got != "00:00:00.0" {
		t.Fatalf("formatted zero = %q", got)
	}
}

func TestHelp(t *testing.T) {
	if err := run([]string{"-h"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected help, got %v", err)
	}
}

func TestInvalidArguments(t *testing.T) {
	for _, args := range [][]string{nil, {"1", "2"}, {"abc"}, {"-1"}, {"--", "-1"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if err := run(args); err == nil {
				t.Fatal("expected an argument error")
			}
		})
	}
}

func TestInvalidPolyphony(t *testing.T) {
	for _, budget := range []string{"-1", "0", "7", "257", "4294967360"} {
		if err := run([]string{"-polyphony", budget, "115"}); err == nil || !strings.Contains(err.Error(), "polyphony") {
			t.Fatalf("budget %s: expected polyphony error, got %v", budget, err)
		}
	}
}

func TestExplicitArchiveMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "MUSIC.LBX")
	for _, budget := range []string{"8", "64", "128", "256"} {
		if err := run([]string{"-polyphony", budget, "-lbx", path, "115"}); err == nil || !strings.Contains(err.Error(), "open music archive") {
			t.Fatalf("budget %s: expected archive open error, got %v", budget, err)
		}
	}
}
