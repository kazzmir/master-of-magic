//go:build js
package system

type JsFS struct {
    *FS
}

func (fs *JsFS) MaybeDownload(path string) {
}

var _ WriteableFS = (*JsFS)(nil)

func MakeFS() WriteableFS {
    return &JsFS{
        FS: NewMemFS(),
    }
}
