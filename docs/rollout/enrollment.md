# Rollout Feature Enrollment Interpretation

When a pilot rollout feature enrollment checked,
the system first check if there is any manual configuration for the given pilot
regarding the rollout feature flag enrollment aspect.
If found, then served.

Else it will then checks if there is a custom decision logic involved
with the rollout feature flag.
If found, remote url is called for enrollment clarification.

Else it will then do a pseudo mathematics based deterministic random dice roll between 0 and 100,
salted with a random seed number that is taken from the rollout feature flag configuration,
and then it checks the rollout feature flag release percentage.
If the dice roll result is smaller or equal with the release percentage,
then the pilot is enrolled, otherwise the pilot is allowed for enrollment.

0% release percentage is an exception from this, 
as it defines a global off state for the flag.
