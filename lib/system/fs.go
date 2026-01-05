package system

import (
    "io"
    "io/fs"
)

type WriteableFS interface {
    fs.FS
    Create(name string) (io.WriteCloser, error)
    Remove(name string) error

    // for js, cause the file to be downloaded by the browser
    MaybeDownload(name string)
}
