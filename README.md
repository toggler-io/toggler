# toggler

`toggler` is a Self hosted Feature Toggle Service.

The service designed to be hosted on public web.
The service expects that public web request will be received from all kind of sources.
Such case is the combined usage from SPA, lambda service and traditional backend services.

It is goal to provide a stable, reliable and free rollout management tooling for teams.
By using feature flags you can decouple the feature release from the deployment or config change process,
and also make it simple to keep feature states in sync for all your users.

The project aims only to be just barely enough for the minimal requirement
that needed to do centralised feature release management.

Other than percentage based feature enrollment for piloting,
every custom decision logic is expected to be implemented by your company trough an HTTP API.

## Is this Service for your team/company ?

Answer the following questions, in order to determine,
if this project is needed for your team or not.

Can my team…
* apply [Dark Launching](docs/DarkLaunch.md) practices ?
* deploy frequently the codebase independently from feature release ?
* confidently deploy to production after the automated tests are passed ?
* perform deployment during normal business hours with negligible downtime?
* complete its work without needing fine-grained communication and coordination with people outside of the team?
* deploy and release its product or service on demand, independently of other services the product or service depend upon?

If your answer yes to most of them,
then you can stop here, because adding this service to your stack would not solve too much.
else, please continue...

## Features

### Rollout management

The service allows you to be able to control feature release, trough a combination of options.

### Manual rollout

The basic scenario where you can enroll users to become a pilot of a new feature,
that you want to measure trough they feedback and usage.
This is useful when you have loyal customers, who love to try out new features early,
and give feedback they personal feedback about it.

### Rollout By Percentage

This option is to enroll users based or percentage.
This happens when a feature flag status is being asked from the service.
If the currently calling User is win a Pseudo random lottery,
then the user is enrolled to become a pilot of the new feature.
The Pseudo random lottery allow the system to have deterministic
and reproducible rollout enrollment result for each pilot ID,
while ensuring that the user pool size can be infinitely big
without having any resource hit on the feature flag service.

Also this grant random like percentage based feature release distribution.
The randomness can be controlled by modifying the feature flag rollout random seed.
While you can manually enroll or blacklist users for piloting a feature,
that approach need to persist this information.
This on the other hand only rely on the fact that the external id for the user is uniq on system level.
The users that lost in the enrollment can still be enrolled when the rollout percentage increase.

#### Global Release on 100 Percentage

In some cases you don't have such information as individual user ids.
Such scenario can be batch jobs behavior change feature releases.
When the rollout percentage set to be 100%, the feature considered to be globally available,
and the the calls that ask for globally enabled features will be replied with yes.

#### A/B Testing Experiments

When it is unknown what will be more suitable for the users,
it is a common practice to test two version on a small subset of the userbase,
and monitor the results closely from the users.
If one of the version turns out to be success,
then it can be released for wider audience.

#### Custom Needs like target groups

Sometimes it is a requirement, to release a feature for certain target groups first,
for various reasons for the business.
For this it is a common practice to use target groups or "experiments".
This service avoid to collect any sensitive information about the pilots,
therefore the only and best system to know about this information is yours.
To work together easily, you can provide an HTTP API url for the feature flag,
to use that as a domain decision logic for the feature release process.

The API will receive information about:
* feature-flag-name
  * flag name that was received by the FeatureFlag service
* pilot-id
  * uniq id that was received by the FeatureFlag service

## Quick Start / Setup

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
- [X] [Redis](https://github.com/antirez/redis)
- [x] [Postgres](https://github.com/postgres/postgres)
- [x] InMemory (for testing purposes only)

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

#### Cache
The cache external resource is an optional addition.
By default, the service don't try to be smart, and use no cache at all.

You choose to have a trade off for your storage system to use a traditional database
that can provide your fact data with cost effectiveness, stability and maintainability perspectives.
But then you don't want to sacrifice the service response times, so you can use a cache system to solve this.
The Caching system do automatic cache invalidation with TTL and on Update/Delete storage operations.

Currently only redis is available, but additional solutions in progress.

To setup the application to use the cache, either provide the `-cache-url` cli option
or the `CACHE_URL` environment variable.

To setup the cache TTL, you can use the `-cache-ttl` cli option or the `CACHE_TTL` environment variable.
A cache ttl duration in string format must be a unsigned sequence of
decimal numbers, each with optional fraction and a unit suffix,
such as "300ms", "1.5h" or "2h45m".
Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".

### Deployment
* [heroku](docs/deploy/heroku.md)
* [on-premises](docs/deploy/on-prem.md)
* [Docker](docs/deploy/docker.md)

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

#### Calling the API

example calls to get started easily

##### checking rollout related flags in one request:

```bash
curl -X GET 'http://localhost:8080/api/v1/rollout/config.json?id=pilotExtID&feature=newFeature' 
#> {"states":{"newFeature":false}}

curl -X GET 'http://localhost:8080/api/v1/rollout/config.json' \
    -d '{"id":"x", "features": ["a", "b", "c"]}' 
#> {"states":{"a":false,"b":false,"c":false}}
```

##### checking a flag individually for
```bash
curl -X GET 'http://localhost:8080/api/v1/rollout/is-enabled.json' \
    -d '{"feature":"myfeature", "id":"public-uniq-id-of-four-user"}'
#> {"enrollment":false}
```

##### checking a flag is globally enabled
```bash
curl -X GET 'http://localhost:8080/api/v1/rollout/is-globally-enabled.json?feature=myfeature' \
    -d '{"feature":"myfeature", "id":"public-uniq-id-of-four-user"}'
#> {"enrollment":false}
```

## For Contributors
* [Please read the design section of the project](docs/design)
* [Backlog](https://github.com/adamluzsi/toggler/projects)

Feel free to open an issue if you see anything

## Thank you for reading about this project! :)
 