---
icon: lucide/circuit-board
---

# List of mutators

## Arithmetic mutators

### arithmetic/base

| Name           | Original | Mutated |
|:---------------|:---------|:--------|
| Plus           | +        | -       |
| Minus          | -        | +       |
| Multiplication | *        | /       |
| Division       | /        | *       |
| Modulus        | %        | *       |

### arithmetic/bitwise

| Name          | Original | Mutated |
|:--------------|:---------|:--------|
| BitwiseAnd    | &        | &#124;  |
| BitwiseOr     | &#124;   | &       |
| BitwiseXor    | ^        | &       |
| BitwiseAndNot | &^       | &       |
| ShiftRight    | \>>      | <<      |
| ShiftLeft     | <<       | \>>     |

### arithmetic/assign_invert

| Name      | Original | Mutated |
|:----------|:---------|:--------|
| AddAssign | +=       | -=      |
| SubAssign | -=       | +=      |
| MulAssign | *=       | /=      |
| QuoAssign | /=       | *=      |
| RemAssign | %=       | *=      |

### arithmetic/assignment

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

## Loop mutators

### loop/break

| Name     | Original | Mutated  |
|:---------|:---------|:---------|
| Break    | break    | continue |
| Continue | continue | break    |

### loop/condition

| Name                   | Original | Mutated |
|:-----------------------|:---------|:--------|
| for k < 100            | k < 100  | 1 < 1   |
| for i := 0; i < 5; i++ | i < 5    | 1 < 1   |

### loop/range_break

It is a loop/condition-like mutator in its purpose: removing iterations from code.  
However, the implementation is slightly different. The mutator adds a break to the beginning of each range loop.

| Name               | Original Body | Mutated Body |
|:-------------------|:--------------|:-------------|
| for i,v := range x | without break | with break   |

## Numbers mutators

### numbers/incrementer

| Name             | Original | Mutated |
|:-----------------|:---------|:--------|
| IncrementInteger | 100      | 101     |
| IncrementFloat   | 10.1     | 11.1    |

### numbers/decrementer

| Name             | Original | Mutated |
|:-----------------|:---------|:--------|
| DecrementInteger | 100      | 99      |
| DecrementFloat   | 10.1     | 9.1     |

## Conditional mutators

### conditional/negated

| Name                            | Original | Mutated |
|:--------------------------------|:---------|:--------|
| GreaterThanNegotiation          | \>       | <=      |
| LessThanNegotiation             | <        | \>=     |
| GreaterThanOrEqualToNegotiation | \>=      | <       |
| LessThanOrEqualToNegotiation    | <=       | \>      |
| Equal                           | ==       | !=      |
| NotEqual                        | !=       | ==      |

If you are looking for simple comparison mutators - see [expression-mutators](#expression-mutators)

## Branch mutators

### branch/case

Empties case bodies.

### branch/if

Empties branches of `if` and `else if` statements.

### branch/else

Empties branches of `else` statements.

## Expression mutators

### expression/comparison

Searches for comparison operators, such as `>` and `<=`, and replaces them with similar operators to catch off-by-one errors, e.g. `>` is replaced by `>=`.

| Name                 | Original | Mutated |
|:---------------------|:---------|:--------|
| GreaterThan          | \>       | \>=     |
| LessThan             | <        | <=      |
| GreaterThanOrEqualTo | \>=      | \>      |
| LessThanOrEqualTo    | <=       | <       |

### expression/remove

Searches for `&&` and <code>\|\|</code> operators and makes each term of the operator irrelevant by using `true` or `false` as replacements.

## Statement mutators

### statement/remove
Removes assignment, increment, decrement and expression statements.

## How do I write my own mutators? { #write-mutation-exec-commands }

Each mutator must implement the `Mutator` interface of the [github.com/leonidboykov/go-mutesting/mutator](https://pkg.go.dev/github.com/leonidboykov/go-mutesting/mutator#Mutator) package. The methods of the interface are described in detail in the source code documentation.

Additionally each mutator has to be registered with the `Register` function of the [github.com/leonidboykov/go-mutesting/mutator](https://pkg.go.dev/github.com/leonidboykov/go-mutesting/mutator#Mutator) package to make it usable by the binary.

Examples for mutators can be found in the [github.com/leonidboykov/go-mutesting/mutator](https://pkg.go.dev/github.com/leonidboykov/go-mutesting/mutator) package and its sub-packages.
