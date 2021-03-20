package contracts

//
//type FixtureFactory struct {
//	testing.FixtureFactory
//}
//
//func (f FixtureFactory) Create(T interface{}) (StructPTR interface{}) {
//	switch T.(type) {
//	case caches.StorageRecord:
//		valueT := fixtures.Random.ElementFromSlice([]interface{}{
//			release.Flag{},
//			release.Rollout{},
//			release.ManualPilot{},
//			deployment.Environment{},
//			security.Token{},
//		})
//
//		sr := &caches.StorageRecord{
//			Key: caches.StorageRecordKey{
//				T:         valueT,
//				Operation: fixtures.Random.String(),
//				Key:       fixtures.Random.String(),
//			},
//			IsList:  false,
//			IsFound: false,
//		}
//
//		if fixtures.Random.Bool() {
//			sr.Value = f.Create(valueT)
//			sr.IsFound = true
//		}
//
//		return sr
//
//	default:
//		return f.FixtureFactory.Create(T)
//	}
//}
