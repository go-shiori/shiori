# Getting Started

Before using `shiori`, make sure it has been installed on your system. By default, `shiori` will store its data in directory `$HOME/.local/share/shiori`. If you want to set the data directory to another location, you can set the environment variable `SHIORI_DIR` (`ENV_SHIORI_DIR` when you are before `1.5.0`) to your desired path.

<!-- TOC -->

- [Running Docker Container](#running-docker-container)
- [How to use Shiori](#how-to-use-shiori)
- [Deleting/Stopping Shiori](#deletingstopping-shiori)

<!-- /TOC -->

## Running Docker Container

> If you are not using `shiori` from Docker image, you can skip this section.

After building or pulling the image, you will be able to start a container from it. To preserve the data, you need to bind the directory for storing database, thumbnails and archive. In this example we're binding the data directory to our current working directory :

```
docker run -d --rm --name shiori -p 8080:8080 -v $(pwd):/shiori ghcr.io/go-shiori/shiori
```

The above command will:

- Creates a new container from image `ghcr.io/go-shiori/shiori`.
- Set the container name to `shiori` (option `--name`).
- Bind the host current working directory to `/shiori` inside container (option `-v`).
- Expose port `8080` in container to port `8080` in host machine (option `-p`).
- Run the container in background (option `-d`).
- Automatically remove the container when it stopped (option `--rm`).

After you've run the container in background, you can access console of the container:

```
docker exec -it shiori sh
```

Now you can use `shiori` like normal.

## How to use Shiori

From this point you have the option to proceed to the [usage section](./Usage.md), where various modes of operation are elaborated upon:

- [How to use the Web Interface](./Usage.md#using-web-interface)
- [Using the CLI](./Usage.md#using-command-line-interface)
- [Improved import from Pocket](./Usage.md#improved-import-from-pocket)
- [Import from Wallabag](./Usage.md#import-from-wallabag)


## Deleting/Stopping Shiori

If you've finished, you can stop and remove the container by running :

```
docker stop shiori
```

