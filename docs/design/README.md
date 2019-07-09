# Design

Toggler is made with the following concerns when it was begin to be designed.

## Long term goals of the design decisions

### it has to be easy to extend the system behavior
The design includes pipeline pattern, null object pattern and some other to make sure,
that toggler code base strictly align to a shared convention, and easy to extend with new behaviors,
without the need to alter existing code base.

My personal opinion is that humans in general bad at programming,
and each code change can alter a system behavior that is already in used by a user in one way or another.

This boils down to the simple rule, that the software design in toggler must be open for extension,
but existing code base behavior is closed for modifications.

### Anything can be used, as long they used by behavior and not by implementation
Toggler design encourage to approach communication between objects trough contracts in the form of shared specification.
This align well with the Dependency inversion principle.

As a nice side effect of this, the components are not coupled together and swappable easily.
This is exploited with the `storage` component in a way, that an inmemory implementation is used for almost all test,
that allow quick and nimble feedback cycles while remove the need to use mocks in generally.
This give a more close to the production environment feeling, and also helps catch mental model related bugs early on,
that mocks would otherwise hide until the E2E testing.

### optimize for development time
By having quick feedbacks and unified coverage,
it allows me to have an easier time to do small changes iteratively.
This is critical to me, because my time is limited,
since I have high priority things ahead of this project,
such as spending quality time with my family.

In order to make the project able to be easy to maintain,
I keep this point always in front of my eye.
That's why manual testing is only allowed for POC level things,
that has high chance to be remade completly.
Also POC is not allowed for used low level business rules and business use-cases.

### The software should be accessible for new contributors
With strict regulations, trying out new things in the code base aimd to be easer,
by having a quick feedback loop for the person who do the experiments.
The code base should scream for its purpose,
a `tree` command should be able to give
a quick peek into a package responsibility and functionality scope.

### Testing by behavior instead of implementation
Toggler design encourage to focuse on behavior specifications.
At first it may sound reasonable, but this decision is applied to the external interfaces as well.
For example even while `toggler` has a `postgres` implementation,
you will find no sql checking and tests that assert trough verifing a table content with sql queries.
This decision already starting to pay off.

### Portability
Toggler is designed in a way, that the application must be portable,
shipped with a single binary.
This is one of the reason why golang was chosen as the language to implement the functionalities.

### Budget
Toggler design also tries to make decisions,
that allow small companies to use it easily on cheap on-prem way.

### Readability
You will definitely see variable and function names that looks way to long,
and I'm fine with it because I optimize for read instead of write.
The code should able to express the goal it tryies to achieve,
without forcing the reader to heavily dive deep into the implementations.
Suggestion are welcome, since this is highly subjective topic.

## [Directory Layout Guide](DirectoryLayout.md)
To understand easily what file belongs to where,
you can go trough the directory layout guideline,
which helps explains what directory responsible to contain what.
