package readability

import (
	"crypto/md5"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

func createDocFromFile(path string) (*goquery.Document, error) {
	// Open file
	src, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Create document
	return goquery.NewDocumentFromReader(src)
}

func hashNode(node *goquery.Selection) string {
	if node == nil {
		return ""
	}

	html, _ := node.Html()
	return fmt.Sprintf("%x", md5.Sum([]byte(html)))
}

func strLen(str string) int {
	return utf8.RuneCountInString(str)
}

func findSeparator(str string, separators ...string) (int, string) {
	words := strings.Fields(str)
	for i, word := range words {
		for _, separator := range separators {
			if word == separator {
				return i, separator
			}
		}
	}

	return -1, ""
}

func hasSeparator(str string, separators ...string) bool {
	idx, _ := findSeparator(str, separators...)
	return idx != -1
}

func removeSeparator(str string, separators ...string) string {
	words := strings.Fields(str)
	finalWords := []string{}

	for _, word := range words {
		for _, separator := range separators {
			if word != separator {
				finalWords = append(finalWords, word)
			}
		}
	}

	return strings.Join(finalWords, " ")
}

func normalizeText(str string) string {
	return strings.Join(strings.Fields(str), " ")
}
