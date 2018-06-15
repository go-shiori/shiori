# Shiori

[![Travis CI](https://travis-ci.org/RadhiFadlillah/shiori.svg?branch=master)](https://travis-ci.org/RadhiFadlillah/shiori)
[![Go Report Card](https://goreportcard.com/badge/github.com/radhifadlillah/shiori)](https://goreportcard.com/report/github.com/radhifadlillah/shiori)
[![Docker Build Status](https://img.shields.io/docker/build/radhifadlillah/shiori.svg)](https://hub.docker.com/r/radhifadlillah/shiori/)

Shiori is a simple bookmarks manager written in Go language. Intended as a simple clone of [Pocket](https://getpocket.com//). You can use it as command line application or as web application. This application is distributed as a single binary, which means it can be installed and used easily.

![Screenshot](https://raw.githubusercontent.com/RadhiFadlillah/shiori/master/screenshot/pc-grid.png)

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

### Linux

You can download the latest version of `shiori` from [the release page](https://github.com/RadhiFadlillah/shiori/releases/latest), then put it in your `PATH`. If you want to build from source, make sure `go` is installed, then run :

```
go get github.com/RadhiFadlillah/shiori
```

### Windows

In order to build this project, you will need to have access to a Linux machine or a Windows Subsystem for Linux environment. The following commands work under Ubuntu 16.04 LTS:

```bash
go get -d github.com/RadhiFadlillah/shiori
sudo apt install mingw-w64
cd $GOPATH/src/github.com/RadhiFadlillah/shiori
GOOS=windows CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags "-extldflags -static"

```

Once this is done, copy `shiori.exe` to your Windows machine or environment. Happy bookmarking!

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

## Advanced

By default, `shiori` will create database in `$HOME/.shiori.db`. If you want to set the database to another location, you can set the environment variable `ENV_SHIORI_DB` to your desired path :

- If `ENV_SHIORI_DB` points to a directory, it will create `.shiori.db` file inside that directory, so the final path for database is `$ENV_SHIORI_DB/.shiori.db`.
- Else, it will create a new database file in the specified path.

## Usage with Docker

There's a Dockerfile that enables you to build your own dockerized Shiori :

```bash
docker build -t shiori .
```

You can also pull the latest automated build from Docker Hub :

```bash
docker pull radhifadlillah/shiori
```

### Run the container

After building the image you will be able to start a container from it. To
preserve the database you need to bind the file. In this example we're locating
the `shiori.db` file in our CWD.

```sh
touch shiori.db
docker run --rm --name shiori -p 8080:8080 -v $(pwd)/shiori.db:/srv/shiori.db radhifadlillah/shiori
```

If you want to run the container in the background add `-d` after `run`.

### Console access for container

```sh
# First open a console to the container (as you will need to enter your password)
# and the default tty does not support hidden inputs
docker exec -it shiori sh
```

### Initialize shiori with password

As after running the container there will be no accounts created, you need to
open a console within your container and run the following command:

```sh
shiori account add <your-desired-username>
Password: <enter-your-password>
```

And you're now ready to go and access shiori via web.

### Run Shiori docker container as systemd image

1. Create a service unit for `systemd` at `/etc/systemd/system/shiori.service`.

    ```ini
    [Unit]
    Description=Shiori container
    After=docker.service

    [Service]
    Restart=always
    ExecStartPre=-/usr/bin/docker rm shiori-1
    ExecStart=/usr/bin/docker run \
      --rm \
      --name shiori-1 \
      -p 8080:8080 \
      -v /srv/machines/shiori/shiori.db:/srv/shiori/shiori.db \
      radhifadlillah/shiori
    ExecStop=/usr/bin/docker stop -t 2 shiori-1

    [Install]
    WantedBy=multi-user.target
    ```

2. Set up data directory

    This assumes, that the Shiori container has a runtime directory to store their
    database, which is at `/srv/machines/shiori`. If you want to modify that,
    make sure, to fix your `shiori.service` as well.

    ```sh
    install -d /srv/machines/shiori
    touch /srv/machines/shiori/shiori.db
    ```

3. Enable and start the container

    ```sh
    systemctl enable --now shiori
    ```

## Examples

*Hint:* If you want to practice the following commands with the docker container,
[run the image](#run-the-container) and [open a
console](#console-access-for-container). After that go along with the examples.

1. Save new bookmark with tags "nature" and "climate-change".

   ```sh
   shiori add https://grist.org/article/let-it-go-the-arctic-will-never-be-frozen-again/ -t nature,climate-change
   ```

2. Print all saved bookmarks.

   ```sh
   shiori print
   ```

2. Print bookmarks with index 1 and 2.

   ```sh
   shiori print 1 2
   ```

3. Search bookmarks that contains "sqlite" in their title, excerpt, url or content.

   ```sh
   shiori search sqlite
   ```

4. Search bookmarks with tag "nature".

   ```sh
   shiori search -t nature
   ```

5. Delete all bookmarks.

   ```sh
   shiori delete
   ```

6. Delete all bookmarks with tag "nature".

   ```sh
   shiori delete $(shiori search -t nature -i)
   ```

7. Update all bookmarks' data and content.

   ```sh
   shiori update
   ```

8. Update bookmark in index 1.

   ```sh
   shiori update 1
   ```

9. Change title and excerpt from bookmark in index 1.

   ```sh
   shiori update 1 -i "New Title" -e "New excerpt"
   ```

10. Add tag "future" and remove tag "climate-change" from bookmark in index 1.

    ```sh
    shiori update 1 -t future,-climate-change
    ```

11. Import bookmarks from HTML Netscape Bookmark file.

    ```sh
    shiori import exported-from-firefox.html
    ```

12. Export saved bookmarks to HTML Netscape Bookmark file.

    ```sh
    shiori export target.html
    ```

13. Open all saved bookmarks in browser.

    ```sh
    shiori open
    ```

14. Open text cache of bookmark in index 1.

    ```sh
    shiori open 1 -c
    ```

15. Serve web app in port 9000.

    ```sh
    shiori serve -p 9000
    ```

16. Create new account for login to web app.

    ```sh
    shiori account add username
    ```

## License

Shiori is distributed using [MIT license](https://choosealicense.com/licenses/mit/), which means you can use and modify it however you want. However, if you make an enhancement for it, if possible, please send a pull request.
