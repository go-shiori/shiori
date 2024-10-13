package core

import (
	"fmt"

	"github.com/go-shiori/shiori/internal/model"
)

var userAgent = fmt.Sprintf("Shiori/%s (+https://github.com/go-shiori/shiori)", model.BuildVersion)
