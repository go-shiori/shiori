package main

import (
	"github.com/go-shiori/shiori/internal/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	shioriCmd := cmd.ShioriCmd()
	if err := shioriCmd.Execute(); err != nil {
		logrus.Fatalln(err)
	}
}
