Content
---

<!-- TOC -->

- [Add bookmark](#add-bookmark)

<!-- /TOC -->

Add bookmark
---

To add bookmark with CLI you can use `shiori add`.

Shiori has flags to add bookmark: `shiori add --help`

```
Bookmark the specified URL

Usage:
  shiori add url [flags]

Flags:
  -e, --excerpt string   Custom excerpt for this bookmark
  -h, --help             help for add
      --log-archival     Log the archival process
  -a, --no-archival      Save bookmark without creating offline archive
  -o, --offline          Save bookmark without fetching data from internet
  -t, --tags strings     Comma-separated tags for this bookmark
  -i, --title string     Custom title for this bookmark

Global Flags:
      --log-caller                 logrus report caller or not
      --log-level string           set logrus loglevel (default "info")
      --portable                   run shiori in portable mode
      --storage-directory string   path to store shiori data
```

Examples:

Add url:
`shiori add https://example.com`

Add url with tags:
`shiori add https://example.com -t "example-1,example-2"`

Add url with custom title:
`shiori add https://example.com --title "example example"`
