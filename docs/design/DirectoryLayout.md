# Project Structure

To understand easily what file belongs to where,
you can go trough the directory layout guideline,
which helps explains what directory responsible to contain what.

```
..
├── bin                        -> executables for toggler local development 
├── cmd                        -> service compilable cli commands
│   └── toggler                -> http service command
│       └── main.go
├── docker-compose.yml         -> testing/local dependencies
├── docs
├── extintf                    -> external interface related implementations
│   ├── httpintf
│   │   ├── httpapi            -> httpapi implementation
│   │   │   └── ServeMux.go    -> ServeMux integrate all the httpapi handler
│   │   ├── httputils
│   │   ├── ServeMux.go
│   │   └── webgui
│   │       ├── assets
│   │       ├── controllers
│   │       ├── ServeMux.go
│   │       └── views
│   └── storages
│       ├── inmemory           -> in memory storage implementation for quick testing purposes
│       └── postgres           -> Postgres storage implementation
│           └── assets
│               └── migrations -> pg db schema migration files
├── Procfile                   -> heroku integration
├── services                   -> service's domain entities, rules and usecase 
│   ├── rollouts               -> rollout related domain implementations
│   │   └── specs              -> rollouts service dependency's specifications 
│   └── security               -> security related domain rules
│       └── specs              -> security service dependency's specifications
├── testing                    -> internal testing pkg
├── tools.go                   -> tooling version control with go mod
└── usecases                   -> all the service integrated into one pkg for refactoring purposes
    └── specs
        └── StorageSpec.go


40 directories, 144 files
```

### extintf
This directory owns the implementation regarding external resources and interfaces.
In case of a external interface change, this directory expected to be affected only.

#### httpintf
This directory responsible for owning HTTP protocol based external interfaces,
such as API, WebGUI or monitoring integration.

##### httpapi
httpapi implements the API that allows clients to interact with

#### storages
Storages pkg act as main entry point for owning storage implementations.
It has a factory method that can return the proper implementation based on the connection string.

##### postgres
postgres fulfills the requirements made by the shared specifications.
Act as a db implementation to store/retrieve data.

### services
This directory collect the business entities and business use-cases.

In every service pkg, there will be a `specs` sub pkg that defines the behavior that is expected 
from the external resources the service use such as storage,
that needs to be fulfilled in order to work well together with this service domain implementations.
[You can read more about the reason here](https://en.wikipedia.org/wiki/Design_by_contract).

#### rollouts
rollouts collects domain logic that is purely related towards the process of defining feature flags,
enrol manually pilots, define the enrollment percentage for the rest of the user-base and checking flag status.

#### security
security collects domain logic that aims to provide security aspects for working with toggler.

### usecases
usecases package is the integration hub for every existing and in the future to be created services pkg.
It allow easy integration both 
for external interfaces (eg.: http input interface) 
and external reasources (eg.: storage).
