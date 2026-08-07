# Meta Rules

Meta rules are a way to describe a family or category of CA rules.

## Notation

The meta rule notation is:
```
meta-rule     = predicates

predicates    = predicate ["," predicate]...
predicate     = rule | all-of | any-of | none-of | one-of

all-of        = "AllOf(" predicates ")"
any-of        = "AnyOf(" predicates ")"
none-of       = "NoneOf(" predicates ")"
one-of        = "OneOf(" predicates ")"

rule          = rule-part ["/" rule-part]...
rule-part     = birth
              | survival
              | and-rule
              | or-rule
              | xor-rule
              | permutation
birth         = "B(" conditions ")"
survival      = "S(" conditions ")"
and-rule      = ("A(" | "&(") conditions ")"
or-rule       = ("O(" | "|(") conditions ")"
xor-rule      = ("X(" | "^(") conditions ")"
permutation   = "P(" ranges ")"

conditions    = condition ["," condition]...
condition     = required
              | forbidden
              | excluded-combination
              | cardinality

required             = "+" digits
forbidden            = "!" digits
excluded-combination = "-" digits
cardinality          = "#" integer ".." integer ":" digits

digits               = digit...
digit                = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8"

ranges               = range ["," range]...
range                = integer | integer "-" integer
```

Notes:
* Tokens are case-insensitive.
* Whitespace may appear freely between tokens.
* Within a rule, `B(...)`, `S(...)`, `A(...)`, `O(...)`, `X(...)` and `P(...)` may each appear at most once.
* Digits within a condition are unordered.
* Duplicate digits are not ignored - and will cause errors if repeated.
* Integers can be specified in base 10, binary (e.g. `0b11111111`), octal (e.g. `0o377`) or hexadecimal (e.g. `0xFF`)
* Line comments `//` are allowed - block comments `/*` are not allowed

## Conditions

### `required`

Example:
```
B(+23)
```
Means that both `2` and `3` must be present in B.

### `forbidden`

Example:
```
B(!45)
```
Means that both `4` and `5` must not be present in B.

### `excluded-combination`

Example:
```
S(-37)
```
Means that `3` and `7` must not appear together in S.

### `cardinality`

Example:
```
S(#2..3:0123)
```
Means that, within S, of digits `0`,`1`,`2` and `3` only 2 to 3 of them can be present.

Cardinality can also be used to ensure B or S is empty, example:
```
B(#0..0:012345678)
```

### `A(conditions)`

ANDs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.

Example:
```
A(+2)
```
Means that `2` must be present in both B and S.

Example:
```
A(!4)
```
Means that `4` may be present in B or S - but not both.

### `O(conditions)`

ORs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.

Example:
```
O(+3)
```
Means that `3` must be present in either B or S.

### `X(conditions)`

XORs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.

Example:
```
X(!4)
```
Means that `4` is either present in both B and S - or neither

### `P(ranges)`

Example:
```
P(512-1024,1234,5678)
```
Means that only permutations 512 to 1024 (inclusive); 1234 and 5678 permutations are permitted.

## Example

```
AllOf(
    B(!2345) / S(+47,!3),
    AnyOf(
        B(+0,!1) / S(!012,-568),
        B(+0,!16) / S(+568,!012),
        B(+07,!1) / S(+568,!012),
        B(+17,!06) / S(+5,-12),
        B(+017,!6) / S(+5,-12,-68),
        B(+1,!067) / S(+5,!0,-12),
        B(+1,!067) / S(+5,!1,-12),
        B(+1,!067) / S(+56,-12),
        B(+01,!678) / S(+5,!0,-12,-68),
        B(+01,!678) / S(+5,!1,-12,-68),
        B(+018,!67) / S(+5,!0,-12,-68),
        B(+018,!67) / S(+5,!1,-12,-68),
        B(+018,!67) / S(+56,!8,-12),
        B(+018,!67) / S(+58,!6,-12)
    )
)
```

Is a meta-rule that describes one family of Life-like rules (222) that produce
worlds dominated by stable thick plus patterns...

```
#C Copied from GoGoL
x = 6, y = 6, rule = B0/S467
6b$2b2o$b4o$b4o$2b2o$6b!
```