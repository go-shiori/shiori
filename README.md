# Shiori

[![IC](https://github.com/go-shiori/shiori/actions/workflows/push.yml/badge.svg?branch=master)](https://github.com/go-shiori/shiori/actions/workflows/push.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-shiori/shiori)](https://goreportcard.com/report/github.com/go-shiori/shiori)
[![#shiori-general:matrix.org](https://img.shields.io/badge/matrix-%23shiori-orange)](https://matrix.to/#/#shiori:matrix.org)
[![Containers](https://img.shields.io/static/v1?label=Container&message=Images&color=1488C6&logo=docker)](https://github.com/go-shiori/shiori/pkgs/container/shiori)

**Check out our latest [Announcements](https://github.com/go-shiori/shiori/discussions/categories/announcements)**

Shiori is a simple bookmarks manager written in the Go language. Intended as a simple clone of [Pocket][pocket]. You can use it as a command line application or as a web application. This application is distributed as a single binary, which means it can be installed and used easily.

![Screenshot][screenshot]

## Features

- Basic bookmarks management i.e. add, edit, delete and search.
- Import and export bookmarks from and to Netscape Bookmark file.
- Import bookmarks from Pocket.
- Simple and clean command line interface.
- Simple and pretty web interface for those who don't want to use a command line app.
- Portable, thanks to its single binary format.
- Support for sqlite3, PostgreSQL and MySQL as its database.
- Where possible, by default `shiori` will parse the readable content and create an offline archive of the webpage.
- [BETA] [web extension][web-extension] support for Firefox and Chrome.

![Comparison of reader mode and archive mode][mode-comparison]

## Documentation

All documentation is available in the [docs folder][documentation]. If you think there is incomplete or incorrect information, feel free to edit it by submitting a pull request.

## License

Shiori is distributed under the terms of the [MIT license][mit], which means you can use it and modify it however you want. However, if you make an enhancement for it, if possible, please send a pull request.

[documentation]: https://github.com/go-shiori/shiori/blob/master/docs/index.md
[mit]: https://choosealicense.com/licenses/mit/
[web-extension]: https://github.com/go-shiori/shiori-web-ext
[screenshot]: https://raw.githubusercontent.com/go-shiori/shiori/master/docs/readme/cover.png
[mode-comparison]: https://raw.githubusercontent.com/go-shiori/shiori/master/docs/readme/comparison.png
[pocket]: https://getpocket.com/
[256]: https://github.com/go-shiori/shiori/issues/256
