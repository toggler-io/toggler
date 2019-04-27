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

### Rollout By Percentage
- [X] domain logic implemented
- [ ] available trough API

This option is to enroll users based or percentage.
This happens when a feature flag status is being asked from the service.
If the currently calling User is win a Pseudo random lottery,
then the user is enrolled to become a pilot of the new feature.
The Pseudo random is based on the user's 
external id FNV-1a 64 bit hash code + pseudo random code generation.
This grant random like percentage based feature release distribution on an unknown amount of user.
While you can manually enroll or blacklist users for piloting a feature,
that approach need to persist this information.
This on the other hand only rely on the fact that the external id for the user is uniq on system level. 
The users that lost in the enrollment can still be enrolled when the rollout percentage increase.

#### Global Release on 100 Percentage
- [X] domain logic implemented
- [ ] available trough API

In some cases you don't have such information as individual user ids.
Such scenario can be batch jobs behavior change feature releases.
When the rollout percentage set to be 100%, the feature considered to be globally available,
and the the calls that ask for globally enabled features will be replied with yes.

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
