package engine

import (
	_ "embed"
	"strings"
)

//go:embed key/key.txt
var embeddedAuthKey string

func clientAuthKey() string { return strings.TrimSpace(embeddedAuthKey) }
