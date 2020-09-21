package memorystorage_test

import (
	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/frameless/resources/specs"

	"testing"

	"github.com/toggler-io/toggler/external/resource/storages/memorystorage"
)

func ExampleMemory() *memorystorage.Memory {
	return memorystorage.NewMemory()
}

func TestMemory(t *testing.T) {
	subject := ExampleMemory()
	specs.Creator{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Test(t)
	specs.Finder{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Test(t)
	specs.Updater{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Test(t)
	specs.Deleter{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Test(t)
	specs.OnePhaseCommitProtocol{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Test(t)
}

func BenchmarkMemory(b *testing.B) {
	subject := ExampleMemory()
	specs.Creator{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Benchmark(b)
	specs.Finder{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Benchmark(b)
	specs.Updater{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Benchmark(b)
	specs.Deleter{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Benchmark(b)
	specs.OnePhaseCommitProtocol{EntityType: exampleEntity{}, Subject: subject, FixtureFactory: fixtures.FixtureFactory{}}.Benchmark(b)
}

type exampleEntity struct {
	ExtID string `ext:"ID"`
	Data  string
}
