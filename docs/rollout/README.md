# Rollout

## [Enrollment logic](/docs/rollout/enrollment.md)

## Management

The service allows you to be able to control feature release, trough a combination of options.

## Manual rollout

The basic scenario where you can enroll users to become a pilot of a new feature,
that you want to measure trough they feedback and usage.
This is useful when you have loyal customers, who love to try out new features early,
and give feedback they personal feedback about it.

## Rollout By Percentage

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

### Global Release on 100 Percentage

In some cases you don't have such information as individual user ids.
Such scenario can be batch jobs behavior change feature releases.
When the rollout percentage set to be 100%, the feature considered to be globally available,
and the the calls that ask for globally enabled features will be replied with yes.

### A/B Testing Experiments

When it is unknown what will be more suitable for the users,
it is a common practice to test two version on a small subset of the userbase,
and monitor the results closely from the users.
If one of the version turns out to be success,
then it can be released for wider audience.

### Custom Needs like target groups

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
  