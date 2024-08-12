package common

import (
    "bytes"
    _ "embed"
    "github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed futura.ttf
var FuturaTTF []byte

func LoadFont() (*text.GoTextFaceSource, error) {
    return text.NewGoTextFaceSource(bytes.NewReader(FuturaTTF))
}
