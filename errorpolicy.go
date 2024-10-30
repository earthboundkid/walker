package walker

import (
	"errors"
	"io/fs"
)

// ErrorPolicy is a function that returns
// whether to continue (true) or halt (false) on error
// by examining the error and the current Entry.
type ErrorPolicy func(err error, e Entry) bool

// OnErrorIgnore is an ErrorPolicy that always continues regardless of errors.
var OnErrorIgnore ErrorPolicy = onErrorIgnore

func onErrorIgnore(error, Entry) bool { return true }

// OnErrorHalt is an ErrorPolicy that halts on any error.
var OnErrorHalt ErrorPolicy = onErrorHalt

func onErrorHalt(error, Entry) bool { return false }

// OnErrorPanic is an ErrorPolicy that panics on any error.
// Not recommended for production.
var OnErrorPanic ErrorPolicy = onErrorPanic

func onErrorPanic(err error, _ Entry) bool { panic(err) }

// OnErrorCollect returns an ErrorPolicy that
// collects errors into the provided slice
// while continuing.
func OnErrorCollect(errs *[]error) ErrorPolicy {
	return func(err error, e Entry) bool {
		*errs = append(*errs, err)
		return true
	}
}

// OnErrPermissionIgnore is an ErrorPolicy
// that continues if an error is fs.ErrPermission;
// otherwise it halts on error.
var OnErrPermissionIgnore ErrorPolicy = func(err error, e Entry) bool {
	return errors.Is(err, fs.ErrPermission)
}
