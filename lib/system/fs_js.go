//go:build js
package system

import (
    "syscall/js"
    "bytes"
    "io"
)

type JsFS struct {
    *FS
}

func (fs *JsFS) MaybeDownload(path string) {
    // Example file contents
    file, err := fs.Open(path)
    if err != nil {
        return
    }

    var data bytes.Buffer
    io.Copy(&data, file)

    file.Close()

    // Create Uint8Array
    uint8Array := js.Global().Get("Uint8Array").New(data.Len())
    if uint8Array.IsUndefined() || uint8Array.IsNull() {
        return
    }
    js.CopyBytesToJS(uint8Array, data.Bytes())

    // Create Blob
    blob := js.Global().Get("Blob").New(
        []any{uint8Array},
        map[string]any{"type": "text/plain"},
    )

    // Create object URL
    url := js.Global().Get("URL").Call("createObjectURL", blob)

    // Create <a download>
    document := js.Global().Get("document")
    a := document.Call("createElement", "a")
    a.Set("href", url)
    a.Set("download", path)

    // Trigger download
    document.Get("body").Call("appendChild", a)
    a.Call("click")
    document.Get("body").Call("removeChild", a)

    // Cleanup
    js.Global().Get("URL").Call("revokeObjectURL", url)
}

var _ WriteableFS = (*JsFS)(nil)

func MakeFS() WriteableFS {
    return &JsFS{
        FS: NewMemFS(),
    }
}
