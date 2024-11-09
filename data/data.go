package data

import (
    "embed"
)

//go:embed data/*
var Data embed.FS
