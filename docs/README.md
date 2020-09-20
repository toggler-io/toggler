# toggler

`toggler` is a Self hosted Feature and Release Management Service.

The service designed to be hosted on the public web.
The service expects that public web request will be received from all kind of sources.
Such case is the combined usage from SPA, lambda service and traditional backend services.

It is goal to provide a stable, reliable and free rollout management tooling for teams.
By using release flags you can decouple the feature release from the deployment or config change process,
also make it simple to keep feature states in sync for all your users.

The project aims only to be just barely enough for the minimal requirement
that needed to do centralised feature release management.

Other than percentage based feature enrollment for piloting,
every custom decision logic is expected to be implemented by your company trough an HTTP API.

- [Why should you or your team choose Toggler?](/docs/why.md)
- [What's the state of the project, how safe to use it for production purposes?](/docs/readiness.md)

## Getting Started

### Installation

#### Local Development

You need to ensure the following dependencies:
- go compiler (must have)
- bash (should have)
    * some script is written in `bash`
- docker-compose/podman-compose (should have)
    * + docker/podman
- [direnv](https://github.com/direnv/direnv) (nice to have)
    * alternatively you can just source the `.envrc` file at the project root directory.

To provision the project, execute the following command:
```sh
. .envrc
./bin/provision
docker-compose up --detach
```

To execute the tests:
```bash
go test ./...
``` 

#### With Container

You can build toggler from the master branch by using the Dockerfile:
```sh
docker build --tag toggler:latest .
```

or grab from hub.docker.com:
```sh
docker pull adamluzsi/toggler:latest
```


### Configuration
The application can be configured trough either CLI option or with environment variables.
It follows the convention that works easily with SaaS platforms or containerization based solutions.

#### Storage
The storage external resource will be used to persist data,
and then using as source of facts.

The toggler doesn't depend on a certain storage system.
It use behavior based specification, and has multiple implementation that fulfil this contract.
This could potentially remove the burden on your team to introduce a new db just for the sake of the project.

You can choose from the following
* [Postgres](https://github.com/postgres/postgres)
* InMemory (for testing purposes only)

The Storage connection can be configured trough the `DATABASE_URL` environment variable
or by providing the `-database-url` cli option to the executable.

To use one of the implementation, all you have to do is
to provide the connection string in the CLI option or in the environment variable.

example connection string:
> postgres://user:passwd@ec2-111.eu-west-1.compute.amazonaws.com:5432/dbname

```bash
export DATABASE_URL="postgres://user:passwd@ec2-111.eu-west-1.compute.amazonaws.com:5432/dbname"
```

#### [Cache](/docs/caches/README.md)

### Deployment
* [heroku](/docs/deploy/heroku.md)
* [AWS Elastic Beanstalk](/docs/deploy/aws/eb/README.md)
* [on-premises](/docs/deploy/on-prem.md)
* [Docker](/docs/deploy/docker.md)

### Usage

#### Security token creation
To gain access to write and update related actions in the system,
you must create a security token that will be used even on the webGUI.

To create a token, execute the following command on the server:
```bash
./toggler -cmd create-token "token-owner-uid"
```

the uniq id of the owner could be a email address for example.
The token will be printed on the STDOUT.
The token cannot be regained if it is not saved after token creation.

#### API Documentation
* [HTTP API documentation](/docs/httpapi/README.md)
* you can find the swagger documentation at the /swagger.json endpoint.
* the webgui also provides swagger-ui out of the box on the /swagger-ui path

## For Contributors
* [High level overview](/docs/contribution/README.md)
* [Local Development Setup](/docs/contribution/setup.md)
* [Backlog](https://github.com/toggler-io/toggler/projects)

Feel free to open an issue if you see anything

## Thank you for reading about this project! :)
