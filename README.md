# pkg

![Coverage](https://img.shields.io/badge/Coverage-50.9%25-yellow)
[![Go Report Card](https://goreportcard.com/badge/github.com/danlock/pkg)](https://goreportcard.com/report/github.com/danlock/pkg)
[![Go Reference](https://pkg.go.dev/badge/github.com/danlock/pkg.svg)](https://pkg.go.dev/github.com/danlock/pkg)

A generic template I use when making new go projects,
and a collection of small, helpful packages that I don't want to rewrite anymore.

## errors
Prefix the calling functions name to errors for simpler, smaller traces.
This package tries to split the difference between github.com/pkg/errors and Go stdlib errors, with first class support for log/slog.


A full stack trace makes it harder to parse errors at a glance, especially if they are wrapped multiple times.
But relying only on the stdlib's fmt.Errorf will have you grepping through the codebase instead, which is not ideal.


This package encourages you to Wrap errors as they bubble up through the call stack. Wrapping multiple times won't punish you with stuttered stack traces.
They will only improve the error with more function names.

And the resulting error will be self-explanatory... Or more realistically as explanatory your package and function names.
https://go.dev/doc/effective_go#names is still great advice.

The error will also include the file:line info from the very first error in the chain.

WrapAttr enables structured errors, allowing metadata to be easily included in logs.
log/slog support is built in. Slogging an error will include the file:line info from the start of an error chain and any metadata from WrapAttr calls.

errors/attr_test.go includes example output.

## ptr

ptr.To and ptr.From have been recreated in every Go codebase dealing with openapi or protobuf code generation, even before generics.
They should probably be in the stdlib.
