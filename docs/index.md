
# Shiori

Shiori is a simple bookmarks manager written in the Go language. You can use it as a command line application or as a web application, distributed as a single binary, which means it can be installed and used easily.

You can use it as a bookmark manager or as an alternative to [Pocket][pocket] or other read-later services due to it's archival and readability features.

![Screenshot][screenshot]

## Features

- Basic **bookmarks management** i.e. add, edit, delete and search.
- **Import and export** bookmarks from and to Netscape Bookmark file.
- Import bookmarks from **Pocket**.
- Simple and clean **command line interface**.
- Simple and pretty **web interface** for those who don't want to use a command line app.
- **Portable**, thanks to its single binary format.
- Support for **sqlite3**, **PostgreSQL** and **MySQL** as its database.
- Where possible and by default `shiori` will **parse the readable content**.
- Optionally **create an offline archive of the webpage** using [warc][warc]. (See [#353](https://github.com/go-shiori/shiori/issues/353))
- Optionally **create an ebook** from the readable content in ePub using [go-epub][go-epub].
- [BETA] [web extension][web-extension] support for Firefox and Chrome.

![Comparison of reader mode and archive mode][mode-comparison]

## License

Shiori is distributed under the terms of the [MIT license][mit], which means you can use it and modify it however you want. However, if you make an enhancement for it, if possible, please send a pull request.

[documentation]: https://github.com/go-shiori/shiori/blob/master/docs/index.md
[mit]: https://choosealicense.com/licenses/mit/
[web-extension]: https://github.com/go-shiori/shiori-web-ext
[screenshot]: https://raw.githubusercontent.com/go-shiori/shiori/master/docs/readme/cover.png
[mode-comparison]: https://raw.githubusercontent.com/go-shiori/shiori/master/docs/readme/comparison.png
[go-epub]: https://github.com/bmaupin/go-epub
[pocket]: https://getpocket.com/
[warc]: https://github.com/go-shiori/warc
[256]: https://github.com/go-shiori/shiori/issues/256


<!-- ## Getting started

- [Installation](./Installation.md)
- [Configuration](./Configuration.md)
- [Storage](./Storage.md)
- [Screenshots](./Screenshots.md)

## Usage

- [Command Line Interface](./CLI.md)
- [Web Interface](./Web-Intarface.md)
- [API (legacy)](./API.md) (**Deprecated!**, see [APIv1](./APIv1.md))
- [APIv1](./APIv1.md) ([What is this?](https://github.com/go-shiori/shiori/issues/640))

## Advanced

- [Contributing](./Contribute.md) -->
