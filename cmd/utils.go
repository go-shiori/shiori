package cmd

import (
	nurl "net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	cIndex    = color.New(color.FgHiCyan)
	cSymbol   = color.New(color.FgHiMagenta)
	cTitle    = color.New(color.FgHiGreen).Add(color.Bold)
	cReadTime = color.New(color.FgHiMagenta)
	cURL      = color.New(color.FgHiYellow)
	cError    = color.New(color.FgHiRed)
	cExcerpt  = color.New(color.FgHiWhite)
	cTag      = color.New(color.FgHiBlue)
)

func normalizeSpace(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

func clearUTMParams(url *nurl.URL) string {
	newQuery := nurl.Values{}
	for key, value := range url.Query() {
		if strings.HasPrefix(key, "utm_") {
			continue
		}

		newQuery[key] = value
	}

	url.RawQuery = newQuery.Encode()
	return url.String()
}

// openBrowser tries to open the URL in a browser,
// and returns whether it succeed in doing so.
func openBrowser(url string) error {
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}

	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Run()
}

func getTerminalWidth() int {
	width, _, _ := terminal.GetSize(int(os.Stdin.Fd()))
	return width
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
