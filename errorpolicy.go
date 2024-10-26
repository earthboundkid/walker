package walker

import (
	"errors"
	"io/fs"
)

// ErrorPolicy is a function that returns
// whether to continue (true) or halt (false) on error
// by examining the current state of the Walker.
type ErrorPolicy func(w *Walker) bool

// IgnoreErrors is an ErrorPolicy that always continues regardless of errors.
var IgnoreErrors ErrorPolicy = func(*Walker) bool { return true }

// HaltOnError is an ErrorPolicy that halts on any error.
var HaltOnError ErrorPolicy = func(*Walker) bool { return false }

// CollectErrors returns an ErrorPolicy that
// collects errors into the provided slice
// while continuing.
func CollectErrors(errs *[]error) ErrorPolicy {
	return func(w *Walker) bool {
		*errs = append(*errs, w.Err())
		return true
	}
}

// IgnoreErrPermission is an ErrorPolicy
// that continues if an error is fs.ErrPermission;
// otherwise it halts on error.
var IgnoreErrPermission ErrorPolicy = func(w *Walker) bool {
	return errors.Is(w.Err(), fs.ErrPermission)
}
