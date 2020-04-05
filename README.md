# skaffold-generator

CLI tool to generate skaffold.yaml to build and deploy only required services

## Motivation

When we migrate [Docker Compose](https://docs.docker.com/compose/) to [Skaffold](https://skaffold.dev/) to develop k8s application at local,
Skaffold doesn't provide some features which Docker Compose provides.

* Launch only specified services (`docker-compose up [<service name> ...]`)
* Manage dependencies (`depends_on`)

Please imagine that many microservices are managed by one skaffold.yaml.
When we develop a service `A`,
it is convenient to build and deploy only the service `A` and services which the service `A` depends on.

So this tool `skaffold-generator` provides these features.

`skaffold-generator` watches the configuration file `skaffold-generator.yaml` and updates `skaffold.yaml` every time `skaffold-generator.yaml` is updated.

In `skaffold-generator.yaml`, we can define `services` and each service's `artifacts` and `manifests`, and service dependencies.

The command `skaffold-generator` takes service names as arguments, and generate `skaffold.yaml` only for the specified services.

Even if over one hundred services are defined in `skaffold-generator.yaml`,
when we run `slaffold-generator blog`, `skaffold.yaml` is generated based on only the service `blog` and services which `blog` depends on.

`skaffold-generator` doesn't use `skaffold`. `skaffold-generator` just update `skaffold.yaml`.
By running `skaffold dev` while running `skaffold-generator`,
`skaffold dev` detects the update by `skaffold-generator` and updates application actually.

## Status

Currently, this tool only supports to overwrite `deploy.kubectl.manifests` and `build.artifacts`.

## Install

Download the binary from the [release page](https://github.com/suzuki-shunsuke/skaffold-generator/releases).

## Getting Started

```
$ git clone https://github.com/suzuki-shunsuke/skaffold-generator
$ cd skaffold-generator/examples
$ tree
.
├── api
│   └── Dockerfile
├── api.yaml
├── foo
│   └── Dockerfile
├── foo.yaml
├── mongodb.yaml
├── skaffold-generator.yaml
└── skaffold.yaml

2 directories, 7 files
```

Currently, `skaffold.yaml` doesn't exist.
Let's run `skaffold-generator`.

```
$ skaffold-generator
2020/04/05 18:19:37 start to watch skaffold-generator.yaml
```

`skaffold-generator` watches the update of `skaffold-generator.yaml` until `Ctrl-C` is clicked.

On the other terminal, please confirm `skaffold.yaml` is generated.

When `skaffold-generator` doesn't take any arguments, `skaffold.yaml` is generated based on all services.

Let's change the image name `foo` to `bar`, and confirm that `skaffold-generator` detect the change and `skaffold.yaml` is updated.

```
$ skaffold-generator
2020/04/05 18:19:37 start to watch skaffold-generator.yaml
2020/04/05 18:27:17 detect the update of skaffold-generator.yaml
```

After confirmation, please reset the change of image name.

By running `skaffold dev`, we can build and deploy applications with `Skaffold`.

```
$ skaffold dev
```

Once stop the command `skaffold-generator` by clicking `Ctrl-C`, and rerun `skaffold-generator` with arguments.

```
$ skaffold-generator api
2020/04/05 18:40:31 start to watch skaffold-generator.yaml
```

The argument `api` is a service name.
We can confirm that `skaffold.yaml` is updated and `artifacts` and `manifests` about `foo` are removed.
We don't specify `mongodb` by arguments but the service `api` depends on `mongodb` so `artifacts` and `manifests` of `mongodb` are also included in `skaffold.yaml`.

## Usage

```
$ skaffold-generator --help
NAME:
   skaffold-generator - generate skaffold.yaml

USAGE:
   skaffold-generator [global options] [arguments...]

VERSION:
   0.1.0-0

GLOBAL OPTIONS:
   --src value, -s value   configuration file path (skaffold-generator.yaml) (default: "skaffold-genera
tor.yaml")
   --dest value, -d value  generated configuration file path (skaffold.yaml) (default: "skaffold.yaml")
   --help, -h              show help (default: false)
   --version, -v           print the version (default: false)
```

## Configuration

skaffold-generator.yaml

```yaml
---
services:
- name: mongodb # service name. This is used to specify the service by command line arguments and depends_on
  manifests: # skaffold.yaml's `deploy.kubectl.manifests`
    - "mongodb.yaml"
  artifacts: # skaffold.yaml's `build.artifacts`
    - image: mongodb
      context: .
      docker:
        dockerfile: mongodb/Dockerfile
      sync:
        infer:
          - 'mongodb/etc/**/*'
- name: api
  depends_on: # service names which this service depends on
    - mongodb
  manifests:
    - "api.yaml"
  artifacts:
    - image: api
      context: .
      docker:
        dockerfile: api/Dockerfile
# base `skaffold.yaml`.
# `deploy.kubectl.manifests` and `build.artifacts` are overwritten.
base:
  apiVersion: skaffold/v2alpha4
  kind: Config
```

## Change detection

To detect the update of `skaffold-generator.yaml`, we use the third party library [github.com/radovskyb/watcher](https://github.com/radovskyb/watcher).

## License

[MIT](LICENSE)
