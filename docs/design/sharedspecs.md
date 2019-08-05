# Why the so many Shared Specs ?

Since this project initially was my hobby project to solve an issue at my company,
I had to be aware of the limitations.
Such is the fact that I have to work on this project mainly alone,
before it will have any meaningful impact to anyone,
that it would worth someone else time to contribute to this project.

I favor greatly [XP](https://en.wikipedia.org/wiki/Extreme_programming) practices, especially [Pair Programming](https://en.wikipedia.org/wiki/Pair_programming).
Since in the beginning of the project, it was sure that I will be alone, 
it was unrealistic that the traditional Pair Programming can be used in the project,
so I decided that I combine a few ideology into a convention that can help me solve this problem.

* [Pair Programming](https://en.wikipedia.org/wiki/Pair_programming)
* [Design By Contract](https://en.wikipedia.org/wiki/Design_by_contract)
* [Behavior Driven Development](https://en.wikipedia.org/wiki/Behavior-driven_development)

## Components

### Behavior Driven Development
BDD allows me to create specifications,
that allows the system to be easily refactored at implementation level,
while making sure that the expected system behavior remains constant.
Certain behaviors expected to be implemented for dependencies 
that being used in use-case/interactors.
This is achieved by small edge case specification collections,
where implementation details are forbidden to be used.

### Pair Programming
From pair programming, I decided to use the mental model thinking part.
There is a Ping-Pong like game where you have to switch between navigator and driver.
I used this approach to switch between system behavior test design writing,
and implementation creation.
In my perspective, system behavior tests has higher priority,
and therefore need to be created with more care,
and the implementation can only fulfil the behavior expectations.
Anything made without a test is considered as a POC/draft in the project.

### Design By Contract
Some idea used slightly similar like in Eiffel,
but I put the biggest focus on the preconditions part in this subject.
Certain preconditions have to be meet for certain dependencies,
in order to use it.
This is achieved by Interface compositions. 

## The Combination
By combining DbC, PP and BDD to my problem which is the lack of time, 
money and human resources,
I came up with the solution of shared specs.

By creating shared specifications, I can force myself to use PP Navigator mental model,
design the system trough behavior, define the expectations how it should behave,
and write down in a specification that can be used later on.
Many small behavior specifications than can be combined later on together
for a higher more sophisticated expectations.
I apply this for every dependency, my interactors have.
Trough this, the interactors then can safely assume certain behaviors,
and use the interface contract to interact with them.

In general, creating an acceptable behavior contract (shared spec) requires much more energy 
than doing the actual implementation.
Therefore I spend the time I have in the weekends with navigator roles,
and then in the nights, I can focus on the implementation more freely.

## Funny side effect

This sometimes result in swearing for me during the night, 
when I try to argue myself why the shared spec is probably wrong,
until I realize trough rubber duck debugging, 
that the focus on the implementation detail blinds me from seeing the high level concepts.
Exactly what PP can help in big companies easily.

There is nothing more annoying, 
that the feeling you get when you past self is right
after a 15m one sided conversation.
