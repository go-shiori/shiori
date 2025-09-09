There are several installation methods available :

<!-- TOC -->

- [Supported](#supported)
  - [Using Precompiled Binary](#using-precompiled-binary)
  - [Building From Source](#building-from-source)
  - [Using Docker Image](#using-docker-image)
- [Community provided](#community-provided)
  - [Using Kubernetes manifests](#using-kubernetes-manifests)
- [Managed Hosting](#managed-hosting)
  - [PikaPods](#pikapods)

<!-- /TOC -->

## Supported

### Using Precompiled Binary

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

### Building From Source

Shiori uses Go module so make sure you have version of `go >= 1.14.1` installed, then run:

```
go get -u -v github.com/go-shiori/shiori
```

### Using Docker Image

To use Docker image, you can pull the latest automated build from Docker Hub :

```
docker pull ghcr.io/go-shiori/shiori
```

If you want to build the Docker image on your own, Shiori already has its [Dockerfile](https://github.com/go-shiori/shiori/blob/master/Dockerfile), so you can build the Docker image by running :

```
docker build -t shiori .
```

## Community provided

Below this there are other ways to deploy Shiori which are not supported by the team but were provided by the community to help others have a starting point.

### Using Kubernetes manifests

If you're self-hosting with a Kubernetes cluster, here are manifest files that
you can use to deploy Shiori:

`deploy.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: shiori
  labels:
    app: shiori
spec:
  replicas: 1
  selector:
    matchLabels:
      app: shiori
  template:
    metadata:
      labels:
        app: shiori
    spec:
      volumes:
      - name: app
        hostPath:
          path: /path/to/data/dir
      - name: tmp
        emptyDir:
          medium: Memory
      containers:
      - name: shiori
        image: ghcr.io/go-shiori/shiori:latest
        command: ["/usr/bin/shiori", "serve"]
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        env:
        - name: SHIORI_DIR
          value: /srv/shiori
        volumeMounts:
        - mountPath: /srv/shiori
          name: app
        - mountPath: /tmp
          name: tmp
```

Here we are using a local directory to persist Shiori's data. You will need
to replace `/path/to/data/dir` with the path to the directory where you want
to keep your data. We are also mounting an `EmptyDir` volume for `/tmp` so
we can successfully generate ebooks.

Since we haven't configured a database in particular,
Shiori will use SQLite. I don't think Postgres or MySQL is worth it for
such an app, but that's up to you. If you decide to use SQLite, I strongly
suggest to keep `replicas` set to 1 since SQLite usually allows at most
one writer to proceed concurrently.

To route requests to your deployment, you will need a `Service` that gets used
by an `Ingress` to handle routing. If you wand to add a path suffix or use a
sub domain, you can do so through the ingress config. We only show the bare
minimum config to get you started.

`service.yaml`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: shiori
spec:
  type: LoadBalancer
  selector:
    app: shiori
  ports:
    - port: 8080
      targetPort: 8080
```

This is using a `LoadBalancer` type which gives the most flexibility.

`ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: shiori
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: shiori
            port:
              number: 8080
```

## Managed Hosting

If you don't manage your own server, the below providers will host Shiori for you. None are endorsed by or affiliated with the team. Support is provided by the providers.

### CloudBreak

[CloudBreak](https://cloudbreak.app/products/shiori?utm_medium=referral&utm_source=shiori-docs&rby=shiori-docs) offers Shiori hosting from $12/year ($1/month).  Get $3 off with coupon `SHIORI`.

<a href="https://cloudbreak.app/products/shiori?utm_medium=referral&utm_source=shiori-docs&rby=shiori-docs">
  <img src="https://cloudbreak.app/external/subscribe-button.png" alt="Subscribe on CloudBreak" width="149" height="64">
</a>

### PikaPods

[PikaPods](https://www.pikapods.com/) offers Shiori hosting from $1.20/month with $5 free welcome credit. EU and US regions available. Updates are applied weekly and user data backed up daily.

[![Run on PikaPods](https://www.pikapods.com/static/run-button.svg)](https://www.pikapods.com/pods?run=shiori)
