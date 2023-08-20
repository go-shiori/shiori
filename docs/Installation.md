# Installation

[Go back to documentation index](./index.md)

There are several supported installation methods available :

## Using Precompiled Binary

Download the latest version of `shiori` from [the release page](https://github.com/go-shiori/shiori/releases/latest), then put it in your `PATH`.

On Linux or MacOS, you can do it by adding this line to your profile file (either `$HOME/.bash_profile` or `$HOME/.profile`):

``` sh
export PATH=$PATH:/path/to/shiori
```

Note that this will not automatically update your path for the remainder of the session. To do this, you should run:

``` sh
source $HOME/.bash_profile
or
source $HOME/.profile
```

On Windows, you can simply set the `PATH` by using the advanced system settings.

## Install building from source

Shiori uses Go module so make sure you have version of `go >= 1.14.1` installed, then run:

``` sh
go install github.com/go-shiori/shiori@latest
```

> Ensure your `PATH` includes the `$GOPATH/bin` directory so your commands can be easily used:
>
> ``` sh
> export PATH=$PATH:$(go env GOPATH)/bin
> ```

## Using Docker Image

To use Docker image, you can pull the latest automated build from Docker Hub :

```
docker pull ghcr.io/go-shiori/shiori
```

> This docker image is based on the [Dockerfile.ci](../.github/workflows/docker/Dockerfile.ci)

## Building docker image

If you want to build the Docker image on your own, Shiori already has its [Dockerfile](https://github.com/go-shiori/shiori/blob/master/Dockerfile), so you can build the Docker image by running :

```
docker build -t shiori .
```

> This docker image is based on the root [Dockerfile](../Dockerfile)
