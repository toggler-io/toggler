# The Dark Launch

Dark launching is the process of releasing production-ready features to a subset of your users prior to a full release.
This enables you to decouple deployment from release, get real user feedback, test for bugs, and assess infrastructure performance.

## Scenario A – “I want to release an experimental One Click Checkout to my users to see if it increases sales.”

To dark launch this feature, you would enable it for 1% of your users, then to 5%, and 10%.. and assess performance along the way.
If you see that the checkout is increasing sales, then you can gradually increase the rollout percentage.
However, if you see that sales are worse, customers are confused, or that it degrades your app’s performance,
then you can simply roll back the feature for further evaluation and refinement.

## Scenario B – “I want to test new application infrastructure without switching all of my traffic.”

Before switching all of your traffic to a new system architecture,
you can dark launch new infrastructure by routing traffic via a toggle/flag specifically geared for configuration management.
For instance, suppose you want to stop maintaining your own queuing system and switch to a managed service.
You might create a flag that sends some jobs to the new managed service,
while still sending most to the old queuing system (and you have workers set up to listen on both).
Then, you can gradually transition to an all-managed service as you monitor performance and other metrics.
For another example, you can ramp traffic up and down during the daytime hours (when your DevOps team is awake),
rather than perform a hard cutover during the midnight hours.

## Scenario C – “I want to specifically release a feature for beta testing.”

Also referred to as a canary launch, this type of dark launch specifically target users in your ‘beta’ group or target users via an ID (like email or UID).
This will enable the new feature for these particular users, while all of your other users receive the current feature set.
At any time, you can add or remove beta users, get feedback, and assess system performance.

## Scenario D – “I want to release a new feature to my VIP users first.”

Dark launching via feature flags and toggles will allow you to roll out a new feature to your VIP users before doing a widespread release.
You can also allow your users to opt-in and out of dark launched features, similar to a self-selected beta.
