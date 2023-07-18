package views

import "embed"

//go:embed assets/*
var Assets embed.FS

//go:embed *.html
var Templates embed.FS
