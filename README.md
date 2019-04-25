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

### Rollout management

The service allows you to be able to control feature release, trough a combination of options.

### Manual rollout
- [ ] domain logic implemented
- [ ] available trough API 

The basic scenario where you can enroll users to become a pilot of a new feature,
that you want to measure trough they feedback and usage.
This is useful when you have loyal customers, who love to try out new features early,
and give feedback they personal feedback about it.

### Global Release
- [ ] domain logic implemented
- [ ] available trough API

This option allow you to turn on or off a certain feature.
This is used for fully release features.
Some scenario are preventing new features to be able to work with an individual user,
such as batch processing contexts. 
For those, the global release is an option, and in case of malfunctioning,
rolling back the feature becomes easier without release. 

### Rollout By Percentage
- [X] domain logic implemented
- [ ] available trough API

This option is to enroll users based or percentage.
This happens when a feature flag status is being asked from the service.
If the currently calling User is win a Pseudo random lottery,
then the user is enrolled to become a pilot of the new feature.
When a user already failed to become a pilot for a new feature,
the user will be rejected from being able to participate in the feature,
until either the feature rollout percentage is increased,
or a rollout manager enrol the user manually,
or the feature being released to the global audience  
By this, the behavior of the rollout process gives a more consistent feeling


#### Custom Needs like target groups
- [ ] domain logic implemented
- [ ] available trough API

Sometimes it is a requirement, to release a feature for certain target groups first,
for various reasons for the business.
For this it is a common practice to use target groups or "experiments".
This service avoid to collect any sensitive information about the pilots,
therefore the only and best system to know about this information is yours.
To work together easily, you can provide an HTTP API url for the feature flag,
to use that as a domain decision logic for the feature release process.

The API will receive information about:
* feature-flag-name
  * guaranteed
* external-id of the user
  * optional, based on if it was received


### Feature Status check
- [X] Is enabled for a given User
- [X] Is rolled out globally
    
### Storage support
- [ ] [Redis](https://github.com/antirez/redis)
- [ ] [BoltDB](https://github.com/boltdb/bolt)
- [ ] [Postgres](https://github.com/postgres/postgres)

The application do don't depend on a certain storage system,
therefore it is planned to support multiple one.
This would remove the burden on your team to introduce a new db,
which requires new ops experience to maintain.
    
## [Backlog](https://github.com/adamluzsi/FeatureFlags/projects)

I use Github projects for backlog tracking,
and idea brainstorming.

Feel free to open an issue if you see anything
