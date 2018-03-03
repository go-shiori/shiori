package cmd

import (
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/RadhiFadlillah/shiori/model"
	"github.com/spf13/cobra"
)

var (
	exportCmd = &cobra.Command{
		Use:   "export target-file",
		Short: "Export bookmarks into HTML file in Netscape Bookmark format",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := exportBookmarks(args[0])
			if err != nil {
				cError.Println(err)
				return
			}

			fmt.Println("Export finished")
		},
	}
)

func init() {
	rootCmd.AddCommand(exportCmd)
}

func exportBookmarks(dstPath string) error {
	// Read bookmarks from database
	bookmarks, err := DB.GetBookmarks(false)
	if err != nil {
		return err
	}

	if len(bookmarks) == 0 {
		return fmt.Errorf("No bookmarks saved yet")
	}

	// Open destination file
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Create template
	funcMap := template.FuncMap{
		"unix": func(str string) int64 {
			t, err := time.Parse("2006-01-02 15:04:05", str)
			if err != nil {
				return time.Now().Unix()
			}

			return t.Unix()
		},
		"combine": func(tags []model.Tag) string {
			strTags := make([]string, len(tags))
			for i, tag := range tags {
				strTags[i] = tag.Name
			}

			return strings.Join(strTags, ",")
		},
	}

	tplFile := `<!DOCTYPE NETSCAPE-Bookmark-file-1>` +
		`<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">` +
		`<TITLE>Bookmarks</TITLE>` +
		`<H1>Bookmarks</H1>` +
		`<DL><p>` +
		`{{range $book := .}}` +
		`<DT><A HREF="{{$book.URL}}" ADD_DATE="{{unix $book.Modified}}" TAGS="{{combine $book.Tags}}">{{$book.Title}}</A>` +
		`{{if gt (len $book.Excerpt) 0}}<DD>{{$book.Excerpt}}{{end}}{{end}}` +
		`</DL><p>`

	tpl, err := template.New("export").Funcs(funcMap).Parse(tplFile)
	if err != nil {
		return err
	}

	return tpl.Execute(dstFile, &bookmarks)
}
