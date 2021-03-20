# Local Development Setup

The local development requires 3 tooling to provision the project fully.

* [golang v1.11+](https://golang.org/) (required)
* [docker-compose](https://docs.docker.com/compose/gettingstarted/) (required)
* [direnv](https://direnv.net/) (optional)

## To ensure local development environment variables

First of all, make sure that the `.envrc` file content is sourced into your shell.
The file contains environment variables for local development.

> If you have `direnv` installed in your shell, you can skip this step.

```bash
. .envrc
```

## To ensure project external resource dependencies

```bash
docker-compose up -d
```

## To provision tooling for the project (go modules)

Then to install tooling execute the below mentioned next command.
This will install all the tooling in your `GOPATH/bin` directory that the depends on.

The `.envrc` append your `GOPATH/bin` to the `PATH` variable.

> to vendor tooling the project relies on go modules

```bash
go generate tools.go
```

## To generate assets and mocks

```bash
go generate ./...
```

## to execute test suite

```bash
go test ./...
```

## To execute benchmark suite

```bash
go test -bench . ./...
```

## To run the service locally with go compiler

```bash
export DATABASE_URL=${TEST_DATABASE_URL_POSTGRES}
go run cmd/toggler/main.go create-token "token-owner-name"
go run cmd/toggler/main.go http-server
```
