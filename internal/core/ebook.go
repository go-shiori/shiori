package core

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-shiori/shiori/internal/model"
	"github.com/pkg/errors"
)

func EbookGenerate(req ProcessRequest) (book model.Bookmark, isFatalErr bool, err error) {
	// variable for store generated html code
	var html string

	book = req.Bookmark

	// Make sure bookmark ID is defined
	if book.ID == 0 {
		return book, true, fmt.Errorf("bookmark ID is not valid")
	}

	// cheak archive and thumb
	strID := strconv.Itoa(book.ID)

	imagePath := fp.Join(req.DataDir, "thumb", fmt.Sprintf("%d", book.ID))
	archivePath := fp.Join(req.DataDir, "archive", fmt.Sprintf("%d", book.ID))

	if _, err := os.Stat(imagePath); err == nil {
		book.ImageURL = fp.Join("/", "bookmark", strID, "thumb")
	}

	if _, err := os.Stat(archivePath); err == nil {
		book.HasArchive = true
	}
	ebookPath := fp.Join(req.DataDir, "ebook", fmt.Sprintf("%d.epub", book.ID))
	// if epub exist finish prosess else continue
	if _, err := os.Stat(ebookPath); err == nil {
		book.HasEbook = true
		return book, false, nil
	}
	contentType := req.ContentType
	if strings.Contains(contentType, "application/pdf") {
		return book, true, errors.Wrap(err, "can't create ebook for pdf")
	}

	ebookDir := fp.Join(req.DataDir, "ebook")
	// check if directory not exsist create that
	if _, err := os.Stat(ebookDir); os.IsNotExist(err) {
		err := os.MkdirAll(ebookDir, model.DataDirPerm)
		if err != nil {
			return book, true, errors.Wrap(err, "can't create ebook directory")
		}
	}
	// create epub file
	epubFile, err := os.Create(ebookPath)
	if err != nil {
		return book, true, errors.Wrap(err, "can't create ebook")
	}
	defer epubFile.Close()

	// Create zip archive
	epubWriter := zip.NewWriter(epubFile)
	defer epubWriter.Close()

	// Create the mimetype file
	mimetypeWriter, err := epubWriter.Create("mimetype")
	if err != nil {
		return book, true, errors.Wrap(err, "can't create mimetype")
	}
	_, err = mimetypeWriter.Write([]byte("application/epub+zip"))
	if err != nil {
		return book, true, errors.Wrap(err, "can't write into mimetype file")
	}

	// Create the container.xml file
	containerWriter, err := epubWriter.Create("META-INF/container.xml")
	if err != nil {
		return book, true, errors.Wrap(err, "can't create container.xml")
	}

	_, err = containerWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
	<rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))
	if err != nil {
		return book, true, errors.Wrap(err, "can't write into container.xml file")
	}

	contentOpfWriter, err := epubWriter.Create("OEBPS/content.opf")
	if err != nil {
		return book, true, errors.Wrap(err, "can't create content.opf")
	}
	_, err = contentOpfWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="BookId">
  <metadata>
    <dc:title>` + book.Title + `</dc:title>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
    <item id="content" href="content.html" media-type="application/xhtml+xml"/>
	<item id="id" href="../style.css" media-type="text/css"/>
  </manifest>
  <spine toc="ncx">
    <itemref idref="content"/>
  </spine>
</package>`))
	if err != nil {
		return book, true, errors.Wrap(err, "can't write into container.opf file")
	}

	// Create the style.css file
	styleWriter, err := epubWriter.Create("style.css")
	if err != nil {
		return book, true, errors.Wrap(err, "can't create content.xml")
	}
	_, err = styleWriter.Write([]byte(`body {
    text-indent: 0;
    margin: 0;
    text-align: justify;
}

head, form, script {
    display: none;
}

/* EPUB container of each individual HTML file */
DocFragment {
    page-break-before: always;
}

/* Headings */
h1, h2, h3, h4, h5, h6 {
    margin-top: 0.7em;
    margin-bottom: 0.5em;
    font-weight: bold;
    text-align: center;
    text-indent: 0;
    hyphenate: none;
    adobe-hyphenate: none;
}
h1, h2, h3 {
    page-break-before: always;
    page-break-inside: avoid;
    page-break-after: avoid;
}
h4, h5, h6 {
    page-break-inside: avoid;
    page-break-after: avoid;
}
h1 { font-size: 150%; }
h2 { font-size: 140%; }
h3 { font-size: 130%; }
h4 { font-size: 120%; }
h5 { font-size: 110%; }

/* Block elements */
div {
    margin: 1px;
}
p {
    text-align: justify;
    text-indent: 1.2em;
    margin-top: 0;
    margin-bottom: 0;
}
hr {
    height: 2px;
    background-color: #808080;
    margin-top: 0.5em;
    margin-bottom: 0.5em;
}

/* Lists */
ul {
    list-style-type: disc;
    margin-left: 1em;
}
ol {
    list-style-type: decimal;
    margin-left: 1em;
}
li {
    display: list-item;
    text-indent: 0;
}

/* Definitions */
dl {
    margin-left: 0;
}
dt {
    margin-left: 0;
    margin-top: 0.3em;
    font-weight: bold;
}
dd {
    margin-left: 1.3em;
}

