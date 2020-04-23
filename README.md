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

### Scenario Command
Takes in a scenario configuration file and runs it.</br>
Queries are sent to workers sequentially.

#### Schema
The current schema is represented using a yaml configuration file.</br>
It contains a single Scenario object containing configuration attributes as well as query definitions.</br>
Here is an example:
```
---
Database: test
Collection: test
Parallel: 2
BufferSize: 1000
Repeat: 1000
Queries:
- Name: testmany
  Action: InsertMany
  Meta:
    Data:
    - Name: Ash
      Age: 10
      City: Pallet Town
    - Name: Misty
      Age: 10
      City: Cerulean City
```

A scenario declares the following attributes:
- Database (string)
  - A MongoDB database name.
- Collection (string)
  - A MongoDB collection name.
- Parallel (int, optional) (default: 1)
  - The number of parallel workers to process queries.
  - Must be greater than 0.
- BufferSize (int, optional) (default: 1000)
  - The size of the worker task queue.
  - Must be greater than or equal to 0.
  - If 0, default value is set (1000).
- Repeat (int, optional) (default: 1)
  - How many times should we run the queries.
  - Must be greater than or equal to 0.
  - If 0, repeats indefinitely.
- Queries (List<Query>)
  - Must contain at least one Query definition.

A `Scenario` also declares a `Queries` attribute, which is a list of `Query` definition.
```
---
...escaped
Queries:
- Name: test
  Action: InsertMany
  Meta:
...escaped
```
A query declares the following attributes:
- Name (string)
  - Used as an identifier for the query.
- Action (string)
  - Defines the action to perform against a collection. Available are:
    - InsertOne
    - InsertMany
    - UpdateOne
    - FindOne
    - Find
- Meta (Meta)
  - An object specific to the Action provided.

A `Query` also declares a `Meta` object which contains the payload specific attributes required by the specified Action attriobte.
```
---
...escaped
Queries:
...escaped
  Meta:
    Data:
    - Name: Ash
      Age: 10
      City: Pallet Town
    - Name: Misty
      Age: 10
      City: Cerulean City
    Options:
      Ordered: true
```

Here is a list of schema used for each Action

InsertOne
  - Data (map)
    - A map of key/values representing the document to insert.
  - Options (map, optional)
    - A map of key/values correspongind to the InsertOneOptions type.
      - BypassDocumentValidation (bool)

InsertMany
  - Data (List<map>)
    - A list of map of key/values representing the documents to insert.
  - Options (map, optional)
    - A map of key/values correspongind to the InsertManyOptions type.
      - BypassDocumentValidation (bool)
      - Ordered (bool)

UpdateOne
  - Data (map)
    - A map of key/values representing a document containing update operators.
  - Filter (map)
    - A map of key/values representing the filter to apply.
  - Options (map, optional)
    - A map of key/values correspongind to the UpdateOptions type.
      - ArrayFilters (ArrayFilters)
      - BypassDocumentValidation (bool)
      - Ordered (bool)
      - Collation (Collation)
      - Upsert (bool)

FindOne
  - Filter (map)
    - A map of key/values representing the filter to apply.
  - Options (map, optional)
    - A map of key/values correspongind to the UpdateOptions type.
      - AllowPartialResults (bool)
      - BatchSize (int)
      - Collation (Collation)
      - Comment (string)
      - CursorType (CursorType)
      - Hint (string | map)
      - Max (map)
      - MaxAwaitTime (duration)
      - MaxTime (duration)
      - Min (map)
      - NoCursorTimeout (bool)
      - OplogReplay (bool)
      - Projection (map)
      - ReturnKey (bool)
      - ShowRecordID (bool)
      - Skip (int)
      - Snapshot (bool)
      - Sort (map)

Find
  - Filter (map)
    - A map of key/values representing the filter to apply.
  - Options (map, optional)
    - A map of key/values correspongind to the UpdateOptions type.
      - AllowPartialResults (bool)
      - BatchSize (int)
      - Collation (Collation)
      - Comment (string)
      - CursorType (CursorType)
      - Hint (string | map)
      - Limit (int)
      - Max (map)
      - MaxAwaitTime (duration)
      - MaxTime (duration)
      - Min (map)
      - NoCursorTimeout (bool)
      - OplogReplay (bool)
      - Projection (map)
      - ReturnKey (bool)
      - ShowRecordID (bool)
      - Skip (int)
      - Snapshot (bool)
      - Sort (map)
