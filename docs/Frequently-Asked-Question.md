# Frequently asked questions

<!-- TOC -->

- [Frequently asked questions](#frequently-asked-questions)
- [General](#general)
    - [What is this project ?](#what-is-this-project-)
    - [How does it compare to other bookmarks manager ?](#how-does-it-compare-to-other-bookmarks-manager-)
    - [What are the system requirements ?](#what-are-the-system-requirements-)
    - [What is the status for this app ?](#what-is-the-status-for-this-app-)
    - [Is this app actively maintained ?](#is-this-app-actively-maintained-)
    - [How to make a contribution ?](#how-to-make-a-contribution-)
    - [How to make a donation ?](#how-to-make-a-donation-)
- [Common Issues](#common-issues)
    - [What is the default account to login at the first time ?](#what-is-the-default-account-to-login-at-the-first-time-)
    - [Why my old accounts can't do anything after upgrading Shiori to v1.5.0 ?](#why-my-old-accounts-cant-do-anything-after-upgrading-shiori-to-v150-)
    - [Failed to get bookmarks: failed to fetch data: no such module: fts4 ?](#failed-to-get-bookmarks-failed-to-fetch-data-no-such-module-fts4-)
- [Advanced](#advanced)
    - [How to run shiori on start up ?](#how-to-run-shiori-on-start-up-)

<!-- /TOC -->

# General

## What is this project ?

Shiori is a bookmarks manager that built with Go. I've got the idea to make this after reading a comment on HN back in [April 2017](https://news.ycombinator.com/item?id=14203383) :

```
... for me the dream bookmark manager would be something really simple
with two commands like:

$ bookmark add http://...

That will:

a. Download a static copy of the webpage in a single HTML file, with a
   PDF exported copy, that also take care of removing ads and
   unrelated content from the stored content.
b. Run something like http://smmry.com/ to create a summary of the page
   in few sentences and store it.
c. Use NLP techniques to extract the principle keywords and use them
   as tags

And another command like:

$ bookmark search "..."

That will:

d. Not use regexp or complicated search pattern, but instead;
e. Search titles, tags, page content smartly and interactively, and;
f. Sort/filter results smartly by relevance, number of matches,
   frequency, or anything else useful
g. Storing everything in a git repository or simple file structure
   for easy synchronization, bonus point for browsers integrations.
```

I do like using bookmarks and those idea sounds useful to me. More importantly, it seems possible enough to do. Not too hard that it's impossible for me, but not too easy that it doesn't teach me anything. Looking back now, the only thing that I (kind of) managed to do is a, b, d and e. But it's enough for me, so it's fine I guess :laughing:.

## How does it compare to other bookmarks manager ?

To be honest I don't know. The only bookmarks manager that I've used is Pocket and the one that bundled in web browser. I do like Pocket though. However, since bookmarks is kind of sensitive data, I prefer it stays offline or in my own server.

## What are the system requirements ?

It runs in the lowest tier of Digital Ocean VPS, so I guess it should be able to run anywhere.

## What is the status for this app ?

It's stable enough to use and the database shouldn't be changed anymore. However, my bookmarks at most is only several hundred entries, therefore I haven't test whether it able to process or imports huge amount of bookmarks. If you would, please do try it.

## Is this app actively maintained ?

Yes, however the development pace might be really slow. @fmartingr is the current active maintainer though @RadhiFadlillah or @deanishe may step and work on stuff from time to time or in other [go-shiori projects](https://github.com/go-shiori)

## How to make a contribution ?

Just like other open source projects, you can make a contribution by submitting issues or pull requests.

## How to make a donation ?

If you like this project, you can donate to maintainers via:

- **fmartingr** [PayPal](https://www.paypal.me/fmartingr), [Ko-Fi](https://ko-fi.com/fmartingr)
- **RadhiFadlillah** [PayPal](https://www.paypal.me/RadhiFadlillah), [Ko-Fi](https://ko-fi.com/radhifadlillah)

# Common Issues


## What is the default account to login at the first time ?

The default account is `shiori` with password `gopher`.
It is removed once another 'owner' account is created.

## Why my old accounts can't do anything after upgrading Shiori to v1.5.0 ?

This issue happened because in Shiori v1.0.0 there are no account level, which means everyone is treated as owner. However, in Shiori v1.5.0 there are two account levels i.e. owner and visitor. The level difference is stored in [database](https://github.com/go-shiori/shiori/blob/master/internal/database/sqlite.go#L42-L48) as boolean value in column `owner` with default value false (which means by default all account is visitor, unless specified otherwise).

Because in v1.5.0 by default all account is visitor, when updating from v1.0 to v1.5 all of the old accounts by default will be marked as visitor. Fortunately, when there are no owner registered in database, we can login as owner using default account.

So, as workaround for this issue, you should :

- Login as default account.
- Go to options page.
- Remove your old accounts.
- Recreate them, but now as owner.

For more details see [#148](https://github.com/go-shiori/shiori/issues/148).

## `Failed to get bookmarks: failed to fetch data: no such module: fts4` ?

This happens to SQLite users that upgrade from 1.5.0 to 1.5.1 because of a breaking change. Please check the
[announcement](https://github.com/go-shiori/shiori/discussions/383) to understand how to migrate your database and move forward.

# Advanced


## How to run `shiori` on start up ?

There are several methods to run `shiori` on start up, however the most recommended is running it as a service.

1. Create a service unit for `systemd` at `/etc/systemd/system/shiori.service`.

* Shiori is run via `docker` :

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
      -v /srv/machines/shiori:/shiori \
       ghcr.io/go-shiori/shiori
    ExecStop=/usr/bin/docker stop -t 2 shiori-1

    [Install]
    WantedBy=multi-user.target
    ```

* Shiori without `docker`. Set absolute path to `shiori` binary. `--portable` sets the data directory to be alongside the executable.

    ```ini
    [Unit]
    Description=Shiori service

    [Service]
    ExecStart=/home/user/go/bin/shiori serve --portable
    Restart=always

    [Install]
    WantedBy=multi-user.target
    ```

* Shiori without `docker` and without `--portable` but secure.

   ```ini
   [Unit]
   Description=shiori service
   Requires=network-online.target
   After=network-online.target

   [Service]
   Type=simple
   ExecStart=/usr/bin/shiori serve
   Restart=always
   User=shiori
   Group=shiori

   Environment="SHIORI_DIR=/var/lib/shiori"
   DynamicUser=true
   PrivateUsers=true
   ProtectHome=true
   ProtectKernelLogs=true
   RestrictAddressFamilies=AF_INET AF_INET6
   StateDirectory=shiori
   SystemCallErrorNumber=EPERM
   SystemCallFilter=@system-service
   SystemCallFilter=~@chown
   SystemCallFilter=~@keyring
   SystemCallFilter=~@memlock
   SystemCallFilter=~@setuid
   DeviceAllow=

   CapabilityBoundingSet=
   LockPersonality=true
   MemoryDenyWriteExecute=true
   NoNewPrivileges=true
   PrivateDevices=true
   PrivateTmp=true
   ProtectControlGroups=true
   ProtectKernelTunables=true
   ProtectSystem=full
   ProtectClock=true
   ProtectKernelModules=true
   ProtectProc=noaccess
   ProtectHostname=true
   ProcSubset=pid
   RestrictNamespaces=true
   RestrictRealtime=true
   RestrictSUIDSGID=true
   SystemCallArchitectures=native
   SystemCallFilter=~@clock
   SystemCallFilter=~@debug
   SystemCallFilter=~@module
   SystemCallFilter=~@mount
   SystemCallFilter=~@raw-io
   SystemCallFilter=~@reboot
   SystemCallFilter=~@privileged
   SystemCallFilter=~@resources
   SystemCallFilter=~@cpu-emulation
   SystemCallFilter=~@obsolete
   UMask=0077

   [Install]
   WantedBy=multi-user.target
   ```

2. Set up data directory if Shiori with `docker`

    This assumes, that the Shiori container has a runtime directory to store their
    database, which is at `/srv/machines/shiori`. If you want to modify that,
    make sure, to fix your `shiori.service` as well.

    ```sh
    install -d /srv/machines/shiori
    ```

3. Enable and start the service

    ```sh
    systemctl enable --now shiori
    ```
