# Project Structure

```
..
├── bin                        -> executables for toggler local development 
├── cmd                        -> service compilable cli commands
│   └── toggler                -> http service command
│       └── main.go
├── docker-compose.yml         -> testing/local dependencies
├── docs
├── external                   
│   ├── interface              -> external interface implementations can be located here.
│   │   └── httpintf           -> http interface of the application
│   │       ├── httpapi        -> httpapi implementation
│   │       └── webgui         -> POC Web GUI of implementation
│   └── resource               -> external resource implementations can be located here.
│       └── storages           -> storage implementations that supply toggler use-case requirements
├── Procfile                   -> heroku integration
├── domains                    -> domain entities, rules, interactors regarding differend domain spaces
│   ├── toggler                -> all the service integrated into one pkg for the toggler project
│   │   └── specs              -> high level specs for the toggler project domain needs from a resource 
│   │       └── Storage.go
│   │ ...
│   ├── rollouts               -> rollout related domain implementations
│   │   └── specs              -> rollouts service dependency's specifications
│   │ ... 
│   └── security               -> security related domain rules
│       └── specs              -> security service dependency's specifications
├── testing                    -> internal testing pkg
└── tools.go                   -> tooling version control with go mod
```
