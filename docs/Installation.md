There are several installation methods available :

<!-- TOC -->

- [Supported](#supported)
  - [Using Precompiled Binary](#using-precompiled-binary)
  - [Building From Source](#building-from-source)
  - [Using Docker Image](#using-docker-image)
- [Community provided](#community-provided)
  - [Using Kubernetes manifests](#using-kubernetes-manifests)

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
      containers:
      - name: shiori
        image: ghcr.io/go-shiori/shiori:latest
        command: ["/usr/bin/shiori", "serve", "--webroot", "/shiori"]
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        volumeMounts:
        - mountPath: /srv/shiori
          name: app
        env:
        - name: SHIORI_DIR
          value: /srv/shiori
        - name: HTTP_ROOT_PATH
          value: "/shiori"
```

Here we are using a local directory to persist Shiori's data. You will need
to replace `/path/to/data/dir` with the path to the directory where you want
to keep your data. Since we haven't configured a database in particular,
Shiori will use SQLite. I don't think Postgres or MySQL is worth it for
such an app, but that's up to you. If you decide to use SQLite, I strongly
suggest to keep `replicas` set to 1 since SQLite usually allows at most
one writer to proceed concurrently.

Also, not that we're serving the app on the `/shiori` suffix. This is
only necessary if you want to access Shiori with an URL that looks like:
`http://your_domain_name/shiori`. This is also why we override the container's
command: to pass the webroot. If you want to use such suffix, you'll probably
need to deploy an ingress as well:

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
      - path: /shiori
        pathType: Prefix
        backend:
          service:
            name: shiori
            port:
              number: 8080
```

Finally, here is the service's config:

`service.yaml`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: shiori
spec:
  type: NodePort
  selector:
    app: shiori
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
    nodePort: 32654
```

I'm using the NodePort type for the service so I can access it easily on
my local network, but it's not necessary if you setup the ingress.
