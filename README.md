# mongoperf
`mongoperf` provides a CLI for running predefined scenarios against a mongodb database and report performance statistics.

## Overview
`mongoperf` provides a CLI for running predefined scenarios against a mongodb database and report performance statistics.

Currently, **`mongoperf` requires Go version 1.13 or greater**.

## Table of Contents

- [Overview](#overview)
- [Development](#development)
- [Build](#build)
- [CLI](#cli)
  - [Usage](#usage)

## Development
You can use the provided `Dockerfile` to build an image that will provide a clean environment for development purposes.</br>
Instructions that follow assumes you are running `Windows`, have `Docker Desktop` installed and its daemon is running.

Clone this repository and build the image
```
$ git clone https://github.com/devodev/mongoperf
$ cd ./mongoperf
$ docker build --tag=mongoperf .
```

Run a container using the previously built image while mounting the CWD
```
$ docker run \
    --rm \
    --volume="$(pwd -W):/srv/src/github.com/devodev/mongoperf" \
    --tty \
    --interactive \
    mongoperf
$ root@03e67598a37f:/srv/src/github.com/devodev/mongoperf#
```

Start deving
```
$ go run ./cmd/mongoperf
```

### Build
Build the CLI for a target platform (Go cross-compiling feature), for example linux, by executing:
```
$ mkdir $HOME/src
$ cd $HOME/src
$ git clone https://github.com/devodev/mongoperf.git
$ cd mongoperf
$ env GOOS=linux go build -o mongoperf_linux ./cmd/mongoperf
```
If you are a Windows user, substitute the $HOME environment variable above with %USERPROFILE%.

## CLI
### Usage
```
Run performance tests scenarios on a mongodb instance or cluster.

Usage:
  mongoperf [command]

Available Commands:
  demo        Run small demo that inserts, update and delete entries.
  help        Help about any command
  scenario    Run a scenario.

Flags:
  -h, --help      help for mongoperf
      --version   version for mongoperf

Use "mongoperf [command] --help" for more information about a command.
```
