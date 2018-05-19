package cmd

import (
	"errors"
	"io"
	"net/http"
	nurl "net/url"
	"os"
	"os/exec"
	fp "path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

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

	errInvalidIndex = errors.New("Index is not valid")
)

func normalizeSpace(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

func clearUTMParams(url *nurl.URL) {
	newQuery := nurl.Values{}
	for key, value := range url.Query() {
		if !strings.HasPrefix(key, "utm_") {
			newQuery[key] = value
		}
	}

	url.RawQuery = newQuery.Encode()
}

func downloadFile(url, dstPath string, timeout time.Duration) error {
	// Fetch data from URL
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Make sure destination directory exist
	err = os.MkdirAll(fp.Dir(dstPath), os.ModePerm)
	if err != nil {
		return err
	}

	// Create destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Write response body to the file
	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}

	return nil
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

// parseIndexList converts a list of indices to their integer values
func parseIndexList(indices []string) ([]int, error) {
	var listIndex []int
	for _, strIndex := range indices {
		if !strings.Contains(strIndex, "-") {
			index, err := strconv.Atoi(strIndex)
			if err != nil || index < 1 {
				return nil, errInvalidIndex
			}

			listIndex = append(listIndex, index)
			continue
		}

		parts := strings.Split(strIndex, "-")
		if len(parts) != 2 {
			return nil, errInvalidIndex
		}

		minIndex, errMin := strconv.Atoi(parts[0])
		maxIndex, errMax := strconv.Atoi(parts[1])
		if errMin != nil || errMax != nil || minIndex < 1 || minIndex > maxIndex {
			return nil, errInvalidIndex
		}

		for i := minIndex; i <= maxIndex; i++ {
			listIndex = append(listIndex, i)
		}
	}
	return listIndex, nil
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
