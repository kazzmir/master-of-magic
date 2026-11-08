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

func TestExplicitArchiveMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "MUSIC.LBX")
	if err := run([]string{"-lbx", path, "115"}); err == nil || !strings.Contains(err.Error(), "open music archive") {
		t.Fatalf("expected archive open error, got %v", err)
	}
}
