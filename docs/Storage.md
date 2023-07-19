# Storage

Shiori requires a folder to store several pieces of data, such as the bookmark archives, thumbnails, ebooks, and others. If the database engine used is sqlite, then the database file will also be stored in this folder.

You can specify the storage folder by using `--storage-dir` or `--portable` flags when running Shiori.

If none specified, Shiori will try to find the correct app folder for your OS.

For example:
- In Windows, Shiori will use `%APPDATA%`.
- In Linux, it will use `$XDG_CONFIG_HOME` or `$HOME/.local/share` if `$XDG_CONFIG_HOME` is not set.
- In macOS, it will use `$HOME/Library/Application Support`.

> For more and up to date information about app folder discovery check [muesli/go-app-paths](https://github.com/muesli/go-app-paths)
