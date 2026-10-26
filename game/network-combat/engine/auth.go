package engine

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed key/*
var embeddedAuthFiles embed.FS

func readClientAuthKey(files fs.FS) (string, error) {
	data, err := fs.ReadFile(files, "key/key.txt")
	if err != nil {
		return "", fmt.Errorf("load client auth key: %w", err)
	}
	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", fmt.Errorf("client auth key is empty")
	}
	return key, nil
}

func clientAuthKey() (string, error) { return readClientAuthKey(embeddedAuthFiles) }
