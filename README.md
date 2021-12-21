# Handshakes Seeker

## Overview

Handshakes Seeker is a service for finding shortest path between 2 entities in a knowledge base.

## Prerequisites

1. Docker

## Usage

To setup Docker volume and create required database:

```bash
make setup-db
```

This command creates `handshakes_db` volume for Postgres and fulfill it with `handshakes` database.

To spin up application locally run:

```bash
make run-dev
```

To rebuild app:

```bash
make build-compose
```

To create migration:

```bash
. ./scripts/create-migration.sh <migration_name> # under the hood it calls make add-migration

# or

MIGRATION_NAME=<migration_name> make add-migration
```

To investigate all available commands just run:

```bash
make
```

## Configuration

Env variables, that can be passed to service:

- `HANDSHAKES_WIKI_PLUGIN_DELAY` - positive number, Wikipedia plugin delay between requests
- `HANDSHAKES_WIKI_QUEUE_SIZE` - positive number, Wikipedia plugin queue size

## TODO

- [x] bootstrap repo
- [x] configure docker-compose
- [x] implement queue
- [x] configure Postgres container and setup connection within the service
- [x] setup db migrations (go-migrate)
- [x] create tables in db
- [x] implement application logic
- [x] implement wikipedia plugin logic
- [x] add opportunity to configure using env variables
- [ ] ☹️ setup graceful shutdown
- [ ] ☹️ refactor
- [ ] ☹️ store secretes in env files
- [ ] ☹️ add tests
- [ ] ☹️ find a way for service scaling (running multiple instances in parallel without repeating tasks)


