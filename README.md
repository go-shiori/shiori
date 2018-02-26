# Shiori

Shiori is a simple bookmarks manager written in Go language. You can use it as command line application (like [`buku`](https://github.com/jarun/Buku)) or as web application (like [Pocket](https://getpocket.com//) or [Pinterest](https://www.pinterest.com/)). This application is distributed as a single binary, which make it can be installed and used easily.

![Screenshot](https://raw.githubusercontent.com/RadhiFadlillah/shiori/master/screenshot.png)

## Table of Contents

- Features
- Installation
- Usage
- License

## Features

- Simple and clean command.
- Basic bookmarks management i.e. add, edit and delete.
- Search bookmarks by its titles, tags, url and page content.
- Import and export bookmarks from and to Netscape Bookmark file.
- Portable, thanks to its single binary format and sqlite3 database.
- Simple web interface for those who doesn't used to command line app.
- Where possible, by default `shiori` will download a static copy of the webpage in simple text and HTML format, which later can be used as offline archive for that page.

## Installation

You can download the latest version of `shiori` from the release page, then put it in your `PATH`. If you want to build from source, make sure `go` is installed, then run :

```
go get github.com/RadhiFadlillah/go-readability
```

## Usage

```
Simple command-line bookmark manager built with Go.

Usage:
  shiori [command]

Available Commands:
  account     Manage account for accessing web interface.
  add         Bookmark the specified URL.
  delete      Delete the saved bookmarks.
  export      Export bookmarks into HTML file in Netscape Bookmark format.
  help        Help about any command
  import      Import bookmarks from HTML file in Netscape Bookmark format.
  open        Open the saved bookmarks.
  print       Print the saved bookmarks.
  search      Search bookmarks by submitted keyword.
  serve       Serve web app for managing bookmarks.
  update      Update the saved bookmarks.

Flags:
  -h, --help   help for shiori

Use "shiori [command] --help" for more information about a command.
```

## License

Shiori is distributed using [MIT license](https://choosealicense.com/licenses/mit/), which means you can use and modify it however you want. However, if you make an enchancement for shiori, if possible, please send the pull request.
