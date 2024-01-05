# Project go-notes

Secure and scalable RESTful API that allows users to create, read, update, and delete notes. The application should also allow users to share their notes with other users and search for notes based on keywords.

## Getting Started

golang should be installed
For api overview -> [Postman](https://elements.getpostman.com/redirect?entityId=20103509-452fa703-9f86-44b5-ad92-471ca7bd42b8&entityType=collection)

## MakeFile

run all make commands with clean tests

```bash
make all build
```

build the application

```bash
make build
```

run the application

```bash
make run
```

Create DB container

```bash
make docker-run
```

Shutdown DB container

```bash
make docker-down
```

live reload the application

```bash
make watch
```

run the test suite

```bash
make test
```

clean up binary from the last build

```bash
make clean
```
