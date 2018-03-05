# Shiori

Shiori is a simple bookmarks manager written in Go language. Intended as a simple clone of [Pocket](https://getpocket.com//). You can use it as command line application or as web application. This application is distributed as a single binary, which means it can be installed and used easily.

![Screenshot](https://raw.githubusercontent.com/RadhiFadlillah/shiori/master/screenshot.png)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Advanced](#advanced)
- [Examples](#examples)
- [License](#license)

## Features

- Simple and clean command line interface.
- Basic bookmarks management i.e. add, edit and delete.
- Search bookmarks by their title, tags, url and page content.
- Import and export bookmarks from and to Netscape Bookmark file.
- Portable, thanks to its single binary format and sqlite3 database
- Simple web interface for those who don't want to use a command line app.
- Where possible, by default `shiori` will download a static copy of the webpage in simple text and HTML format, which later can be used as an offline archive for that page.

## Installation

You can download the latest version of `shiori` from the release page, then put it in your `PATH`. If you want to build from source, make sure `go` is installed, then run :

```
go get github.com/RadhiFadlillah/shiori
```

## Usage

```
Simple command-line bookmark manager built with Go.

Usage:
  shiori [command]

Available Commands:
  account     Manage account for accessing web interface
  add         Bookmark the specified URL
  delete      Delete the saved bookmarks
  export      Export bookmarks into HTML file in Netscape Bookmark format
  help        Help about any command
  import      Import bookmarks from HTML file in Netscape Bookmark format
  open        Open the saved bookmarks
  print       Print the saved bookmarks
  search      Search bookmarks by submitted keyword
  serve       Serve web app for managing bookmarks
  update      Update the saved bookmarks

Flags:
  -h, --help   help for shiori

Use "shiori [command] --help" for more information about a command.
```

### Advanced

By default, `shiori` will create database in the location where you run it. For example, if you run `shiori`. To set the database to a specific location, you can set the environment variable `ENV_SHIORI_DB` to your desired path. 

## Examples

1. Save new bookmark with tags "nature" and "climate-change".

   ```
   shiori add https://grist.org/article/let-it-go-the-arctic-will-never-be-frozen-again/ -t nature,climate-change
   ```

2. Print all saved bookmarks.

   ```
   shiori print
   ```

2. Print bookmarks with index 1 and 2.

   ```
   shiori print 1 2
   ```

3. Search bookmarks that contains "sqlite" in their title, excerpt, url or content.

   ```
   shiori search sqlite
   ```

4. Search bookmarks with tag "nature".

   ```
   shiori search -t nature
   ```

5. Delete all bookmarks.

   ```
   shiori delete
   ```

6. Delete all bookmarks with tag "nature".

   ```
   shiori delete $(shiori search -t nature -i)
   ```

7. Update all bookmarks' data and content.

   ```
   shiori update
   ```

8. Update bookmark in index 1.

   ```
   shiori update 1
   ```

9. Change title and excerpt from bookmark in index 1.

   ```
   shiori update 1 -i "New Title" -e "New excerpt"
   ```

10. Add tag "future" and remove tag "climate-change" from bookmark in index 1.

    ```
    shiori update 1 -t future,-climate-change
    ```

11. Import bookmarks from HTML Netscape Bookmark file.

    ```
    shiori import exported-from-firefox.html
    ```

12. Export saved bookmarks to HTML Netscape Bookmark file.

    ```
    shiori export target.html
    ```

13. Open all saved bookmarks in browser.

    ```
    shiori open
    ```

14. Open text cache of bookmark in index 1.

    ```
    shiori open 1 -c
    ```

15. Serve web app in port 9000.

    ```
    shiori serve -p 9000
    ```

16. Create new account for login to web app.

    ```
    shiori account add username
    ```

## License

Shiori is distributed using [MIT license](https://choosealicense.com/licenses/mit/), which means you can use and modify it however you want. However, if you make an enhancement for it, if possible, please send a pull request.
