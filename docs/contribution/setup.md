# Local Development Setup

The local development requires 3 tooling to provision the project fully.

* [golang v1.11+](https://golang.org/) (required)
* [docker-compose](https://docs.docker.com/compose/gettingstarted/) (required)
* [direnv](https://direnv.net/) (optional)

First of all, make sure that the `.envrc` file content is sourced into your shell.
The file contains environment variables for local development.

> If you have `direnv` installed in your shell, you can skip this step.

```bash
. .envrc
```

Then provision the project's external resource dependencies with `docker-compose`

```bash
docker-compose up -d
``` 

Then to install tooling execute the below mentioned next command.
This will install all the tooling in your GOPATH/bin directory that the depends on.

> to vendor tooling the project relies on go modules

```bash
go generate tools.go
```

And then you can generate assets with the `generate` `go` subcmd

```bash
go generate ./...
```

And Finally, you can verify your current version by running the testing suite.

```bash
go test ./...
```

Optionally, you can execute the benchmark suite as well.

```bash
go test -bench . ./...
```
