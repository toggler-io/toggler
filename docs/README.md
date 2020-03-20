# toggler

`toggler` is a Self hosted Feature Toggle Service.
There is a SaaS solutions as well deployed at [toggler.io](https://toggler.io),
[you can read more about it under this link](/docs/toggler-io/README.md).

The service designed to be hosted on public web.
The service expects that public web request will be received from all kind of sources.
Such case is the combined usage from SPA, lambda service and traditional backend services.

It is goal to provide a stable, reliable and free rollout management tooling for teams.
By using release flags you can decouple the feature release from the deployment or config change process,
and also make it simple to keep feature states in sync for all your users.

The project aims only to be just barely enough for the minimal requirement
that needed to do centralised feature release management.

Other than percentage based feature enrollment for piloting,
every custom decision logic is expected to be implemented by your company trough an HTTP API.

## What is a Feature Toggler service ?

## Is this Service for your team/company ?

Answer the following questions, in order to determine,
if this project is needed for your team or not.

Can my team…

* apply [Dark Launching](/docs/release/DarkLaunch.md) practices ?
* deploy frequently the codebase independently from feature release ?
* confidently deploy to production after the automated tests are passed ?
* perform deployment during normal business hours with negligible downtime?
* complete its work without needing fine-grained communication and coordination with people outside of the team?
* deploy and release its product or service on demand, independently of other services the product or service depend upon?

If your answer yes to most of them,
then you can stop here, because adding this service to your stack would not solve too much.
else, please continue...

## Why toggler, why not ${PRODUCT_NAME} ?

`toggler` primary goal is to provide a vendor-lock free solution to a wide range of audience.
The API's follow simple conventions, making transition easy to almost any other service
in case the service is not enough for the company/team needs.

The secondary goal is derived from the primary, which is to own the data.
Regardless how and where you use, the ownership of the data belongs to you.
There is no `free` plan that aims you to lock in by data, or by other means.

If you are satisfied with the service,
then the `toggler` project did well.

## State of the code

### Architecture

The `toggler` project core design is based on `the clean architechture` principles,
and split across architecture layers.

The folder structure also try to represent it trough `screaming architecture` elements.

### Testability & Maintainability

The `toggler` project coverage is made with *behavior driven development* principles,
and as such the tests aim to justify system behavior, not implementation.
You probably heard this already many times, but in this case thing about in a way
where you have `postgres` implementation without a single query or db table assertion. :)
Purely just behavior testes.

Why this is good ? Through this, I can have as many different implementations,
and share the expectations from separate components such as storage implementation,
and any contributor can jump in and contribute to it, even without deep TDD or BDD practices.
Also refactoring the internal implementation of the project components is easier.
[There is an in depth explanation in this section.](/docs/design/sharedspecs.md)

## Scalability

The Service follows the [12 factor app](https://12factor.net/) principles,
and scale out via the process model.
The application don't use external resource dependent implementation,
so as long the external resource you use can be scale out, you will be fine.

If you need to add a new storage implementation,
because you need to use that,
feel free to create a Issue or a PR.
If you decide to implement your own integration with a storage,
the expected behavior requirements/tests/coverage can be located under the `usecases/specs.StorageSpec` object.
For examples, you can check the already existing storage implementations as well.

## [Rollout Features](/docs/release/README.md)

## [Design](/docs/design/README.md)

## Quick Start / Setup

### Local Development

You need to ensure the following dependencies:
- go compiler (must have)
- bash (should have)
    * some script is written in `bash`
- docker-compose (should have)
    * +docker
- [direnv](https://github.com/direnv/direnv) (nice to have)

To provision the project, execute the following command:
```bash
. .envrc
./bin/provision
# or
provision
docker-compose up --detach
```

To execute the tests:
```bash
go test ./...
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

* [Redis](https://github.com/antirez/redis)
* [Postgres](https://github.com/postgres/postgres)
* InMemory (for testing purposes only)

The Storage connection can be configured trough the `DATABASE_URL` environment variable
or by providing the `-database-url` cli option to the executable.

To use one of the implementation, all you have to do is
to provide the connection string in the CLI option or in the environment variable.

example connection strings:
> redis://user:passwd@ec2-111.eu-west-1.compute.amazonaws.com:17379

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
