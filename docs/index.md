---
icon: lucide/book-open
---

# Introduction

## What is mutation testing?

The definition of mutation testing is best quoted from Wikipedia:

!!! quote ""

    Mutation testing (or Mutation analysis or Program mutation) is used to design new software tests and evaluate the
    quality of existing software tests. Mutation testing involves modifying a program in small ways. Each mutated version
    is called a mutant and tests detect and reject mutants by causing the behavior of the original version to differ from
    the mutant. This is called killing the mutant. Test suites are measured by the percentage of mutants that they kill.
    New tests can be designed to kill additional mutants.

    – <cite>[https://en.wikipedia.org/wiki/Mutation_testing](https://en.wikipedia.org/wiki/Mutation_testing)</cite>


!!! quote ""
    Tests can be created to verify the correctness of the implementation of a given software system, but the creation of
    tests still poses the question whether the tests are correct and sufficiently cover the requirements that have
    originated the implementation.
    
    – <cite>[https://en.wikipedia.org/wiki/Mutation_testing](https://en.wikipedia.org/wiki/Mutation_testing)</cite>

Although the definition states that the main purpose of mutation testing is finding implementation cases which are not
covered by tests, other implementation flaws can be found too. Mutation testing can for example uncover dead and
unneeded code.

Mutation testing is also especially interesting for comparing automatically generated test suites with manually written
test suites. This was the original intention of go-mutesting which is used to evaluate the generic fuzzing and
delta-debugging framework [Tavor](https://github.com/zimmski/tavor).


## Quick example

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

The next mutation shows code from the `removeNode` method of a [linked
list](https://github.com/zimmski/container/blob/master/list/linkedlist/linkedlist.go) implementation.

``` diff
	}

	l.first = nil
-	l.last = nil
+
	l.len = 0
}
```

We know that the code originates from a remove method which means that the mutation introduces a leak by ignoring the
removal of a reference. This can be
[tested](https://github.com/zimmski/container/commit/142c3e16a249095b0d63f2b41055d17cf059f045) with
[go-leaks](https://github.com/zimmski/go-leak).
