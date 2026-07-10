package playwright

import (
	"fmt"

	"github.com/mxschmitt/playwright-go"
)

func init() {
	// Surface installation errors instead of silently continuing, otherwise a
	// failed driver install only shows up later as a confusing "please install
	// the driver first" error when the tests try to start.
	if err := playwright.Install(); err != nil {
		panic(fmt.Sprintf("could not install playwright driver: %v", err))
	}
}
