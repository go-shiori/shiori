# Shiori

[![Travis CI](https://travis-ci.org/RadhiFadlillah/shiori.svg?branch=master)](https://travis-ci.org/RadhiFadlillah/shiori)
[![Go Report Card](https://goreportcard.com/badge/github.com/radhifadlillah/shiori)](https://goreportcard.com/report/github.com/radhifadlillah/shiori)
[![Docker Build Status](https://img.shields.io/docker/build/radhifadlillah/shiori.svg)](https://hub.docker.com/r/radhifadlillah/shiori/)

Shiori is a simple bookmarks manager written in Go language. Intended as a simple clone of [Pocket](https://getpocket.com//). You can use it as command line application or as web application. This application is distributed as a single binary, which means it can be installed and used easily.

![Screenshot](https://raw.githubusercontent.com/RadhiFadlillah/shiori/master/screenshot/pc-grid.png)

## Features

- Simple and clean command line interface.
- Basic bookmarks management i.e. add, edit and delete.
- Search bookmarks by their title, tags, url and page content.
- Import and export bookmarks from and to Netscape Bookmark file.
- Portable, thanks to its single binary format and sqlite3 database
- Simple web interface for those who don't want to use a command line app.
- Where possible, by default `shiori` will download a static copy of the webpage in simple text and HTML format, which later can be used as an offline archive for that page.

## Documentation

All documentation is available in [wiki](https://github.com/RadhiFadlillah/shiori/wiki). If you think there are incomplete or incorrect information, feels free to edit it.

## License

Shiori is distributed using [MIT license](https://choosealicense.com/licenses/mit/), which means you can use and modify it however you want. However, if you make an enhancement for it, if possible, please send a pull request.
