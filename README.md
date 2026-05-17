# go-mutesting
[![Go Reference](https://pkg.go.dev/badge/github.com/leonidboykov/go-mutesting.svg)](https://pkg.go.dev/github.com/leonidboykov/go-mutesting)
[![Go package](https://github.com/leonidboykov/go-mutesting/actions/workflows/go-test.yml/badge.svg)](https://github.com/leonidboykov/go-mutesting/actions/workflows/go-test.yml)
[![codecov](https://codecov.io/github/leonidboykov/go-mutesting/graph/badge.svg?token=5MIQGV48K9)](https://codecov.io/github/leonidboykov/go-mutesting)

go-mutesting is a framework for performing mutation testing on Go source code. Its main purpose is to find source code,
which is not covered by any tests.



## <a name="table-of-content"></a>Table of content

- [What is mutation testing?](#what-is-mutation-testing)
- [How do I use go-mutesting?](#how-do-i-use-go-mutesting)
- [How do I write my own mutation exec commands?](#write-mutation-exec-commands)
- [Which mutators are implemented?](#list-of-mutators)
- [Other mutation testing projects and their flaws](#other-projects)
- [Can I make feature requests and report bugs and problems?](#feature-request)

## <a name="how-do-i-use-go-mutesting"></a>How do I use go-mutesting?

go-mutesting includes a binary which is go-getable.

```bash
go get -t -v github.com/leonidboykov/go-mutesting/...
```

The binary's help can be invoked by executing the binary without arguments or with the `--help` argument.

```bash
go-mutesting --help
```

> **Note**: This README describes only a few of the available arguments. It is therefore advisable to examine the output
> of the `--help` argument.

The targets of the mutation testing can be defined as arguments to the binary. Every target can be either a Go source
file, a directory or a package. Directories and packages can also include the `...` wildcard pattern which will search
recursively for Go source files. Test source files with the suffix `_test` are excluded, since this would interfere with
the testing process most of the time.

The following example gathers all Go files which are defined by the targets and generate mutations with all available
mutators of the binary.

```bash
go-mutesting parse.go example/ github.com/leonidboykov/go-mutesting/mutator/...
```

Every mutation has to be tested using an [exec command](#write-mutation-exec-commands). By default the built-in exec
command is used, which tests a mutation using the following steps:

- Replace the original file with the mutation.
- Execute all tests of the package of the mutated file.
- Report if the mutation was killed.

The execution will print the following output.

> **Note**: This output is from an older version of go-mutesting. Up to date versions of go-mutesting will have different mutations.

```diff
PASS "/tmp/go-mutesting-422402775//home/avito-tech/go/src/github.com/leonidboykov/go-mutesting/example/example.go.0" with checksum b705f4c99e6d572de509609eb0a625be
PASS "/tmp/go-mutesting-422402775//home/avito-tech/go/src/github.com/leonidboykov/go-mutesting/example/example.go.1" with checksum eb54efffc5edfc7eba2b276371b29836
PASS "/tmp/go-mutesting-422402775//home/avito-tech/go/src/github.com/leonidboykov/go-mutesting/example/example.go.2" with checksum 011df9567e5fee9bf75cbe5d5dc1c81f
--- Original
+++ New
@@ -16,7 +16,7 @@
        }

        if n < 0 {
-               n = 0
+
        }

        n++
FAIL "/tmp/go-mutesting-422402775//home/avito-tech/go/src/github.com/leonidboykov/go-mutesting/example/example.go.3" with checksum 82fc14acf7b561598bfce25bf3a162a2
PASS "/tmp/go-mutesting-422402775//home/avito-tech/go/src/github.com/leonidboykov/go-mutesting/example/example.go.4" with checksum 5720f1bf404abea121feb5a50caf672c
PASS "/tmp/go-mutesting-422402775//home/avito-tech/go/src/github.com/leonidboykov/go-mutesting/example/example.go.5" with checksum d6c1b5e25241453128f9f3bf1b9e7741
--- Original
+++ New
@@ -24,7 +24,6 @@
        n += bar()

        bar()
-       bar()

        return n
 }
FAIL "/tmp/go-mutesting-422402775//home/avito-tech/go/src/github.com/leonidboykov/go-mutesting/example/example.go.6" with checksum 5b1ca0cfedd786d9df136a0e042df23a
PASS "/tmp/go-mutesting-422402775//home/avito-tech/go/src/github.com/leonidboykov/go-mutesting/example/example.go.8" with checksum 6928f4458787c7042c8b4505888300a6
The mutation score is 0.750000 (6 passed, 2 failed, 0 skipped, total is 8)
```

The output shows that eight mutations have been found and tested. Six of them passed which means that the test suite
failed for these mutations and the mutations were therefore killed. However, two mutations did not fail the test suite.
Their source code patches are shown in the output which can be used to investigate these mutations.

The summary also shows the **mutation score** which is a metric on how many mutations are killed by the test suite and
therefore states the quality of the test suite. The mutation score is calculated by dividing the number of passed
mutations by the number of total mutations, for the example above this would be 6/8=0.75. A score of 1.0 means that all
mutations have been killed.

### <a name="black-list-false-positives"></a>Blacklist false positives

Mutation testing can generate many false positives since mutation algorithms do not fully understand the given source
code. `early exits` are one common example. They can be implemented as optimizations and will almost always trigger a
false-positive since the unoptimized code path will be used which will lead to the same result. go-mutesting is meant to
be used as an addition to automatic test suites. It is therefore necessary to mark such mutations as false-positives.
This is done with the `--blacklist` argument. The argument defines a file which contains in every line a MD5 checksum of
a mutation. These checksums can then be used to ignore mutations.

> **Note**: The blacklist feature is currently badly implemented as a change in the original source code will change all
> checksums.

The example output of the [How do I use go-mutesting?](#how-do-i-use-go-mutesting) section describes a mutation
`example.go.6` which has the checksum `5b1ca0cfedd786d9df136a0e042df23a`. If we want to mark this mutation as a
false-positive, we simple create a file with the following content.

```
5b1ca0cfedd786d9df136a0e042df23a
```

The blacklist file, which is named `example.blacklist` in this example, can then be used to invoke go-mutesting.

```bash
go-mutesting --blacklist example.blacklist github.com/leonidboykov/go-mutesting/example
```

The execution will print the following output.

> **Note**: This output is from an older version of go-mutesting. Up to date versions of go-mutesting will have different mutations.

```diff
PASS "/tmp/go-mutesting-208240643/example.go.0" with checksum b705f4c99e6d572de509609eb0a625be
PASS "/tmp/go-mutesting-208240643/example.go.1" with checksum eb54efffc5edfc7eba2b276371b29836
PASS "/tmp/go-mutesting-208240643/example.go.2" with checksum 011df9567e5fee9bf75cbe5d5dc1c81f
--- Original
+++ New
@@ -16,7 +16,7 @@
        }

        if n < 0 {
-               n = 0
+
        }

        n++
FAIL "/tmp/go-mutesting-208240643/example.go.3" with checksum 82fc14acf7b561598bfce25bf3a162a2
PASS "/tmp/go-mutesting-208240643/example.go.4" with checksum 5720f1bf404abea121feb5a50caf672c
PASS "/tmp/go-mutesting-208240643/example.go.5" with checksum d6c1b5e25241453128f9f3bf1b9e7741
PASS "/tmp/go-mutesting-208240643/example.go.8" with checksum 6928f4458787c7042c8b4505888300a6
The mutation score is 0.857143 (6 passed, 1 failed, 0 skipped, total is 7)
```

By comparing this output to the original output we can state that we now have 7 mutations instead of 8.

## <a name="write-mutation-exec-commands"></a>How do I write my own mutation exec commands?

A mutation exec command is invoked for every mutation which is necessary to test a mutation. Commands should handle at
least the following phases.

1. **Setup** the source to include the mutation.
2. **Test** the source by invoking the test suite and possible other test functionality.
3. **Cleanup** all changes and remove all temporary assets.
4. **Report** if the mutation was killed.

It is important to note that each invocation should be isolated and therefore stateless. This means that an invocation
must not interfere with other invocations.

A set of environment variables, which define exactly one mutation, is passed on to the command.

| Name            | Description                                                               |
|:----------------|:--------------------------------------------------------------------------|
| MUTATE_CHANGED  | Defines the filename to the mutation of the original file.                |
| MUTATE_DEBUG    | Defines if debugging output should be printed.                            |
| MUTATE_ORIGINAL | Defines the filename to the original file which was mutated.              |
| MUTATE_PACKAGE  | Defines the import path of the origianl file.                             |
| MUTATE_TIMEOUT  | Defines a timeout which should be taken into account by the exec command. |
| MUTATE_VERBOSE  | Defines if verbose output should be printed.                              |
| TEST_RECURSIVE  | Defines if tests should be run recursively.                               |

A command must exit with an appropriate exit code.

| Exit code | Description                                                                                                   |
|:----------|:--------------------------------------------------------------------------------------------------------------|
| 0         | The mutation was killed. Which means that the test led to a failed test after the mutation was applied.       |
| 1         | The mutation is alive. Which means that this could be a flaw in the test suite or even in the implementation. |
| 2         | The mutation was skipped, since there are other problems e.g. compilation errors.                             |
| >2        | The mutation produced an unknown exit code which might be a flaw in the exec command.                         |

Examples for exec commands can be found in the [scripts](/scripts/exec) directory.

## <a name="list-of-mutators"></a>Which mutators are implemented?

### Arithmetic mutators

#### arithmetic/base

| Name           | Original | Mutated |
|:---------------|:---------|:--------|
| Plus           | +        | -       |
| Minus          | -        | +       |
| Multiplication | *        | /       |
| Division       | /        | *       |
| Modulus        | %        | *       |

#### arithmetic/bitwise

| Name          | Original | Mutated |
|:--------------|:---------|:--------|
| BitwiseAnd    | &        | &#124;  |
| BitwiseOr     | &#124;   | &       |
| BitwiseXor    | ^        | &       |
| BitwiseAndNot | &^       | &       |
| ShiftRight    | \>>      | <<      |
| ShiftLeft     | <<       | \>>     |

#### arithmetic/assign_invert

| Name      | Original | Mutated |
|:----------|:---------|:--------|
| AddAssign | +=       | -=      |
| SubAssign | -=       | +=      |
| MulAssign | *=       | /=      |
| QuoAssign | /=       | *=      |
| RemAssign | %=       | *=      |

#### arithmetic/assignment

| Name             | Original | Mutated |
|:-----------------|:---------|:--------|
| AddAssignment    | +=       | =       |
| SubAssignment    | -=       | =       |
| MulAssignment    | *=       | =       |
| QuoAssignment    | /=       | =       |
| RemAssignment    | %=       | =       |
| AndAssignment    | &=       | =       |
| OrAssignment     | &#124;=  | =       |
| XorAssignment    | ^=       | =       |
| SHLAssignment    | <<=      | =       |
| SHRAssignment    | \>>=     | =       |
| AndNotAssignment | &^=      | =       |


### Loop mutators
#### loop/break
| Name     | Original | Mutated  |
|:---------|:---------|:---------|
| Break    | break    | continue |
| Continue | continue | break    |

#### loop/condition
| Name                   | Original | Mutated |
|:-----------------------|:---------|:--------|
| for k < 100            | k < 100  | 1 < 1   |
| for i := 0; i < 5; i++ | i < 5    | 1 < 1   |

#### loop/range_break
It is a loop/condition-like mutator in its purpose: removing iterations from code.  
However, the implementation is slightly different. The mutator adds a break to the beginning of each range loop.

| Name               | Original Body | Mutated Body |
|:-------------------|:--------------|:-------------|
| for i,v := range x | without break | with break   |

### Numbers mutators
#### numbers/incrementer

| Name             | Original | Mutated |
|:-----------------|:---------|:--------|
| IncrementInteger | 100      | 101     |
| IncrementFloat   | 10.1     | 11.1    |

#### numbers/decrementer

| Name             | Original | Mutated |
|:-----------------|:---------|:--------|
| DecrementInteger | 100      | 99      |
| DecrementFloat   | 10.1     | 9.1     |

### Conditional mutators
#### conditional/negated

| Name                            | Original | Mutated |
|:--------------------------------|:---------|:--------|
| GreaterThanNegotiation          | \>       | <=      |
| LessThanNegotiation             | <        | \>=     |
| GreaterThanOrEqualToNegotiation | \>=      | <       |
| LessThanOrEqualToNegotiation    | <=       | \>      |
| Equal                           | ==       | !=      |
| NotEqual                        | !=       | ==      |

If you are looking for simple comparison mutators - see [expression-mutators](#expression-mutators)

### Branch mutators
#### branch/case
Empties case bodies.

#### branch/if
Empties branches of `if` and `else if` statements.

#### branch/else
Empties branches of `else` statements.

### Expression mutators
#### expression/comparison
Searches for comparison operators, such as `>` and `<=`, and replaces them with similar operators to catch off-by-one errors, e.g. `>` is replaced by `>=`.

| Name                 | Original | Mutated |
|:---------------------|:---------|:--------|
| GreaterThan          | \>       | \>=     |
| LessThan             | <        | <=      |
| GreaterThanOrEqualTo | \>=      | \>      |
| LessThanOrEqualTo    | <=       | <       |

#### expression/remove
Searches for `&&` and <code>\|\|</code> operators and makes each term of the operator irrelevant by using `true` or `false` as replacements.

### Statement mutators
#### statement/remove
Removes assignment, increment, decrement and expression statements.

## Config file

There is a configuration file where you can fine-tune mutation testing.  
The config must be written in YAML format.  
If `--config` is presented, the library will use the given config. Otherwise, no default config file will be used.  
The config contains the following parameters:  


| Name                 | Default value | Description                                                                                                                                                        |
|:---------------------|:--------------|:-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| skip_without_test    | true          | Skip files without _test.go tests.                                                                                                                                 |
| skip_with_build_tags | true          | If in _test.go file we have --build tag - then skip it.                                                                                                            |
| json_output          | false         | Make report.json file with a mutation test report.                                                                                                                 |
| silent_mode          | false         | Do not print mutation stats.                                                                                                                                       |
| exclude_dirs         | []string(nil) | Directories for excluding. In fact, there are not directories. These are the prefix for a path when we scan a file system. So this parameter is sensitive for args |

## <a name="write-mutators"></a>How do I write my own mutators?

Each mutator must implement the `Mutator` interface of the [github.com/leonidboykov/go-mutesting/mutator](https://godoc.org/github.com/leonidboykov/go-mutesting/mutator#Mutator) package. The methods of the interface are described in detail in the source code documentation.

Additionally each mutator has to be registered with the `Register` function of the [github.com/leonidboykov/go-mutesting/mutator](https://godoc.org/github.com/leonidboykov/go-mutesting/mutator#Mutator) package to make it usable by the binary.

Examples for mutators can be found in the [github.com/leonidboykov/go-mutesting/mutator](https://godoc.org/github.com/leonidboykov/go-mutesting/mutator) package and its sub-packages.

## <a name="other-projects"></a>Other mutation testing projects and their flaws

go-mutesting is not the first project to implement mutation testing for Go source code. A quick search uncovers the following projects.

- https://github.com/darkhelmet/manbearpig
- https://github.com/kisielk/mutator
- https://github.com/StefanSchroeder/Golang-Mutation-testing

All of them have significant flaws in comparison to go-mutesting:

- Only one type (or even one case) of mutation is implemented.
- Can only be used for one mutator at a time (manbearpig, Golang-Mutation-testing).
- Mutation is done by content which can lead to lots of invalid mutations (Golang-Mutation-testing).
- New mutators are not easily implemented and integrated.
- Can only be used for one package or file at a time.
- Other scenarios as `go test` cannot be applied.
- Do not properly clean up or handle fatal failures.
- No automatic tests to ensure that the algorithms are working at all.
- Uses another language (Golang-Mutation-testing).

## <a name="feature-request"></a>Can I make feature requests and report bugs and problems?

Sure, just submit an [issue via the project tracker](https://github.com/leonidboykov/go-mutesting/issues/new) and we will see what I can do.

[![This content is 100% human-authored](https://slop-free.org/logos/slop-free.svg)](https://slop-free.org/)
