//go:build js
package system

var _ WriteableFS = (*FS)(nil)

func MakeFS() WriteableFS {
    return NewMemFS()
}
