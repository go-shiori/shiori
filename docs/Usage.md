Before using `shiori`, make sure it has been installed on your system. By default, `shiori` will store its data in directory `$HOME/.local/share/shiori`. If you want to set the data directory to another location, you can set the environment variable `SHIORI_DIR` (`ENV_SHIORI_DIR` when you are before `1.5.0`) to your desired path.

<!-- TOC -->

- [Using Command Line Interface](#using-command-line-interface)
  - [Search syntax](#search-syntax)
- [How to use the Web Interface](#using-web-interface)
  - [Adding bookmarks manually](#adding-bookmarks-manually)
- [Improved import from Pocket](#improved-import-from-pocket)
- [Import from Wallabag](#import-from-wallabag)

<!-- /TOC -->

## Using Command Line Interface

Shiori is composed by several subcommands. To see the documentation, run `shiori -h` :

```
Simple command-line bookmark manager built with Go

Usage:
  shiori [command]

Available Commands:
  add         Bookmark the specified URL
  check       Find bookmarked sites that no longer exists on the internet
  delete      Delete the saved bookmarks
  export      Export bookmarks into HTML file in Netscape Bookmark format
  help        Help about any command
  import      Import bookmarks from HTML file in Netscape Bookmark format
  open        Open the saved bookmarks
  pocket      Import bookmarks from Pocket's exported HTML file
  print       Print the saved bookmarks
  serve       Serve web interface for managing bookmarks
  server      Run the Shiori webserver [alpha]
  update      Update the saved bookmarks
  version     Output the shiori version

Flags:
  -h, --help       help for shiori
      --portable   run shiori in portable mode

Use "shiori [command] --help" for more information about a command.
```

### Search syntax
With the `print` command line interface, you can use `-s` flag to submit keywords that will be searched either in url, title, excerpts or cached content.
You may also use `-t` flag to include tags and `-e` flag to exclude tags.


## How to use the Web Interface

To access web interface run `shiori serve` or start Docker container following tutorial above. If you want to use a different port instead of 8080, you can simply run `shiori serve -p <portnumber>`. Once started you can access the web interface in `http://localhost:8080` or `http://localhost:<portnumber>` if you customized it. You will be greeted with login screen like this :

![Login screen](https://raw.githubusercontent.com/go-shiori/shiori/master/docs/screenshots/01-login.png)

Since this is our first time, we don't have any account registered yet. With that said, we can use the default user to access web interface :

```
username: shiori
password: gopher
```

Once login succeed you will be able to use the web interface. To add the new account, open the settings page and add accounts as needed :

![Options page](https://raw.githubusercontent.com/go-shiori/shiori/master/docs/screenshots/04-options.png)

The first new account you add will become the owner and it will deactivate the "shiori:gopher" default user automatically.

When searching for bookmarks, you may use `tag:tagname` to include tags and `-tag:tagname` to exclude tags in the search bar. You can also use tags dialog to do this :

- `Click` on the tag name to include it;
- `Alt + Click` on the tag name to exclude it.


### Adding bookmarks manually

- Within the web application, you can add a new bookmark by following the steps depicted in the screenshots below:

![gs-01](./screenshots/17-gs-add-bookmark.png)
![gs-02](./screenshots/18-gs-add-bookmark.png)
![gs-03](./screenshots/19-gs-add-bookmark.png)

This action will generate a new entry, as illustrated in the final screenshot.

- Following this step, you should proceed to select the __Update Archive__ button in order to initiate the process of creating a local copy of the page.

![gs-04](./screenshots/20-gs-add-bookmark.png)
![gs-05](./screenshots/21-gs-add-bookmark.png)

- At this point, you will be able to observe your downloaded bookmark, which has been successfully stored locally.

![gs-06](./screenshots/22-gs-add-bookmark.png)


## Improved import from Pocket

Shiori offers a [Command Line Interface](https://github.com/go-shiori/shiori/blob/master/docs/Usage.md#using-command-line-interface) with the command `shiori pocket` to import Pocket entries but with this can only import them as links and not as complete entries.

To import your bookmarks from [Pocket](https://getpocket.com/) with the text and images follow these simple steps (based on [Issue 252](https://github.com/go-shiori/shiori/issues/252)):

1. Export your entries from Pocket by visiting https://getpocket.com/export

2. Download [this shell script](https://gist.github.com/fmartingr/88a258bfad47fb00a3ef9d6c38e5699e). [*You need to download this in your docker container or on the device that you are hosting shiori*]. Name it for instance `pocket2shiori.sh`.

   > Tip: checkout the documentation for [opening a console in the docker container](https://github.com/go-shiori/shiori/blob/master/docs/Usage.md#running-docker-container).

3. Execute the shell script.

Here are the commands you need to run:
   ```sh
   wget 'https://gist.githubusercontent.com/fmartingr/88a258bfad47fb00a3ef9d6c38e5699e/raw/a21afb20b56d5383b8b975410e0eb538de02b422/pocket2shiori.sh'
   chmod +x pocket2shiori.sh
   pocket2shiori.sh 'path_to_your/pocket_export.html'
   ```

   > Tip: If you’re using shiori's docker container, ensure that the exported HTML from pocket is accessible inside the docker container.

You should now see `shiori` importing your Pocket entries properly with the text and images.
This is optional, but once the import is complete you can clean up by running:

```sh
rm pocket2shiori.sh 'path_to_your/pocket_export.html'
```

##  Import from Wallabag

1. Export your entries from Wallabag as a json file

2. Install [jq](https://stedolan.github.io/jq/download/). You will need this installed before running the script.

3. Download the shell script
[here](https://gist.githubusercontent.com/Aerex/01499c66f6b36a5d997f97ca1b0ab5b1/raw/bf793515540278fc675c7769be74a77ca8a41e62/wallabag2shiori). Similar to the `pocket2shiori.sh` script if you are shiori is in a docker container you will next to run this script
inside the container.

4. Execute the script. Here are the commands that you can run.

  ```sh
    curl -sSOL
    https://gist.githubusercontent.com/Aerex/01499c66f6b36a5d997f97ca1b0ab5b1/raw/bf793515540278fc675c7769be74a77ca8a41e62/wallabag2shiori'
    chmod +x wallabag2shiori
    ./wallabag2shiori 'path/to/to/wallabag_export_json_file'
  ```
