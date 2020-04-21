# mongo-tester
`mongo-tester` provides a CLI for running predefined scenarios against a mongodb database and report performance statistics.

## Overview
`mongo-tester` provides a CLI for running predefined scenarios against a mongodb database and report performance statistics.

Currently, **`mongo-tester` requires Go version 1.13 or greater**.

## Table of Contents

- [Overview](#overview)
- [Development](#development)
- [Build](#build)
- [CLI](#cli)
  - [Usage](#usage)
  - [Scenario File](#scenario-file)
    - [Definition](#definition)
    - [Example](#example)

## Development
You can use the provided `Dockerfile` to build an image that will provide a clean environment for development purposes.</br>
Instructions that follow assumes you are running `Windows`, have `Docker Desktop` installed and its daemon is running.

Clone this repository and build the image
```
$ git clone https://github.com/devodev/mongo-tester
$ cd ./mongo-tester
$ docker build --tag=mongo-tester .
```

Run a container using the previously built image while mounting the CWD
```
$ docker run \
    --rm \
    --volume="$(pwd -W):/srv/src/github.com/devodev/mongo-tester" \
    --tty \
    --interactive \
    mongo-tester
$ root@03e67598a37f:/srv/src/github.com/devodev/mongo-tester#
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
$ git clone https://github.com/devodev/mongo-tester.git
$ cd mongo-tester
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

### Scenario File
Current schema used for defining scenarios.

#### Definition
- Database: mongodb database.
- Collection: mongodb collection.
- Queries: List of ScenarioQuery. Contains individual query to execute against the Scenario collection.
- ScenarioQuery.Name: Used for logging purposes.
- ScenarioQuery.Action: The MongoDB query type to execute. Ex.: InsertOne, InsertMany, etc.
- ScenarioQuery.Meta: List of ScenarioMeta. Contains abstract payload data used by Actions.
- ScenarioMeta.Payload: A BSON.D (map) representation of the data needed by ScenarioQuery.Action.
- ScenarioMeta.PayloadList: A BSON.A (list) representation of the data needed by ScenarioQuery.Action.

#### Example
```
---
Scenario:
  Database: test
  Collection: test
  Queries:
  - Name: testone
    Action: InsertOne
    Meta:
      Payload:
        Name: Ash
        Age: 10
        City: Pallet Town
      PayloadList:
  - Name: testmany
    Action: InsertMany
    Meta:
      Payload:
      PayloadList:
      - Name: Misty
        Age: 10
        City: Cerulean City
      - Name: Brock
        Age: 15
        City: Pewter City
```
