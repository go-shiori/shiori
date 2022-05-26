There are several installation methods available :

<!-- TOC -->

- [Using Precompiled Binary](#using-precompiled-binary)
- [Building From Source](#building-from-source)
- [Using Docker Image](#using-docker-image)

<!-- /TOC -->

## Using Precompiled Binary

Download the latest version of `shiori` from [the release page](https://github.com/go-shiori/shiori/releases/latest), then put it in your `PATH`.

On Linux or MacOS, you can do it by adding this line to your profile file (either `$HOME/.bash_profile` or `$HOME/.profile`):

```
export PATH=$PATH:/path/to/shiori
```

Note that this will not automatically update your path for the remainder of the session. To do this, you should run:

```
source $HOME/.bash_profile
or
source $HOME/.profile
```

On Windows, you can simply set the `PATH` by using the advanced system settings.

## Building From Source

Shiori uses Go module so make sure you have version of `go >= 1.14.1` installed, then run:

```
go get -u -v github.com/go-shiori/shiori
```

## Using Docker Image

To use Docker image, you can pull the latest automated build from Docker Hub :

```
docker pull ghcr.io/go-shiori/shiori
```

If you want to build the Docker image on your own, Shiori already has its [Dockerfile](https://github.com/go-shiori/shiori/blob/master/Dockerfile), so you can build the Docker image by running :

```
docker build -t shiori .
```
