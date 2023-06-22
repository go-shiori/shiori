package core

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-shiori/shiori/internal/model"
)

func EbookGenerate(req ProcessRequest) (isFatalErr bool, err error) {
	// variable for store generated html code
	var html string

	book := req.Bookmark

	// Make sure bookmark ID is defined
	if book.ID == 0 {
		return true, fmt.Errorf("bookmark ID is not valid")
	}

	strID := strconv.Itoa(book.ID)
	ebookDir := fp.Join(req.DataDir, "ebook")

	// check if directory not exsist create that
	if _, err := os.Stat(ebookDir); os.IsNotExist(err) {
		os.MkdirAll(ebookDir, model.DataDirPerm)
	}
	ebookPath := fp.Join(req.DataDir, "ebook", strID+".epub")

	// TODO: can cheak with bookmark.hasEbook
	// if epub exist finish prosess else continue
	if _, err := os.Stat(ebookPath); err == nil {
		return false, nil
	}

	// create epub file
	epubFile, err := os.Create(ebookPath)
	if err != nil {
		return true, fmt.Errorf("can't create ebook")
	}
	defer epubFile.Close()

	// Create zip archive
	epubWriter := zip.NewWriter(epubFile)
	defer epubWriter.Close()

	// Create the mimetype file
	mimetypeWriter, err := epubWriter.Create("mimetype")
	if err != nil {
		return true, fmt.Errorf("can't create mimetype")
	}
	mimetypeWriter.Write([]byte("application/epub+zip"))

	// Create the container.xml file
	containerWriter, err := epubWriter.Create("META-INF/container.xml")
	if err != nil {
		return true, fmt.Errorf("can't create container.xml")
	}

	containerWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
	<rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	contentOpfWriter, err := epubWriter.Create("OEBPS/content.opf")
	if err != nil {
		return true, fmt.Errorf("can't create content.opf")
	}
	contentOpfWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="BookId">
  <metadata>
    <dc:title>` + book.Title + `</dc:title>
    <dc:language>en</dc:language>
    <dc:identifier id="BookId">urn:uuid:12345678-1234-5678-1234-567812345678</dc:identifier>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="content" href="content.html" media-type="application/xhtml+xml"/>
  </manifest>
  <spine toc="ncx">
    <itemref idref="content"/>
  </spine>
</package>`))

	// Create the toc.ncx file
	tocNcxWriter, err := epubWriter.Create("OEBPS/toc.ncx")
	if err != nil {
		return true, fmt.Errorf("can't create toc.ncx")
	}
	tocNcxWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN"
  "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="urn:uuid:12345678-1234-5678-1234-567812345678"/>
    <meta name="dtb:depth" content="1"/>
    <meta name="dtb:totalPageCount" content="0"/>
    <meta name="dtb:maxPageNumber" content="0"/>
  </head>
  <docTitle>
    <text>` + book.Title + `</text>
  </docTitle>
  <navMap>
    <navPoint id="navPoint-1" playOrder="1">
      <navLabel>
        <text >` + book.Title + `</text>
      </navLabel>
      <content src="content.html"/>
    </navPoint>
  </navMap>
</ncx>`))

	containerWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	// get list of images tag in html
	// TODO: if image present in html twice it will download twice too.
	imageList, _ := getImages(book.HTML)
	imgRegex := regexp.MustCompile(`<img.*?src="([^"]*)".*?>`)

	// Download image in html file and generate new html
	html = book.HTML
	for _, match := range imgRegex.FindAllStringSubmatch(book.HTML, -1) {
		imageURL := match[1]
		if _, ok := imageList[imageURL]; ok {
			// Download the image
			resp, err := http.Get(imageURL)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			// Get the image data
			imageData, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return true, fmt.Errorf("can't get image from the internet")
			}

			fileName := fp.Base(imageURL)
			filePath := "images/" + fileName
			imageWriter, err := epubWriter.Create(filePath)
			if err != nil {
				log.Fatal(err)
			}

			// Write the image to the file
			_, err = imageWriter.Write(imageData)
			if err != nil {
				return true, fmt.Errorf("can't create image file")
			}
			// Replace the image tag with the new downloaded image
			html = strings.ReplaceAll(html, match[0], fmt.Sprintf(`<img src="../%s"/>`, filePath))
		}
	}
	// Create the content.html file
	contentHtmlWriter, err := epubWriter.Create("OEBPS/content.html")
	if err != nil {
		return true, fmt.Errorf("can't create content.xml")
	}
	contentHtmlWriter.Write([]byte("<?xml version='1.0' encoding='utf-8'?>\n<html xmlns=\"http://www.w3.org/1999/xhtml\"><head><title>" + book.Title + "</title></head><body><h1 dir=\"auto\">" + book.Title + "</h1>" + "<content dir=\"auto\">" + html + "</content>" + "</body></html>"))
	return false, nil
}

// function get html and return list of image url inside html file
func getImages(html string) (map[string]string, error) {
	// Regular expression to match image tags and their URLs
	imageTagRegex := regexp.MustCompile(`<img.*?src="(.*?)".*?>`)

	// Find all matches in the HTML string
	imageTagMatches := imageTagRegex.FindAllStringSubmatch(html, -1)
	// Create a dictionary to store the image URLs
	images := make(map[string]string)

	// Check if there are any matches
	if len(imageTagMatches) == 0 {
		return nil, nil
	}

	// Loop through all the matches and add them to the dictionary
	for _, match := range imageTagMatches {
		imageURL := match[1]
		images[imageURL] = match[0]
	}

	return images, nil
}