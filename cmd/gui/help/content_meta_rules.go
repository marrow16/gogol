package help

var contentMetaRules = content{
	"The ", bold("Meta Rules"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
	"Select an existing ", "meta rule from the dropdown, or enter a new name to create a new meta rule.\n\n",
	"When a meta rule is selected, choose the ", button("Edit"), " radio button to view or edit the meta rule definition. Changes to the definition are reflected in the number of ", bold("Matched rules"), " shown at the bottom of the popout.\n",
	"If there is an error in the meta rule definition - it is displayed below the edit area.\n\n",
	"Select the ", button("Matching Rules"), " radio button to display the rules matched by the current meta rule. Clicking a rule in the list sets it as the current rule.\n\n",
	"Use the ", button("Delete"), " button to remove the selected meta rule.\n\n",
	"For details of the meta rule language and syntax, see the ", MetaRulesRef.link(), ".",
}

var contentMetaRulesRef = content{
	"The following provides the normative reference for Meta Rules syntax:",
	indent{indent: 20, content: content{
		codeBlock{code: `meta-rule     = predicates

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
range                = integer | integer "-" integer`},
	}},
	indent{indent: 20, content: content{
		"Notes:",
		hanging{gap: 10, indent: 10,
			prefix:  "•",
			content: content{"Tokens are case-insensitive."},
		},
		hanging{gap: 10, indent: 10,
			prefix:  "•",
			content: content{"Whitespace may appear freely between tokens."},
		},
		hanging{gap: 10, indent: 10,
			prefix:  "•",
			content: content{"Within a rule, ", code("B(...)"), ", ", code("S(...)"), ", ", code("A(...)"), ", ", code("O(...)"), ", ", code("X(...)"), " and ", code("P(...)"), " may each appear at most once."},
		},
		hanging{gap: 10, indent: 10,
			prefix:  "•",
			content: content{"Digits within a condition are unordered."},
		},
		hanging{gap: 10, indent: 10,
			prefix:  "•",
			content: content{"Duplicate digits are not ignored - and will cause errors if repeated."},
		},
		hanging{gap: 10, indent: 10,
			prefix:  "•",
			content: content{"Integers can be specified in base 10, binary (e.g. ", code("0b11111111"), ") octal (e.g. ", code("0o377"), ") or hexadecimal (e.g. ", code("0xFF"), ")"},
		},
		hanging{gap: 10, indent: 10,
			prefix:  "•",
			content: content{"Line comments ", code(`//`), " are allowed - block comments ", code(`/*`), " are not allowed"},
		},
		separator{},
		h4("Conditions", 8),
		header{level: 5, text: "required", mono: true},
		indent{indent: 20, content: content{
			"Example:",
			codeBlock{code: `B(+23)`},
			"Means that both ", code(`2`), " and ", code(`3`), " must be present in ", bold("B"), ".",
		}},
		header{level: 5, text: "forbidden", mono: true, spaceBefore: 8},
		indent{indent: 20, content: content{
			"Example:",
			codeBlock{code: `B(!45)`},
			"Means that both ", code(`4`), " and ", code(`5`), " must not be present in ", bold("B"), ".",
		}},
		header{level: 5, text: "excluded-combination", mono: true, spaceBefore: 8},
		indent{indent: 20, content: content{
			"Example:",
			codeBlock{code: `S(-37)`},
			"Means that ", code(`3`), " and ", code(`7`), " must not appear together in ", bold("S"), ".",
		}},
		header{level: 5, text: "cardinality", mono: true, spaceBefore: 8},
		indent{indent: 20, content: content{
			"Example:",
			codeBlock{code: `S(#2..3:0123)`},
			"Means that, within ", bold("S"), ", of digits ", code(`0`), " ", code(`1`), ",", code(`2`), " and ", code(`3`), " only 2 to 3 of them can be present.\n",
			"Cardinality can also be used to ensure that ", bold("B"), " or ", bold("S"), " are empty, example:",
			codeBlock{code: `B(#0..0:012345678)`},
		}},
		header{level: 5, text: "A(conditions)", mono: true, spaceBefore: 8},
		indent{indent: 20, content: content{
			"ANDs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.\n",
			"Example:",
			codeBlock{code: `A(+2)`},
			"Means that ", code(`2`), " must be present in both ", bold("B"), " and ", bold("S"), ".\n",
			"Example:",
			codeBlock{code: `A(!4)`},
			"Means that ", code(`4`), " may be present in ", bold("B"), " or ", bold("S"), " - but not both.",
		}},
		header{level: 5, text: "O(conditions)", mono: true, spaceBefore: 8},
		indent{indent: 20, content: content{
			"ORs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.\n",
			"Example:",
			codeBlock{code: `O(+3)`},
			"Means that ", code(`3`), " must be present in either ", bold("B"), " or ", bold("S"), ".",
		}},
		header{level: 5, text: "X(conditions)", mono: true, spaceBefore: 8},
		indent{indent: 20, content: content{
			"XORs the high-order 9 bits (birth) with the low order 9 bits (survives) and evaluates conditions on the resultant.\n",
			"Example:",
			codeBlock{code: `X(!4)`},
			"Means that ", code(`4`), " is either present in both ", bold("B"), " and ", bold("S"), " - or neither.",
		}},
		header{level: 5, text: "P(ranges)", mono: true, spaceBefore: 8},
		indent{indent: 20, content: content{
			"Example:",
			codeBlock{code: `P(512-1024,1234,5678)`},
			"Means that only permutations 512 to 1024 (inclusive); 1234 and 5678 are permitted.",
		}},
		h4("Example", 16),
		codeBlock{code: `AllOf(
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
)`}, "Is a meta-rule that describes one family of Life-like rules (222) that produce worlds dominated by stable thick plus patterns...",
		Image{name: "plus-pattern.png"},
		codeBlock{code: `#C Copied from GoGoL
x = 6, y = 6, rule = B0/S467
6b$2b2o$b4o$b4o$2b2o$6b!`},
	}},
}
