package lseq

// Whether to enable tighter runtime assertions
const debug = false

// A no-op type, mostly compatible with `log.Logger`.
// Unfortunately, calls to this do not get inlined.
type fakeLogger struct{}

func (fakeLogger) Printf(...interface{})  {}
func (fakeLogger) Println(...interface{}) {}
