//go:build !js
package system

import (
    "io"
    "io/fs"
    "os"
)

type StandardFS struct {
}

func (standard *StandardFS) Create(name string) (io.WriteCloser, error) {
    return os.Create(name)
}

func (standard *StandardFS) Open(name string) (fs.File, error) {
    return os.Open(name)
}

func (standard *StandardFS) Remove(name string) error {
    return os.Remove(name)
}

var _ WriteableFS = (*StandardFS)(nil)

func MakeFS() WriteableFS {
    return &StandardFS{}
    // return NewMemFS()
}
