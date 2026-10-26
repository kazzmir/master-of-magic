package system

import (
	"errors"
	"io"
	"io/fs"
	"reflect"
	"testing"
)

func requireMemFile(t *testing.T, filesystem fs.FS, path, want string) {
	t.Helper()
	got, err := fs.ReadFile(filesystem, path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("read %q = %q, want %q", path, got, want)
	}
}

func TestMemFSCreate(t *testing.T) {
	filesystem := NewMemFS()
	writer, err := filesystem.Create("save")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"hello", " world"} {
		if n, err := io.WriteString(writer, text); err != nil || n != len(text) {
			t.Fatalf("write = (%d, %v), want (%d, nil)", n, err, len(text))
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	requireMemFile(t, filesystem, "save", "hello world")
	if _, err := writer.Write([]byte("closed")); !errors.Is(err, fs.ErrClosed) {
		t.Fatalf("write after close: %v", err)
	}
	if err := writer.Close(); !errors.Is(err, fs.ErrClosed) {
		t.Fatalf("second close: %v", err)
	}

	writer, err = filesystem.Create("save")
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	requireMemFile(t, filesystem, "save", "")
}

func TestMemFSWriteFileAndOpen(t *testing.T) {
	filesystem := NewMemFS()
	if err := filesystem.WriteFile("file", []byte("abcdef"), 0640); err != nil {
		t.Fatal(err)
	}
	first, err := filesystem.Open("file")
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	info, err := first.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Name() != "file" || info.Size() != 6 || info.Mode() != 0640 || info.IsDir() {
		t.Fatalf("unexpected file info: %+v", info)
	}
	buf := make([]byte, 2)
	if _, err := io.ReadFull(first, buf); err != nil || string(buf) != "ab" {
		t.Fatalf("initial read = (%q, %v)", buf, err)
	}
	// Another handle must start at the beginning, not share the read offset.
	requireMemFile(t, filesystem, "file", "abcdef")
	seeker, ok := first.(io.Seeker)
	if !ok {
		t.Fatal("opened file does not implement io.Seeker")
	}
	if offset, err := seeker.Seek(-2, io.SeekEnd); err != nil || offset != 4 {
		t.Fatalf("seek = (%d, %v), want (4, nil)", offset, err)
	}
	if got, err := io.ReadAll(first); err != nil || string(got) != "ef" {
		t.Fatalf("read after seek = (%q, %v)", got, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := first.Read(buf); !errors.Is(err, fs.ErrClosed) {
		t.Fatalf("read after close: %v", err)
	}
	if _, err := first.Stat(); !errors.Is(err, fs.ErrClosed) {
		t.Fatalf("stat after close: %v", err)
	}
	if _, err := seeker.Seek(0, io.SeekStart); !errors.Is(err, fs.ErrClosed) {
		t.Fatalf("seek after close: %v", err)
	}
	if err := filesystem.WriteFile("file", []byte("x"), 0640); err != nil {
		t.Fatal(err)
	}
	requireMemFile(t, filesystem, "file", "x")
}

func TestMemFSDirectories(t *testing.T) {
	filesystem := NewMemFS()
	for _, path := range []string{".", "saves/nested", "saves/nested"} {
		if err := filesystem.MkdirAll(path, 0750); err != nil {
			t.Fatalf("mkdir %q: %v", path, err)
		}
	}
	for _, name := range []string{"z", "a"} {
		if err := filesystem.WriteFile("saves/"+name, []byte(name), 0600); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := fs.ReadDir(filesystem, "saves")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
		info, err := entry.Info()
		if err != nil {
			t.Fatal(err)
		}
		wantDir := entry.Name() == "nested"
		if entry.IsDir() != wantDir || info.IsDir() != wantDir || entry.Type().IsDir() != wantDir {
			t.Fatalf("incorrect directory type for %q", entry.Name())
		}
		if wantDir && info.Mode().Perm() != 0750 {
			t.Fatalf("directory permissions = %v, want 0750", info.Mode())
		}
	}
	if !reflect.DeepEqual(names, []string{"a", "nested", "z"}) {
		t.Fatalf("directory entries = %v", names)
	}

	handle, err := filesystem.Open("saves")
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Close()
	reader := handle.(fs.ReadDirFile)
	for _, want := range names {
		batch, err := reader.ReadDir(1)
		if err != nil && !errors.Is(err, io.EOF) {
			t.Fatal(err)
		}
		if len(batch) != 1 || batch[0].Name() != want {
			t.Fatalf("ReadDir(1) = %v, want %q", batch, want)
		}
	}
	if batch, err := reader.ReadDir(1); len(batch) != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("exhausted ReadDir = (%v, %v)", batch, err)
	}
	if batch, err := reader.ReadDir(-1); len(batch) != 0 || err != nil {
		t.Fatalf("exhausted ReadDir(-1) = (%v, %v)", batch, err)
	}

	sub, err := filesystem.Sub("saves")
	if err != nil {
		t.Fatal(err)
	}
	requireMemFile(t, sub, "a", "a")
	if _, err := fs.ReadFile(sub, "saves/a"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("subtree should use relative paths: %v", err)
	}
}

func TestMemFSErrors(t *testing.T) {
	filesystem := NewMemFS()
	for _, path := range []string{"", "/absolute", "../parent", "a/../b", "a//b", "a/"} {
		t.Run(path, func(t *testing.T) {
			if _, err := filesystem.Open(path); !errors.Is(err, fs.ErrInvalid) {
				t.Errorf("Open: %v", err)
			}
			if _, err := filesystem.Create(path); !errors.Is(err, fs.ErrInvalid) {
				t.Errorf("Create: %v", err)
			}
			if err := filesystem.WriteFile(path, nil, 0600); !errors.Is(err, fs.ErrInvalid) {
				t.Errorf("WriteFile: %v", err)
			}
			if err := filesystem.MkdirAll(path, 0700); !errors.Is(err, fs.ErrInvalid) {
				t.Errorf("MkdirAll: %v", err)
			}
		})
	}
	if _, err := filesystem.Open("missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("open missing: %v", err)
	}
	if _, err := filesystem.Create("missing/file"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("create without parent: %v", err)
	}
	if err := filesystem.WriteFile("file", nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := filesystem.MkdirAll("file/child", 0700); !errors.Is(err, fs.ErrInvalid) {
		t.Errorf("mkdir through file: %v", err)
	}
	if _, err := filesystem.Open("file/child"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("open through file: %v", err)
	}
	if err := filesystem.MkdirAll("dir", 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := filesystem.Create("dir"); !errors.Is(err, fs.ErrExist) {
		t.Errorf("create over directory: %v", err)
	}
	if _, err := filesystem.Create("."); !errors.Is(err, fs.ErrInvalid) {
		t.Errorf("create root: %v", err)
	}
	if err := filesystem.WriteFile(".", nil, 0600); !errors.Is(err, fs.ErrInvalid) {
		t.Errorf("write root: %v", err)
	}
}

func TestMemFSOpenHook(t *testing.T) {
	denied := errors.New("hook denied access")
	var calls []string
	filesystem := NewMemFS(WithOpenHook(func(path string, content []byte, origErr error) ([]byte, error) {
		calls = append(calls, path)
		switch path {
		case "existing":
			if origErr != nil || string(content) != "original" {
				t.Errorf("hook received (%q, %v), want (original, nil)", content, origErr)
			}
			return []byte("replacement"), nil
		case "virtual":
			if !errors.Is(origErr, fs.ErrNotExist) || len(content) != 0 {
				t.Errorf("missing file hook received (%q, %v)", content, origErr)
			}
			return []byte("generated"), nil
		default:
			return nil, denied
		}
	}))
	if err := filesystem.WriteFile("existing", []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	requireMemFile(t, filesystem, "existing", "replacement")
	requireMemFile(t, filesystem, "virtual", "generated")
	if _, err := filesystem.Open("denied"); !errors.Is(err, denied) {
		t.Fatalf("hook error: %v", err)
	}
	root, err := filesystem.Open(".")
	if err != nil {
		t.Fatalf("directory should bypass hook: %v", err)
	}
	root.Close()
	if !reflect.DeepEqual(calls, []string{"existing", "virtual", "denied"}) {
		t.Fatalf("hook calls = %v", calls)
	}
}
