package main

import (
	"errors"
	"flag"
	"path/filepath"
	"strings"
	"testing"
)

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
