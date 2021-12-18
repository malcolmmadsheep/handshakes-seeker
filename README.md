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

To investigate all available commands just run:

```bash
make
```

## TODO

- [x] bootstrap repo
- [x] configure docker-compose
- [x] implement queue
- [x] configure Postgres container and setup connection within the service
- [ ] setup db migrations (go-migrate)
- [ ] create tables in db
- [ ] implement application logic
- [ ] implement wikipedia plugin logic
- [ ] add opportunity to configure using file and env variables
- [ ] setup graceful shutdown
- [ ] refactor
- [ ] store secretes in env files
- [ ] add tests
- [ ] find a way for service scaling (running multiple instances in parallel without repeating tasks)

