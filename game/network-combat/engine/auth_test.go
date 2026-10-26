package engine

import (
	"errors"
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestReadClientAuthKey(t *testing.T) {
	files := fstest.MapFS{"key/key.txt": {Data: []byte("  test-key\n")}}
	key, err := readClientAuthKey(files)
	if err != nil || key != "test-key" {
		t.Fatal("embedded auth key not loaded/trimmed")
	}
	key, err = readClientAuthKey(fstest.MapFS{"key/empty": {}})
	if key != "" || !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("missing key not reported", err)
	}
	key, err = readClientAuthKey(fstest.MapFS{"key/key.txt": {Data: []byte(" \n")}})
	if key != "" || err == nil {
		t.Fatal("empty key accepted")
	}
}
