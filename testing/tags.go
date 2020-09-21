package testing

const (
	// TagBlackBox means that the test currently using mechanism that makes the system act like black box.
	// Such case is when the system is used E2E through an external interface such as HTTP over TCP.
	//
	// When black box tag is provided it is expected that components in the tests should behave as real as possible.
	// For example there shouldn't be transaction bounded with the context, or similar things.
	TagBlackBox = `black-box`
)
