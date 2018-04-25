package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
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

func getTerminalWidth() int {
	width, _, _ := terminal.GetSize(int(os.Stdin.Fd()))
	return width
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func configDir() string {
	homeDir, err := homedir.Dir()
	checkError(err)
	configDir := homeDir + "/.shiori"
	return configDir
}

func configPath() string {
	configPath := configDir() + "/config.json"
	return configPath
}
