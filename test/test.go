// Package test provides helper functions for testing.
package test

import "testing"

func splitMsgs(t testing.TB, msgs ...any) (string, []any) {
	if len(msgs) == 0 {
		return "", nil
	}

	msg, ok := msgs[0].(string)
	if !ok {
		t.Helper()
		t.Fatalf("first msg must be a string instead of a %T", msgs[0])
	}

	if len(msgs) == 1 {
		return msg, nil
	}

	return msg, msgs[1:]
}

// FailOnError calls t.Errorf if err is not nil with the error and any additional args passed in.
func FailOnError(t testing.TB, err error, msgs ...any) {
	msg, args := splitMsgs(t, msgs...)
	if err != nil {
		t.Helper()
		t.Errorf(msg+`|err="%+v"`, append(args, err)...)
	}
}

// AbortOnError calls t.Fatalf if err is not nil with the error and any additional args passed in.
func AbortOnError(t testing.TB, err error, msgs ...any) {
	msg, args := splitMsgs(t, msgs...)
	if err != nil {
		t.Helper()
		t.Fatalf(msg+`|err="%+v"`, append(args, err)...)
	}
}

// AbortOnErrorVal calls t.Fatalf if err is not nil with the error and any additional args passed in.
func AbortOnErrorVal[T any](val T, err error) func(t testing.TB, msgs ...any) T {
	return func(t testing.TB, msgs ...any) T {
		if err != nil {
			t.Helper()
			msg, args := splitMsgs(t, msgs...)
			t.Fatalf(msg+`|err="%+v"`, append(args, err)...)
		}
		return val
	}
}

// Equality calls t.Errorf if wanted != expected with any additional args passed in.
func Equality[T comparable](t testing.TB, wanted, actual T, msgs ...any) {
	msg, args := splitMsgs(t, msgs...)
	if wanted != actual {
		t.Helper()
		t.Errorf(msg+`|wanted="%v",actual="%v"`, append(args, wanted, actual)...)
	}
}

// EqualityOrAbort calls t.Fatalf if wanted != expected with any additional args passed in.
func EqualityOrAbort[T comparable](t testing.TB, wanted, actual T, msgs ...any) {
	msg, args := splitMsgs(t, msgs...)
	if wanted != actual {
		t.Helper()
		t.Fatalf(msg+`|wanted="%v",actual="%v"`, append(args, wanted, actual)...)
	}
}

// Truth calls t.Errorf if actual != true with any additional args passed in.
func Truth(t testing.TB, actual bool, msgs ...any) {
	msg, args := splitMsgs(t, msgs...)
	if !actual {
		t.Helper()
		t.Errorf(msg, args...)
	}
}

// TruthOrAbort calls t.Fatalf if actual != true with any additional args passed in.
func TruthOrAbort(t testing.TB, actual bool, msgs ...any) {
	msg, args := splitMsgs(t, msgs...)
	if !actual {
		t.Helper()
		t.Fatalf(msg, args...)
	}
}
