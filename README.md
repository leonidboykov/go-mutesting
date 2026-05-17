# go-mutesting
[![Go Reference](https://pkg.go.dev/badge/github.com/leonidboykov/go-mutesting.svg)](https://pkg.go.dev/github.com/leonidboykov/go-mutesting)
[![Go package](https://github.com/leonidboykov/go-mutesting/actions/workflows/go-test.yml/badge.svg)](https://github.com/leonidboykov/go-mutesting/actions/workflows/go-test.yml)
[![codecov](https://codecov.io/github/leonidboykov/go-mutesting/graph/badge.svg?token=5MIQGV48K9)](https://codecov.io/github/leonidboykov/go-mutesting)

go-mutesting is a framework for performing mutation testing on Go source code. Its main purpose is to find source code,
which is not covered by any tests.

## Quick Example

go-mutesting includes a binary which is go-getable.

```bash
go get -t -v github.com/leonidboykov/go-mutesting/...
```

The following command mutates the go-mutesting project with all available mutators.

``` bash
go-mutesting github.com/leonidboykov/go-mutesting/...
```

The execution of this command prints for every mutation if it was successfully tested or not. If not, the source code
patch is printed out, so the mutation can be investigated. The following shows an example for a patch of a mutation.

``` diff
for _, d := range opts.Mutator.DisableMutators {
	pattern := strings.HasSuffix(d, "*")

-	if (pattern && strings.HasPrefix(name, d[:len(d)-2])) || (!pattern && name == d) {
+	if (pattern && strings.HasPrefix(name, d[:len(d)-2])) || false {
		continue MUTATOR
	}
}
```

The example shows that the right term `(!pattern && name == d)` of the `||` operator is made irrelevant by substituting
it with `false`. Since this change of the source code is not detected by the test suite, meaning the test suite did not
fail, we can mark it as untested code.

For the rest of the readme, check out the documentation at <https://leonidboykov.github.io/go-mutesting>.

## Can I make feature requests and report bugs and problems?

Sure, just submit an [issue via the project tracker](https://github.com/leonidboykov/go-mutesting/issues/new) and we will see what I can do.

## Slop-free software

This software is 100% human-authored. All AI-generated contributions will be rejected.

[![This content is 100% human-authored](https://slop-free.org/logos/slop-free.svg)](https://slop-free.org/)
