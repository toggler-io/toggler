# Contribution

This section deals with explaining about the common caveats of contribution.

## [Local development setup guide](/docs/contribution/setup.md)

## What to expect if I want to contribute to the project ?

If you contribute to the project, first before all,
you can expect my gratitude for investing your time into the project!
Really, seriously thank you even considering it!

Second of all, you can expect help in many form.
If something that puzzling regarding the project,
that means I can improve the documentations.

If you are not sure about how to extend,
that means the project architecture design is either not consistent,
or not self explaining, and can be improved.

And lastly, by contributing, you help the vision of a vendor-lock free future in this field!
 
## What are the Requirements ? 

During the design of your code, 
the project expect that you create coverage for any new logic,
that is part of the business entities or the domain logic interactors.

The tests should not define implementation details.
Why is that ? For those who not yet familiar with TDD/BDD,
the project still open, because refactoring is much easier,
if system behavior is the only thing that is tested.
Thus if you modify, improve something in the code base,
and the tests are green, then you just made a production grade code,
that highly will be deployed soon.

Out of laziness I use PascalCase for file names, 
so I can just copy paste the structure names.
If you see value changing this,
please write down what values you found that can improve the code,
and it can be changed then.

## Can you tell me an example about what can be contributed ?

### By Backlog 

First of all, I use the [github projects page](https://github.com/adamluzsi/toggler/projects) as backlog.
I distinct projects across product lines in the project.
Hopefully each of those product line can give more value to the users of this service for they own development.
They are not perfectly split, and I'm open for any constructive criticism!

### By Challenging

If you see something that violates business values, and can be improved,
or you measured something that should be improved, 
because it create bottleneck for the system,
PRs are welcome and we can iterate on the mather

### By Your company Needs

If your company use a certain storage system, 
like [Consul by HashiCorp](https://www.consul.io/) and you want to use it as an attached resource to toggler,
you can use the shared specification that has all the behavioral test edge case expectations from a new implementation.
All you have to do is create a test something like the example below, 
and the from the RED/GREEN/REFACTOR the reds will be already given. 

I highly suggest to use `-test.failfast` option,
because the long list of breaking tests in the beginning can be overwhelming.  

```go
package consul_test

import (
	"github.com/adamluzsi/toggler/extintf/storages/redis"
	testing2 "github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases/specs"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConsul(t *testing.T) {
	r, err := consul.New(getTestConsulConnstr(t))
	require.Nil(t, err)
	
	specs.StorageSpec{
		Subject:        r,
		FixtureFactory: testing2.NewFixtureFactory(),
	}.Test(t)
}

func getTestConsulConnstr(t *testing.T) string {
	value, isSet := os.LookupEnv(`TEST_STORAGE_URL_CONSUL`)

	if !isSet {
		t.Skip(`redis url is not set in "TEST_STORAGE_URL_CONSUL"`)
	}

	return value
}
``` 

## What to expect from a PR review ?

You can expect the following:

* verification for code Maintainability factors
  * product values & vision navigator 
    * if the implementation violates core values of the project, 
      I will humbly ask you to help solve the issue, 
      and try to give as much context about the "why" as I can to help this process
  * rolling release compatibility
    * the feature you add must be compatible with the existing usage of the app, 
      and should not require downtime to deploy it
  * trunk based commits even in feature branch
    * each commit should represent a stable state that can be deployed to production
      * in case you don't have experience with this, we can work it around with squash merge,
        but I would appreciate it if we don't have to rely on that, 
        to keep the history intact for the reason for changes in them 
  * testability
    * no singleton objects
    * re-usability if it has value
    * how self explaining the test
  * observability 
    * rainy-path logging
    * system performance runtime measurement
    * ops health checks
  * maintainability
    * code is self enplaning for a fresh eye
    * code don't have coupling between SRP actors
    * implementation open for extension, but closed for modification
      * like there should be no or the least from magic functions 
        that do different things based on the input
    * testing coverage
        * at least 80% is expected
            * all happy path
            * rainy paths with business values
            * simple rainy paths with infrastructure failure can be ignored
                * like DB is dead, the system is not expected to handle it magically
        * this helps to keep the difficulty of the development rather low

What you should not expect:

* if you used one solution that can be expressed in many different, 
  it is up to you, as long it has coverage
* premature optimization expectations
  * if something can be improved, it must be measured first!
* I create tests for you in case you want add a new feature
  * I'm more than happy to help, but not to the extent of creating coverage for the feature
* If the tests coverage is breaking for the PR, 
  it is expected to be fixed in the implementation, before approval
* receive feedback to fix typos
* receive feedback to fix white spaces
* receive PC compatibility issues (as long it was not intentionally made)
