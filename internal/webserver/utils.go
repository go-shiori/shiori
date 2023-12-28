package webserver

import (
	"html/template"
	"io"
	"net"
	"os"
	"syscall"
)

func FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}

func createTemplate(filename string, funcMap template.FuncMap) (*template.Template, error) {
	// Open file
	src, err := assets.Open(filename)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Read file content
	srcContent, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	// Create template
	return template.New(filename).Delims("$$", "$$").Funcs(funcMap).Parse(string(srcContent))
}

func checkError(err error) {
	if err == nil {
		return
	}

	// Check for a broken connection, as it is not really a
	// condition that warrants a panic stack trace.
	if ne, ok := err.(*net.OpError); ok {
		if se, ok := ne.Err.(*os.SyscallError); ok {
			if se.Err == syscall.EPIPE || se.Err == syscall.ECONNRESET {
				return
			}
		}
	}

	panic(err)
}
