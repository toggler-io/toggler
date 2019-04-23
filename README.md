# Feature Flags Service

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

## Features 

- Rollout management
    - [ ] rolling out between users based on percentage
    - [ ] rolling out a flag globally for everyone
    - [ ] event and audit log about every rollout
    - [ ] API callback based decision logic enrollment to support custom needs like:  
      * creating control groups
      * A / B Testing
      * other to your company/team related rollout logic

- Feature Status check
    - [X] Is enabled for a given User
    - [X] Is rolled out globally
    
- [ ] HTTP API for Public Web access
- [ ] tooling with cli client to manage feature rollout

- Multiple storage backend support
    - [ ] [Redis](https://github.com/antirez/redis)
    - [ ] [BoltDB](https://github.com/boltdb/bolt)
    - [ ] [Postgres](https://github.com/postgres/postgres)
    
## [Backlog](https://github.com/adamluzsi/FeatureFlags/projects)

I use Github projects for backlog tracking,
and idea brainstorming.

Feel free to open an issue if you see anything
