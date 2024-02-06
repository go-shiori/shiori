package views

import "embed"

//go:embed assets/js/* assets/css/* assets/res/*
var Assets embed.FS

//go:embed *.html
var Templates embed.FS