/* Tables */
table {
    font-size: 80%;
    margin: 3px;
}
td, th {
    text-indent: 0;
    padding: 3px;
}
th {
    font-weight: bold;
    text-align: center;
    background-color: #DDD;
}
table caption {
    text-indent: 0;
    padding: 4px;
    background-color: #EEE;
}

/* Monospace block and inline elements */
pre {
    white-space: pre;
    font-family: "Droid Sans Mono", "Liberation Mono", "DeJaVu Sans Mono", monospace;
    text-align: left;
    margin-top: 0.5em;
    margin-bottom: 0.5em;
    /* background-color: #BFBFBF; */
}
code {
    white-space: pre;
    font-family: "Droid Sans Mono", "Liberation Mono", "DeJaVu Sans Mono", monospace;
}

/* Inline elements (all unknown elements default now to display: inline) */
sup                     { font-size: 70%; vertical-align: super; }
sub                     { font-size: 70%; vertical-align: sub; }
small                   { font-size: 80%; }
big                     { font-size: 130%; }
b, strong               { font-weight: bold; }
i, em, dfn, var, cite   { font-style: italic; }
u                       { text-decoration: underline; }
del, s, strike          { text-decoration: line-through; }
a                       { text-decoration: underline; color: gray; }

nobr {
    display: inline;
    hyphenate: none;
    white-space: nowrap;
}



/* Old element or className based selectors involving display: that
 * we need to support for older gDOMVersionRequested
 * DO NOT MODIFY BELOW to not break past highlights */

/* Images are now inline by default, and no more block with exceptions.
 * Note that when 'block', lvrend.cpp displays the title="" content
 * under the image */
img {
    -cr-ignore-if-dom-version-greater-or-equal: 20180528;
    text-align: center;
    text-indent: 0;
  	margin: auto;
    display: block;
}
sup img { -cr-ignore-if-dom-version-greater-or-equal: 20180528; display: inline; }
sub img { -cr-ignore-if-dom-version-greater-or-equal: 20180528; display: inline; }
a img   { -cr-ignore-if-dom-version-greater-or-equal: 20180528; display: inline; }
p img   { -cr-ignore-if-dom-version-greater-or-equal: 20180528; display: inline; }
p image { -cr-ignore-if-dom-version-greater-or-equal: 20180528; display: inline; } /* non html */

/* With dom_version < 20180528, unknown elements defaulted to 'display: inherit'
 * These ones here were explicitely set to inline (and some others not
 * specified here were also set to inline in fb2def.h */
b, strong, i, em, dfn, var, q, u, del, s, strike, small, big, sub, sup, acronym, tt, sa mp, kbd, code {
    -cr-ignore-if-dom-version-greater-or-equal: 20180528;
    display: inline;
}

.title, .title1, .title2, .title3, .title4, .title5, .subtitle {
    -cr-ignore-if-dom-version-greater-or-equal: 20180528;
    display: block;
}
.fb2_info { -cr-ignore-if-dom-version-greater-or-equal: 20180528; display: block; }
.code     { -cr-ignore-if-dom-version-greater-or-equal: 20180528; display: block; }

`))
	if err != nil {
		return book, true, errors.Wrap(err, "can't write into style.css file")
	}
	// Create the toc.ncx file
	tocNcxWriter, err := epubWriter.Create("OEBPS/toc.ncx")
	if err != nil {
		return book, true, fmt.Errorf("can't create toc.ncx")
	}
	_, err = tocNcxWriter.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
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
	if err != nil {
		return book, true, fmt.Errorf("can't write into toc.ncx file")
	}

	// get list of images tag in html
	imageList, _ := getImages(book.HTML)
	imgRegex := regexp.MustCompile(`<img.*?src="([^"]*)".*?>`)

	// Create a set to store unique image URLs
	imageSet := make(map[string]bool)

	// Download image in html file and generate new html
	html = book.HTML
	for _, match := range imgRegex.FindAllStringSubmatch(book.HTML, -1) {
		imageURL := match[1]
		if _, ok := imageList[imageURL]; ok && !imageSet[imageURL] {
			// Add the image URL to the set
			imageSet[imageURL] = true

			// Download the image
			resp, err := http.Get(imageURL)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			// Get the image data
			imageData, err := io.ReadAll(resp.Body)
			if err != nil {
				return book, true, fmt.Errorf("can't get image from the internet")
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
				return book, true, fmt.Errorf("can't create image file")
			}
			// Replace the image tag with the new downloaded image
			html = strings.ReplaceAll(html, match[0], fmt.Sprintf(`<img src="../%s"/>`, filePath))
		}
	}
	// Create the content.html file
	contentHtmlWriter, err := epubWriter.Create("OEBPS/content.html")
	if err != nil {
		return book, true, fmt.Errorf("can't create content.xml")
	}
	_, err = contentHtmlWriter.Write([]byte("<?xml version='1.0' encoding='utf-8'?>\n<html xmlns=\"http://www.w3.org/1999/xhtml\">\n<head>\n\t<title>" + book.Title + "</title>\n\t<link href=\"../style.css\" rel=\"stylesheet\" type=\"text/css\"/>\n</head>\n<body>\n\t<h1 dir=\"auto\">" + book.Title + "</h1>" + "\n<content dir=\"auto\">\n" + html + "\n</content>" + "\n</body></html>"))
	if err != nil {
		return book, true, fmt.Errorf("can't write into content.html")
	}
	book.HasEbook = true
	return book, false, nil
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
