# Handshakes Seeker

## Overview

Handshakes Seeker is a service for finding shortest path between 2 entities in a knowledge base.

## Prerequisites

1. Docker

## Usage

To spin up application locally run:

```bash
make run-dev
```

To investigate all available commands just run:

```bash
make
```

## TODO

- [x] bootstrap repo
- [x] configure docker-compose
- [x] implement queue
- [ ] configure Postgres container and setup connection within the service
- [ ] setup db migrations (go-migrate)
- [ ] create tables in db
- [ ] implement application logic
- [ ] implement wikipedia plugin logic
- [ ] add opportunity to configure using file and env variables
- [ ] setup graceful shutdown
- [ ] refactor
- [ ] add tests
- [ ] find a way for service scaling (running multiple instances in parallel without repeating tasks)

