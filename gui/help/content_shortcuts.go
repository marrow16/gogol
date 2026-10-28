package help

import (
	"github.com/marrow16/gogol/gui/shortcuts"
)

var contentShortcuts = content{
	"The ", bold("Shortcuts"), " popout is available from the ", MainMenu.link("main menu"), ".\n\n",
	"Shortcuts are named macros that can execute an action or series of actions in GoGoL.\n\n",
	"Select an existing shortcut ", " from the dropdown, or enter a new name to create a new shortcut.\n\n",
	"If the shortcut name matches a keyboard key, it can be invoked by pressing ", keys{altMac}, " plus that key.",
	" For example, a shortcut named ", bold("A"), " can be invoked by pressing ", keys{altMac, "A"},
	" - likewise, a shortcut named ", bold("F1"), " can be invoked with key ", keys{altMac, "F1"}, "\n",
	italic("Note: Shortcut keys can override can override GoGoL assigned keys."), "\n\n",
	"Having selected or created a shortcut, use the ", bold("Actions"), " area to edit the actions associated - each action on a new line. See ", ShortCutsRef, ".",
	"\n\n", "During actions editing, any syntax errors in the actions will be displayed. Syntax errors are only a validation of recognised actions and expected operands for those actions - not whether values are valid at shortcut run.",
	"\nSyntax errors do not prevent the shortcut being saved and/or run.\n\n",
	"Errors encountered during shortcut run are logged in the file: ", code(`[user]/Documents/GoGoL/errors.log`),
}

var contentShortCutsRef = content{
	"The following reference describes all of the actions that can be used within ", ShortCuts.link("shortcuts"),
	table{
		borders: true,
		padding: 4,
		columns: []tableColumn{
			{
				width:  30,
				header: "Action",
			},
			{
				header: "Description",
			},
		},
		rows: []tableRow{
			{{codeCopyable(shortcuts.Borders.Syntax())}, {"Sets the current grid cell borders on/off (", code("true"), "|", code("false"), ")"}},
			{{codeCopyable(shortcuts.BoundaryMode.Syntax())}, {"Sets the current grid boundary ", italic("mode"), " (", code(`Dead`), "|", code(`Alive`), ")"}},
			{{codeCopyable(shortcuts.Call.Syntax())}, {"Calls another shortcut by ", italic("name")}},
			{{codeCopyable(shortcuts.CellColorAlive.Syntax())}, {"Sets the current grid cell alive ", italic("color"), " (as HTML hex color, ", code(`#rrggbb`), ")"}},
			{{codeCopyable(shortcuts.CellColorDead.Syntax())}, {"Sets the current grid cell dead ", italic("color"), " (as HTML hex color, ", code(`#rrggbb`), ")"}},
			{{codeCopyable(shortcuts.CellColorBorder.Syntax())}, {"Sets the current grid cell border ", italic("color"), " (as HTML hex color, ", code(`#rrggbb`), ")"}},
			{{codeCopyable(shortcuts.CellSize.Syntax())}, {"Sets the current grid cell size to ", italic("n"), " (3-100)"}},
			{{codeCopyable(shortcuts.Clear.Syntax())}, {"Clears the current grid"}},
			{{codeCopyable(shortcuts.Export.Syntax())}, {"Exports the current grid as RLE"}},
			{{codeCopyable(shortcuts.ExportImage.AltSyntax(1))}, {"Exports the current grid as PNG"}},
			{{codeCopyable(shortcuts.ExportImage.Syntax())}, {
				"Exports the current grid as PNG with ", italic("metadata"),
				"\nWhere ", italic("metadata"), " is a semi-colon delimited list of key/value pairs. Example:\n",
				"  ", codeCopyable("export-image:Author=Me;Rule=%rule"),
				"\nValues can contain format tokens.", shortcutFormatTokens,
			}},
			{{codeCopyable(shortcuts.GridHeight.Syntax())}, {"Sets the current grid height to ", italic("n"), " (2-1000)"}},
			{{codeCopyable(shortcuts.GridSize.Syntax())}, {"Sets the current grid size to ", italic("w"), " (width) X ", italic("h"), " (height)"}},
			{{codeCopyable(shortcuts.GridWidth.Syntax())}, {"Sets the current grid width to ", italic("n"), " (2-1000)"}},
			{{codeCopyable(shortcuts.HeatMap.AltSyntax(1))}, {"Turns on the heat map instrument (defaults to ", code(`Activity`), ")"}},
			{{codeCopyable(shortcuts.HeatMap.Syntax())}, {"Turns on the heat map instrument with ", italic("type"), " (", code(`Activity`), "|", code(`Age`), "|", code(`Births`), "|", code(`Freshness`), "|", code(`Longevity`), "|", code(`Occupancy`), "|", code(`Phase Parity`), "|", code(`All`), ")\n", italic("(other values will turn heat mapper off)")}},
			{{codeCopyable(shortcuts.HeatMapColors.Syntax())}, {"Sets the colors used for heat map images - where ", italic("colors"), " is a semi-colon delimited list of colors (as HTML hex color, ", code(`#rrggbb`), ")\n",
				italic("The list must contain at least two valid colors - anything else resets heat map colors.")}},
			{{codeCopyable(shortcuts.HeatMapReveal.Syntax())}, {"Reveals (shows) the current heat map"}},
			{{codeCopyable(shortcuts.HeatMapSave.AltSyntax(1))}, {
				"Saves the current heat map as PNG",
				"\n", italic("Note: If "), code("heat-map:type"), italic(" is set to "), code("All"), italic(" then images for each heat map type are created. No images are created when heat mapping is off."),
			}},
			{{codeCopyable(shortcuts.HeatMapSave.Syntax())}, {
				"Saves the current heat map as PNG with ", italic("metadata"),
				"\nWhere ", italic("metadata"), " is a semi-colon delimited list of key/value pairs. Example:\n",
				"  ", codeCopyable("heat-map-save:Author=Me;Rule=%rule"),
				"\nValues can contain format tokens.", shortcutFormatTokens,
			}},
			{{codeCopyable(shortcuts.Log.Syntax())}, {"Logs a message to ", code(`output.log`), " file - the ", italic("message"), " is any string, possibly containing format tokens.", shortcutFormatTokens}},
			{{codeCopyable(shortcuts.MaxAdjacents.Syntax())}, {"Alters the current grid so that no cell can have more than ", italic("n"), " alive neighbours"}},
			{{codeCopyable(shortcuts.Name.Syntax())}, {"Sets the name for output files (e.g. exports, heat map images)\nThe name can be any string that is a valid filename/filepath - plus any format tokens.", shortcutFormatTokens}},
			{{codeCopyable(shortcuts.RandomAdditions.AltSyntax(1))}, {"Applies random additions (alive cells) to the current grid (according to current randomization setting)"}},
			{{codeCopyable(shortcuts.RandomAdditions.Syntax())}, {"Applies random additions (alive cells) to the current grid - according to ", italic("n"), " (0-100)"}},
			{{codeCopyable(shortcuts.RandomChanges.AltSyntax(1))}, {"Applies random changes (noise) to the current grid (according to current randomization setting)"}},
			{{codeCopyable(shortcuts.RandomChanges.Syntax())}, {"Applies random changes (noise) to the current grid - according to ", italic("n"), " (0-100)"}},
			{{codeCopyable(shortcuts.RandomCull.AltSyntax(1))}, {"Applies random cull to the current grid (according to current randomization setting)"}},
			{{codeCopyable(shortcuts.RandomCull.Syntax())}, {"Applies random cull to the current grid - according to ", italic("n"), " (0-100)"}},
			{{codeCopyable(shortcuts.Randomization.AltSyntax(2))}, {"Increments the randomization setting"}},
			{{codeCopyable(shortcuts.Randomization.AltSyntax(-2))}, {"Decrements the randomization setting"}},
			{{codeCopyable(shortcuts.Randomization.Syntax())}, {"Sets the randomization setting to ", italic("n"), " (0-100)"}},
			{{codeCopyable(shortcuts.Randomize.AltSyntax(1))}, {"Randomize the current grid (according to current randomization setting)"}},
			{{codeCopyable(shortcuts.Randomize.Syntax())}, {"Randomize the current grid  - according to ", italic("n"), " (0-100)"}},
			{{codeCopyable(shortcuts.RandomizePopulation.AltSyntax(1))}, {"Randomize the population of the current grid (according to current randomization setting)\n", italic("Note: This differs from `randomize` because that randomizes each cell in isolation - therefore it is not at all guaranteed that the final grid will have the specified % density. This action ensures that the overall grid density matches the specified randomization setting.")}},
			{{codeCopyable(shortcuts.RandomizePopulation.Syntax())}, {"Randomize the population of the current grid - according to ", italic("n"), " (0-100)"}},
			{{codeCopyable(shortcuts.Record.AltSyntax(1))}, {"Turns the record instrument on"}},
			{{codeCopyable(shortcuts.Record.Syntax())}, {"Turns the record instrument on or off"}},
			{{codeCopyable(shortcuts.AnimationSave.Syntax())}, {"Saves the current recording as an animation ", italic("(only works when recording is on - e.g. "), codeCopyable(`record:1`), italic(")")}},
			{{codeCopyable(shortcuts.AnimationFormat.Syntax())}, {"Sets the current recording animation format"}},
			{{codeCopyable(shortcuts.RepeatDetect.AltSyntax(1))}, {"Turns the repeat detect instrument on"}},
			{{codeCopyable(shortcuts.RepeatDetect.Syntax())}, {"Turns the repeat detect instrument on or off"}},
			{{codeCopyable(shortcuts.RepeatDetectSave.Syntax())}, {"Saves the current repeat detection (if enabled) as a report json"}},
			{{codeCopyable(shortcuts.Repeat.Syntax())}, {"Repeats the following actions (comma separated) ", italic("n"), " times\nExample:",
				codeBlock{code: "repeat:100,randomize,step-ahead-by:500,heat-map-save"}, "Repeats can also be 'nested', example:",
				codeBlock{code: "repeat:5,rule-perm++,randomize,heat-map:Activity,repeat:10,step-ahead-by:10,heat-map-save`"}}},
			{{codeCopyable(shortcuts.ReplaySnapshot.Syntax())}, {"Replays a snapshot\nResets the grid to the last snapshot (without removing the snapshot from the stack)\n", italic("Note: If there is no current snapshot - the grid is randomized and a snapshot added")}},
			{{codeCopyable(shortcuts.BornWith.Syntax())}, {"Sets or adjusts the current life rule born with - where ", italic("x"), " can be:",
				hanging{indent: 10, gap: 10, prefix: "•", content: "`23` sets to `B23`"},
				hanging{indent: 10, gap: 10, prefix: "•", content: "`|23` ORs with 2 and 3 (effectively turning just those on)"},
				hanging{indent: 10, gap: 10, prefix: "•", content: "`&23` ANDs with 2 and 3"},
				hanging{indent: 10, gap: 10, prefix: "•", content: "`!23` flips 2 and 3"},
				hanging{indent: 10, gap: 10, prefix: "•", content: "`!` flips all - e.g. `B23` becomes `B0145678`"},
				italic("Note: the current survives with, `S`, is unchanged")}},
			{{codeCopyable(shortcuts.BornWith.AltSyntax(2))}, {"Increments the current life rule born with permutation (high order 9 bits)"}},
			{{codeCopyable(shortcuts.BornWith.AltSyntax(-2))}, {"Decrements the current life rule born with permutation (high order 9 bits)"}},
			{{codeCopyable(shortcuts.RulePerm.Syntax())}, {"Sets the current life rule permutation to ", italic("n"), " (0-262143)"}},
			{{codeCopyable(shortcuts.RulePerm.AltSyntax(2))}, {"Increments the current life rule permutation"}},
			{{codeCopyable(shortcuts.RulePerm.AltSyntax(-2))}, {"Decrements the current life rule permutation"}},
			{{codeCopyable(shortcuts.RuleInt.Syntax())}, {"Sets the current life rule integer to ", italic("n"), " (0-262143)"}},
			{{codeCopyable(shortcuts.RuleInt.AltSyntax(2))}, {"Increments the current life rule integer"}},
			{{codeCopyable(shortcuts.RuleInt.AltSyntax(-2))}, {"Decrements the current life rule integer"}},
			{{codeCopyable(shortcuts.SurvivesWith.Syntax())}, {"Sets or adjusts the current life rule survives with - where ", italic("x"), " can be:",
				hanging{indent: 10, gap: 10, prefix: "•", content: "`23` sets to `S23`"},
				hanging{indent: 10, gap: 10, prefix: "•", content: "`|23` ORs with 2 and 3 (effectively turning just those on)"},
				hanging{indent: 10, gap: 10, prefix: "•", content: "`&23` ANDs with 2 and 3"},
				hanging{indent: 10, gap: 10, prefix: "•", content: "`!23` flips 2 and 3"},
				hanging{indent: 10, gap: 10, prefix: "•", content: "`!` flips all - e.g. `S23` becomes `S0145678`"},
				italic("Note: the current born with, `B`, is unchanged")}},
			{{codeCopyable(shortcuts.SurvivesWith.AltSyntax(2))}, {"Increments the current life rule survives with permutation (low order 9 bits)"}},
			{{codeCopyable(shortcuts.SurvivesWith.AltSyntax(-2))}, {"Decrements the current life rule survives with permutation (low order 9 bits)"}},
			{{codeCopyable(shortcuts.Rule.Syntax())}, {"Sets the current life rule (by ", italic("name"), " or specified ", italic("rle"), ", e.g. `B2/S23`)"}},
			{{codeCopyable(shortcuts.RunRecipe.AltSyntax(1))}, {"Runs the currently selected grid recipe"}},
			{{codeCopyable(shortcuts.RunRecipe.Syntax())}, {"Runs the grid recipe for the specified ", italic("filename")}},
			{{codeCopyable(shortcuts.Run.Syntax())}, {"Runs simulation on the current grid"}},
			{{codeCopyable(shortcuts.Sleep.Syntax())}, {"Sleeps for a given ", italic("ms"), " (millisecond) period"}},
			{{codeCopyable(shortcuts.Snapshot.Syntax())}, {"Snapshots the current grid"}},
			{{codeCopyable(shortcuts.Status.Syntax())}, {
				"Sets a status ", italic(`msg`), " in the ", StatusBar.link("statusbar"), ".",
				" The message can contain format tokens", shortcutFormatTokens,
				"To clear the message use ", codeCopyable(shortcuts.Status.AltSyntax(1)), " (i.e. without msg).",
				"\nThe message is persisted for a short time after the shortcut completes.",
			}},
			{{codeCopyable(shortcuts.StepAhead.AltSyntax(1))}, {"Performs a step ahead (by current step ahead setting)"}},
			{{codeCopyable(shortcuts.StepAhead.Syntax())}, {"Sets the current step ahead setting to ", italic("n"), " (1-9999)"}},
			{{codeCopyable(shortcuts.StepAhead.AltSyntax(2))}, {"Increments the current step ahead setting"}},
			{{codeCopyable(shortcuts.StepAhead.AltSyntax(-2))}, {"Decrements the current step ahead setting"}},
			{{codeCopyable(shortcuts.StepAheadBy.Syntax())}, {"Performs a step ahead by ", italic("n"), " (>0)"}},
			{{codeCopyable(shortcuts.StepBackBy.Syntax())}, {"Performs a step back by ", italic("n"), " (>0) ", italic("(only works when recording is on - e.g. "), codeCopyable(`record:1`), italic(")")}},
			{{codeCopyable(shortcuts.StepDelay.Syntax())}, {"Sets the current step delay to ", italic("ms"), " (0-2000)"}},
			{{codeCopyable(shortcuts.StepDelay.AltSyntax(2))}, {"Increments the current step delay setting"}},
			{{codeCopyable(shortcuts.StepDelay.AltSyntax(-2))}, {"Decrements the current step delay setting"}},
			{{codeCopyable(shortcuts.Step.Syntax())}, {"Performs a single simulation step on the current grid"}},
			{{codeCopyable(shortcuts.Stop.Syntax())}, {"Stops simulation on the current grid"}},
			{{codeCopyable(shortcuts.UndoToSnapshot.Syntax())}, {"Restores grid to last snapshot"}},
			{{codeCopyable(shortcuts.WrapMode.Syntax())}, {"Sets the current grid wrap ", italic("mode"), " (", code(`none`), "|", code(`horizontal`), "|", code(`vertical`), "|", code(`all`), "|", code(`toroidal`), ")"}},
			{{codeCopyable(shortcuts.Zoom.Syntax())}, {"Sets the current zoom factor to ", italic("n%"), " (10-1000)"}},
			{{codeCopyable(shortcuts.NextMetaRule.Syntax())}, {"Skips to the next rule in ", italic(`meta`), " and sets the current rule to that (see ", MetaRules.link("meta rules"), ")"}},
			{{codeCopyable(shortcuts.PreviousMetaRule.Syntax())}, {"Skips to the previous rule in ", italic(`meta`), " and sets the current rule to that (see ", MetaRules.link("meta rules"), ")"}},
			{{codeCopyable(shortcuts.IterateMetaRule.Syntax())}, {"Iterates over a ", italic(`meta`), " rule (similar to `repeat:n`)\n", italic("Note: If the meta rule contains commas, it should enclosed in single or double quotes.")}},
			{{codeCopyable(shortcuts.AddCollectedRule.Syntax())}, {"Adds the current grid rule to the collected rules (see ", CollectedRules.link("collected rules"), ")"}},
			{{codeCopyable(shortcuts.RemoveCollectedRule.Syntax())}, {"Removes the current grid rule from the collected rules"}},
			{{codeCopyable(shortcuts.NextCollectedRule.Syntax())}, {"Skips to the next rule in collected rules"}},
			{{codeCopyable(shortcuts.PreviousCollectedRule.Syntax())}, {"Skips to the previous rule in collected rules"}},
			{{codeCopyable(shortcuts.IterateCollectedRules.Syntax())}, {"Iterates over collected rules"}},
			{{codeCopyable(shortcuts.ButtonAssign.Syntax())}, {"Assigns statusbar button to the current shortcut"}},
			{{codeCopyable(shortcuts.ButtonAssign.AltSyntax(1) + ":A|B|C|D|E|F=name")}, {
				"Assigns statusbar button to the specified shortcut ", italic("name"), ".\n",
				italic("Note: If the name is empty, the button is unassigned."),
			}},
			{{codeCopyable(shortcuts.Javascript.Syntax())}, {"Executes the specified javascript as part of the shortcut ", italic("(see "), ShortCutsJsRef.linkItalic("javascript reference"), italic(")"), "."}},
		},
		spaceBefore: 8,
	},
	expandable{border: true, indent: 4, padding: 4, spaceBefore: 8,
		title: content{bold("Advanced - Variables")},
		content: content{
			"Shortcut variables:",
			indent{indent: 4, content: content{
				hanging{prefix: "• ", content: "Add persistent, reusable state to shortcuts. They can be used to store values, construct or modify values over a sequence of actions, and pass those values to and from other shortcut actions. Variables are shared between shortcuts, including shortcuts invoked using call-shortcut, and their values are retained between application runs."},
				hanging{prefix: "• ", content: "Store values as text. When a variable is used, its value is interpreted according to the type required by the action. This allows variables to supply text, numbers, colours, rule permutations, and other supported values without variables themselves having a fixed type. If a value cannot be converted to the required type, the action/operation is ignored."},
				hanging{prefix: "• ", content: "Values may include substitutions from other variables and format tokens, allowing values to be constructed dynamically from the current application state. Arithmetic and bitwise operations provide further manipulation where appropriate."},
				hanging{prefix: "• ", content: "Can control repetition and stepping, change application settings, construct filenames or colours, manipulate rules, and support more complex macro-like sequences without requiring an expression language."},
			}},
			h4("Declaration"),
			indent{indent: 8, content: content{
				"To declare (set) a variable as an action:",
				codeBlock{code: `variable:name=value`},
				"where:",
				hanging{prefix: content{code("name"), " "}, content: "is a non-empty string name for the variable"},
				hanging{prefix: content{code("value"), " "}, content: "is the initial value for the variable - which may include format tokens and/or references to other variables"},
				italic("Note: Variable names are case-sensitive, must start with a letter ('a'-'z' 'A-Z') and may contain characters: 'a'-'z', 'A-Z', '0'-'9' or '_'"),
			}},
			h4("Manipulation"),
			indent{indent: 8, content: content{
				"Existing variables can be manipulated using the following basic syntax:",
				codeBlock{code: `variable:name operator operand`},
				"where ", code("operator"), " can be:",
				hanging{prefix: content{code("+="), " "}, content: content{
					"adds the operand value to the named variable value\n",
					italic("If both the named variable and operand are integer values then an arithmetic addition is performed - otherwise a string concatenation is performed."),
					hanging{prefix: "Examples: ", content: content{
						codeBlock{code: `variable:myvar=1
variable:myvar += 2
log:$$myvar is now $myvar
variable:myvar +=foo
log:$$myvar is now $myvar`},
					}},
				}},
				hanging{prefix: content{code("-="), " "}, content: content{
					"subtracts the operand value from the named variable value\n",
					italic("Both the named variable and operand must be integer values - otherwise no operation is performed."),
					hanging{prefix: "Example: ", content: content{
						codeBlock{code: `variable:myvar=3
variable:myvar -= 2`},
					}},
				}},
				hanging{prefix: content{code("*="), " "}, content: content{
					"multiplies the named variable value by the operand value (and stores result back to the variable)\n",
					italic("Both the named variable and operand must be integer values - otherwise no operation is performed."),
				}},
				hanging{prefix: content{code("/="), " "}, content: content{
					"divides the named variable value by the operand value (and stores result back to the variable)\n",
					italic("Both the named variable and operand must be integer values - otherwise no operation is performed."),
				}},
				hanging{prefix: content{code("&="), " "}, content: content{
					"performs bit AND of the named variable value and operand value (and stores result back to the variable)\n",
					italic("Both the named variable and operand must be integer values - otherwise no operation is performed."),
				}},
				hanging{prefix: content{code("|="), " "}, content: content{
					"performs bit OR of the named variable value and operand value (and stores result back to the variable)\n",
					italic("Both the named variable and operand must be integer values - otherwise no operation is performed."),
				}},
				hanging{prefix: content{code("^="), " "}, content: content{
					"performs bit XOR of the named variable value and operand value (and stores result back to the variable)\n",
					italic("Both the named variable and operand must be integer values - otherwise no operation is performed."),
				}},
				hanging{prefix: content{code("++"), " "}, content: content{
					boldItalic("(no operand)"), " increments the named variable value\n",
					italic("The named variable must be an integer value - otherwise no operation is performed."),
				}},
				hanging{prefix: content{code("--"), " "}, content: content{
					boldItalic("(no operand)"), " decrements the named variable value\n",
					italic("The named variable must be an integer value - otherwise no operation is performed."),
				}},
			}},
			h4("Usage"),
			indent{indent: 8, content: content{
				"Variables may be used as format tokens by prefixing the variable name with $. The variable reference is replaced with its current value before the action is performed.",
				hanging{prefix: "Example: ", content: content{
					codeBlock{code: `variable:steps=100
step-ahead-by:$steps`},
				}},
				"\nVariable references may also be embedded within other text and combined with other variables or format tokens",
				hanging{prefix: "Example: ", content: content{
					codeBlock{code: `variable:R=AA
variable:G=00
variable:B=80
variable:rgb=#$R$G$B
cell-color-alive:$rgb`},
				}},
			}},
			h4("Clearing"),
			indent{indent: 8, content: content{
				"Variables can also be individually or collectively cleared.",
				hanging{prefix: content{code("clear-variable:name"), " "}, content: content{"clears the named variable."}},
				hanging{prefix: content{code("clear-variables"), " "}, content: content{"clears all variables."}},
				hanging{prefix: content{code("clear-variables:name;name,..."), " "}, content: content{"clears the named variables."}},
				italic("Note: Clearing a variable that has not been defined has no effect and causes no error."),
			}},
			h4("Logging"),
			indent{indent: 8, content: content{
				"For shortcut debugging assistance, all current variables can be logged with:",
				indent{indent: 20, content: content{codeCopyable("log:%variables")}},
				"The current variables, along with their values, will be written to the ", code("output.log"), " file.",
			}},
		},
	},
	expandable{border: true, indent: 4, padding: 4, spaceBefore: 8,
		title: content{bold("Advanced - Conditionals/Stopping/Breaking")},
		content: content{
			bold("Action "), code("if"), bold(" provides simple conditional control over actions execution."),
			" Subsequent actions are only executed when condition evaluates to true.",
			hanging{prefix: "Example: ", content: content{
				codeBlock{code: `if:%MM == 9,log:"it is September"`},
				italic("logs the message if the current month is September."),
			}},
			hanging{spaceAfter: 8, prefix: "Example: ", content: content{
				codeBlock{code: `if:%MM >= 6,if:%MM <= 8,log:"it's summer!"`},
				italic("logs the message only if it's summer."),
			}},
			bold("Action "), code("while"), bold(" provides simple iterative conditional control over actions execution."),
			" Subsequent actions are only executed while the condition evaluates to true.",
			hanging{spaceAfter: 8, prefix: "Example: ", content: content{
				codeBlock{code: `variable:max=10
variable:i=0
while:$i < $max,log:$i,variable:i++`},
				italic("logs the incrementing value of "), boldItalic("$i"), italic(" while it is less than the maximum."),
			}},
			bold("Stopping and breaking provide simple conditional control over shortcut execution."),
			" A shortcut, called shortcut, or repeat loop can be terminated when a specified condition becomes true.",
			hanging{prefix: content{code(`stop-if`), " "}, content: content{
				"stops the entire shortcut execution. This applies regardless of where it occurs, including within a repeat loop or a shortcut invoked using call-shortcut.",
				hanging{prefix: "Example: ", content: content{
					codeBlock{code: `stop-if:%MM > 9`},
					italic("stops the entire shortcut execution if the current month is greater than September"),
				}},
			}},
			hanging{prefix: content{code(`break-if`), " "}, content: content{
				"stops only the current execution scope. Within a repeat loop it breaks from that loop, allowing execution to continue with the following action. Within a called shortcut it stops that shortcut and returns to its caller. At the top level it stops the current shortcut.",
				hanging{prefix: "Example: ", content: content{
					codeBlock{code: `variable:perc50=%grid-width
variable:perc50 *= %grid-height
variable:perc50 /= 2
break-if:%population < $perc50`},
					italic("stops the current shortcut scope if the grid population is less than 50%"),
				}},
			}},
			"Notes:",
			hanging{indent: 8, prefix: "• ", content: content{
				"Allowed comparators are ", code("=="), italic(" (equals)"), ", ", code("!="), italic(" (not equals)"),
				", ", code(">"), italic(" (greater than)"), ", ", code("<"), italic(" (less than)"),
				", ", code(">="), italic(" (greater than or equals)"), " and ", code("<="), italic(" (less than or equals)"),
			}},
			hanging{indent: 8, prefix: "• ", content: content{
				"If both values resolve to integers then a numeric comparison is evaluated - otherwise a case-insensitive string comparison is evaluated.",
			}},
			hanging{indent: 8, prefix: "• ", content: content{
				"Both compared values can be explicit, format tokens or variables.",
			}},
			hanging{indent: 8, prefix: "• ", content: content{
				"Additionally, the are special pre-defined conditions:",
				hanging{indent: 8, prefix: content{codeCopyable("rule-collected"), " "}, content: content{"evaluates to true if the current rule is collected."}},
				hanging{indent: 8, prefix: content{codeCopyable("!rule-collected"), " "}, content: content{"evaluates to true if the current rule is not collected."}},
				hanging{indent: 8, prefix: content{codeCopyable("variable-exists:name"), " "}, content: content{
					"evaluates to true if the named variable is defined.",
					hanging{prefix: "Example: ", content: content{codeBlock{code: `break-if:variable-exists:foo`}}},
				}},
				hanging{indent: 8, prefix: content{codeCopyable("!variable-exists:name"), " "}, content: content{"evaluates to true if the named variable is not defined."}},
				hanging{indent: 8, prefix: content{codeCopyable("variable-exists-n:name"), " "}, content: content{"evaluates to true if the named variable is defined and is numeric."}},
				hanging{indent: 8, prefix: content{codeCopyable("!variable-exists-n:name"), " "}, content: content{
					"evaluates to true if the named variable is not defined or is defined but not numeric.",
					hanging{prefix: "Example: ", content: content{codeBlock{code: `if:!variable-exists-n:set_pop,var:set_pop=10
if:$set_pop<100,var:set_pop++
randomize-population:$set_pop
status:Population = $set_pop`}}},
				}},
			}},
		},
	},
}

var shortcutFormatTokens = expandable{
	border:      true,
	indent:      8,
	rightMargin: 8,
	title:       "Format tokens",
	content: content{
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%rule"), " e.g. transposed to `B2/S23`"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%born"), " or ", codeCopyable("%rule-born-with"), " e.g. transposed to `B2`"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%survives"), " or ", codeCopyable("%rule-survives-with"), " e.g. transposed to `S23`"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%perm"), " or ", codeCopyable("%rule-perm"), " e.g. transposed to `4108`"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%step"), " transposed to the current step of the grid"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%population"), " transposed to the current population of the grid"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%R"), " transposed to last (current) repeat iteration"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%r"), " transposed to repeat iteration(s)"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%rand"), " transposed to current randomization setting"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%repeat-detecting"), " whether repeat detect is on/off (`1`/`0`)"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%repeat-at"), " transposed to repeat at step"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%repeat-first"), " transposed to repeat first step"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%repeat-found"), " transposed to whether a repeat was found (", code(`0`), "|", code(`1`), ")"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%repeat-period"), " transposed to repeat period"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%record"), " whether record is on/off (`1`/`0`)"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%recorded-steps"), " the number of recorded steps"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%heat-map"), " whether heat mapping is on/off (`1`/`0`)"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%heat-map-type"), " the current heat mapping type"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%step-ahead"), " the current step ahead setting"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%step-back"), " the current step back setting"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{
			codeCopyable("%wrap-mode"), ", ", codeCopyable("%boundary-mode"), ", ", codeCopyable("%step-delay"), ", ",
			codeCopyable("%grid-width"), ", ", codeCopyable("%grid-height"), ", ", codeCopyable("%grid-size"), ", ",
			codeCopyable("%cell-size"), ", ", codeCopyable("%borders"), ", ",
			codeCopyable("%cell-color-alive"), ", ", codeCopyable("%cell-color-dead"), ", ", codeCopyable("%cell-color-border"),
			", ", codeCopyable("%zoom"),
			"\nmiscellaneous current settings"},
		},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%now"), " transposed to current date/time (in format `YYYY-MM-DD hh-mm-ss-nnn`)"}},
		hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%timestamp"), " transposed to current date/time (in format `YYYY-MM-DDThh-mm-ss.nnnZ`)"}},
		expandable{border: true, indent: 6, spaceAfter: 6, rightMargin: 6,
			title: content{bold("Individual date/time format tokens")},
			content: content{
				"Format tokens ", codeCopyable(`%now`), " and ", codeCopyable(`%timestamp`), " are pre-defined date/time formats - you can compose your own date/time formatting using the following tokens:",
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%YYYY"), " 4 digit year (e.g. `2026`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%YY"), " 2 digit year (e.g. `26`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%MM"), " 2 digit month (e.g. `09`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%M"), " 1 or 2 digit month (e.g. `9`, `10`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%DD"), " 2 digit day (e.g. `09`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%D"), " 1 or 2 digit day (e.g. `9`, `10`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%Dth"), " suffixed day (e.g. `1st`, `2nd`, `3rd`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%Mmm"), " Abbreviated month name (e.g. `Sep`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%Mmmm"), " Full month name (e.g. `September`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%Ddd"), " Abbreviated day of week (e.g. `Fri`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%Dddd"), " Full day of week (e.g. `Friday`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%hh"), " 2 digit hour, 24-hour clock  (e.g. `16`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%h"), " 1 or 2 digit hour, 12-hour clock (e.g. `4`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%pm"), " AM/PM, 12-hour clock (e.g. `am`|`pm`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%PM"), " AM/PM, 12-hour clock (e.g. `AM`|`PM`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%mm"), " 2 digit minutes"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%ss"), " 2 digit seconds"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%nnn"), " 3 digit milliseconds"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%nnnnnn"), " 6 digit nanoseconds"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%tz"), " Timezone offset (e.g. `±hhmm`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%tzhm"), " Timezone offset (e.g. `±hh:mm`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%tzz"), " Timezone offset (e.g. `Z` or `±hhmm`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%tzzhm"), " Timezone offset (e.g. `Z` or `±hh:mm`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%ww"), " 2 digit ISO-8601 week number (e.g. `01`, `52`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%w"), " 1 or 2 digit ISO-8601 week number (e.g. `1`, `52`)"}},
				hanging{gap: 10, indent: 10, prefix: "•", content: content{codeCopyable("%wy"), " 4 digit ISO-8601 week number year"}},
			},
		},
	},
}

var contentShortCutsJsRef = content{
	"Shortcuts can execute javascript using the following ", ShortCutsRef.link("action"), ":",
	codeBlock{indent: 20, code: "javascript:filename"},
	hanging{indent: 20, prefix: "• ", content: content{"where ", italic("filename"), " is either an absolute path - or is taken as a relative path from ", code(`[user]/Documents/GoGoL/`)}},
	hanging{indent: 20, prefix: "• ", content: content{"the script is compiled as ECMAScript 5.1 - any compilation errors are logged to ", code(`[user]/Documents/GoGoL/errors.log`)}},
	hanging{indent: 20, prefix: "• ", content: content{"the script is not cached - it is compiled from the file each time the shortcut is run. Therefore, between shortcut runs, the script can be modified with external editors."}},
	"The following functions made available:",
	table{
		borders: true,
		padding: 4,
		columns: []tableColumn{
			{
				width:  40,
				header: "Function",
			},
			{
				header: "Description & Syntax",
			},
		},
		rows: []tableRow{
			{{codeCopyable(`addCapturedPattern(Pattern)`)}, {
				"Adds the pattern to ", CapturedPatterns.link("captured patterns"), ".",
				"\nReturns: none (", code(`undefined`), ")",
			}},
			{{codeCopyable(`addCollectedRule([string|int|Rule])`)}, {"Adds a rule to the currently ", CollectedRules.link("collected rules"), ".",
				" If no arg is supplied, the current rule is added.",
				"\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`animationFormat([string])`)}, {"Gets, and optionally sets, the current animation format.",
				italic("\nNote: "), code(`mp4`), italic(" option is only available when "), code("ffmpeg"), italic(" is installed."), " [see ", url("https://ffmpeg.org/"), "]",
				"\nReturns: ", code(`string`), " - current animation format (", code(`gif`), "|", code(`mp4`), ")"}},
			{{codeCopyable(`animationSave([string[,int[,int]]])`)}, {
				"Saves the current recording as an animation ", italic("(only works when recording is on)"), ". The first arg is the filename - if omitted, a filename is generated.",
				"\nThe second, optional, arg is the number of start frames to skip. The third, optional, arg is the maximum number of frames to save.",
				"\nReturns: none (", code(`undefined`), ")",
			}},
			{{codeCopyable(`borders([bool])`)}, {"Gets, and optionally sets, whether cell borders are displayed.", "\nReturns: ", code(`bool`)}},
			{{codeCopyable(`boundaryMode([string])`)}, {"Gets, and optionally sets, the current grid boundary mode.", "\nReturns: ", code(`string`), " - current boundary mode (", code(`alive`), "|", code(`dead`), ")"}},
			{{codeCopyable(`cell(x:int,y:int[,bool])`)}, {
				"Gets, and optionally sets, a specific cell in the grid.",
				"\n", italic("Note: For large drawing operations, performance is better by creating and drawing a pattern."),
				"\nReturns: ", code(`bool`), " - if getting a cell, whether the cell is set (alive). For setting a cell, indicates whether the cell state was changed.",
			}},
			{{codeCopyable(`cellAliveColor([string])`)}, {"Gets, and optionally sets, the current cell alive color.", "\nReturns: ", code(`string`), " - current cell alive color (", code(`#rrggbb`), ")"}},
			{{codeCopyable(`cellBorderColor([string])`)}, {"Gets, and optionally sets, the current cell border color.", "\nReturns: ", code(`string`), " - current cell border color (", code(`#rrggbb`), ")"}},
			{{codeCopyable(`cellDeadColor([string])`)}, {"Gets, and optionally sets, the current cell dead color.", "\nReturns: ", code(`string`), " - current cell dead color (", code(`#rrggbb`), ")"}},
			{{codeCopyable(`cellSize([int])`)}, {"Gets, and optionally sets, the grid cell size (pixels).", "\nReturns: ", code(`int`), " - cell size (1-100)"}},
			{{codeCopyable(`clear()`)}, {"Clears the current grid.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`clearCollectedRules()`)}, {"Clears ", CollectedRules.link("collected rules"), ".",
				"\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`clearVariables([string...])`)}, {"Clears shortcut variables. If no args are supplied, all shortcut variables are cleared - otherwise only the specified variable names are cleared.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`collectedRules()`)}, {"Gets the currently ", CollectedRules.link("collected rules"), ".",
				"\nReturns: ", code(`array[Rule]`), " - list of collected rules",
				ruleObject,
			}},
			{{codeCopyable(`drawPattern(x:int,y:int,pattern)`)}, {
				"Draws the specified pattern onto the grid at position ", italic("x,y"),
				"\nThe pattern arg can be specified as a ", code(`Pattern`), italic(" (see "), code(`newPattern()`), italic(" / "), code(`patterns()`), italic(")"),
				" or as a string representing the pattern rle.",
				"\nReturns: none (", code(`undefined`), ")",
			}},
			{{codeCopyable(`export([string])`)}, {"Exports the current grid as an ", code(`.rle`), " file. The arg is the filename - if omitted, a filename is generated.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`exportImage([string[,obj]])`)}, {"Exports the current grid as a ", code(`.png`), " file. The first arg is the filename - if omitted, a filename is generated. The second arg is optional metadata for the png file.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`gridHeight([int])`)}, {"Gets, and optionally sets, the current grid height.", "\nReturns: ", code(`int`), " - current grid height (3-99999)"}},
			{{codeCopyable(`gridSize([int,int])`)}, {"Gets, and optionally sets, the current grid size (width X height).", "\nReturns: ", code(`string`), " - current grid size (", code(`"wXh"`), ")"}},
			{{codeCopyable(`gridWidth([int])`)}, {"Gets, and optionally sets, the current grid width.", "\nReturns: ", code(`int`), " - current grid width (3-99999)"}},
			{{codeCopyable(`heatMap([string])`)}, {"Gets, and optionally sets, the current heat map instrument.", "\nReturns: ", code(`string`), " (", code(`Activity`), "|", code(`Age`), "|", code(`Births`), "|", code(`Freshness`), "|", code(`Longevity`), "|", code(`Occupancy`), "|", code(`Phase Parity`), "|", code(`All`), " - ", code(`undefined`), " for no heat mapping)"}},
			{{codeCopyable(`heatMapColors([string...])`)}, {
				"Sets the colors used for heat map images.",
				"\nEach arg is a color in the format ", code(`"#rrggbb"`), ". The args must contain at least two valid colors - anything else resets heat map colors to defaults.",
				"\nWhen drawing heat map images, the heat map colors are used as gradient points.",
				"\nReturns: ", code(`array[string]`), " - current heat map colors (empty array indicates default colors)",
				expandable{indent: 10, rightMargin: 10, padding: 4, border: true, title: "Default colors", content: content{
					"The default colors are:",
					codeBlock{code: `[
	"#0f18b4", // deep blue
	"#1e5aff", // blue
	"#3cc8ff", // cyan
	"#f5f5eb", // near white
	"#ffdc50", // yellow
	"#ff8c14", // orange
	"#dc280a", // red
	"#6e0000", // dark red
]`},
				}},
			}},
			{{codeCopyable(`heatMapReveal([string])`)}, {
				"Reveals the current heat map in the UI.",
				italic("\nNote: If the current heat map type is "), code(`All`), italic(", use the optional string arg to denote which to reveal ("), code(`Activity`), "|", code(`Age`), "|", code(`Births`), "|", code(`Freshness`), "|", code(`Longevity`), "|", code(`Occupancy`), "|", code(`Phase Parity`), italic(")."),
				"\nReturns: none (", code(`undefined`), ")",
			}},
			{{codeCopyable(`heatMapSave(string[,object[,string]])`)}, {
				"Saves the current heat map as a ", code(".png"), italic(" (no images are saved when heat mapping is off)"), ".",
				"\nThe first string arg is the filename. The second, optional,  arg is metadata for the png file.",
				italic("\nNote: If the current heat map type is "), code(`All`), italic(", use the third optional string arg to denote which to save ("), code(`Activity`), "|", code(`Age`), "|", code(`Births`), "|", code(`Freshness`), "|", code(`Longevity`), "|", code(`Occupancy`), "|", code(`Phase Parity`), italic(")."),
				"\nReturns: none (", code(`undefined`), ")",
			}},
			{{codeCopyable(`maxAdjacents(int)`)}, {"Alters the current grid so that no cell can have more than the specified number (0-8) alive neighbours", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`newMetaRule(string)`), italic("\nor\n"), codeCopyable(`new MetaRule(string)`)}, {
				"Creates a new ", MetaRulesRef.link("meta rule"), ".",
				"\nIf the arg matches an existing meta rule name, that meta is returned - otherwise the meta rule is interpreted from the provided string (see ", MetaRulesRef.link("reference"), ")",
				"\nReturns: ", code(`MetaRule`),
				metaRuleObject,
			}},
			{{codeCopyable(`newPattern(rle:string)`), italic("\nor\n"), codeCopyable(`new Pattern(rle:string)`)}, {
				"Creates a new pattern from the supplied rle string.",
				"\nReturns: ", code(`Pattern`), italic(" (or "), code(`undefined`), italic(" in case of pattern create error)"),
				patternObject,
			}},
			{{codeCopyable(`newPattern(w:int,h:int,cells:array)`), italic("\nor\n"), codeCopyable(`new Pattern(w:int,h:int,cells:array)`)}, {
				"Creates a new pattern from the supplied args - where:",
				hanging{indent: 10, prefix: content{code("w")}, content: content{" - is the pattern width"}},
				hanging{indent: 10, prefix: content{code("h")}, content: content{" - is the pattern height"}},
				hanging{indent: 10, prefix: content{code("cells")}, content: content{
					" - is the pattern cells - a flat array arranged by rows.",
					italic("\n(the length of the array "), boldItalic("must"), italic(" match width X height)"),
					"\nEach array element is evaluated to ", code(`bool`), " for dead or alive.",
					" For best performance and to avoid coercion - define cells array as ", code(`new Uint8Array(w * h)`),
				}},
				"Returns: ", code(`Pattern`), italic(" (or "), code(`undefined`), italic(" in case of pattern create error)"),
				patternObject,
			}},
			{{codeCopyable(`newRule(string|int)`), italic("\nor\n"), codeCopyable(`new Rule(string|int)`)}, {
				"Creates a new Rule from the supplied arg:",
				hanging{indent: 10, prefix: content{code("string")}, content: content{" - rule name or specified ", italic("rle"), ", e.g. ", code(`"B2/S23"`)}},
				hanging{indent: 10, prefix: content{code("int")}, content: content{"- rule permutation (0-262143)"}},
				"Returns: ", code(`Rule`),
				ruleObject,
			}},
			{{codeCopyable(`patternFromGrid([x,y,w,h:int])`)}, {"Creates a pattern from a portion of the grid",
				" - as defined by args ", boldItalic("x"), ", ", boldItalic("y"), ", ", boldItalic("w"), " & ", boldItalic("h"), ".",
				" If no args are supplied, the entire grid is used.",
				"\nReturns: ", code(`Pattern`),
				patternObject,
			}},
			{{codeCopyable(`patterns()`)}, {"Retrieves all patterns in the current pattern library.",
				"\nReturns: ", code(`array[Pattern]`),
				patternObject,
			}},
			{{codeCopyable(`population()`)}, {"Gets the current grid population.", "\nReturns: ", code(`int`)}},
			{{codeCopyable(`prompt(string[,string...])`)}, {
				"Prompts the user for input. The prompt is displayed in the ", StatusBar.link("statusbar"), " and the function is blocking until input is acquired.",
				"\nThe first string arg is the label for the prompt.",
				" Successive string args are allowed responses (case insensitive):",
				hanging{indent: 8, prefix: "• ", content: "if allowed responses are specified, the function returns as soon as the user types one of them (e.g. key input)."},
				hanging{indent: 8, prefix: "• ", content: content{"if no allowed responses are supplied, the user can input any length and the function returns when the user presses ", keys{keyEnter}, "."}},
				"Returns: ", code(`string`), " - the user input",
			}},
			{{codeCopyable(`promptDialog(string,array[,callback])`)}, {
				"Prompts the user to complete a dialog. The dialog is displayed from the ", StatusBar.link("statusbar"), " and the function is blocking until the user completes the dialog (by pressing either the ", bold("Ok"), " or ", bold("Cancel"), " buttons).",
				"\nThe first string arg is the label shown in the statusbar for the dialog.",
				" The second array arg is the definition for the dialog.",
				"\nReturns: ", code(`object`), " - the dialog response values (or ", code(`undefined`), " if the user presses the ", bold("Cancel"), " button)",
				dialogDefinition,
			}},
			{{codeCopyable(`promptMenu(string[,string...])`)}, {
				"Prompts the user to select from a menu. The prompt is displayed in the ", StatusBar.link("statusbar"), " and the function is blocking until the user selects a menu option.",
				"\nThe first string arg is the label for the prompt.",
				" Successive string args are the menu options (you must supply at least one option).",
				"\nReturns: ", code(`int`), " - the menu option selected",
			}},
			{{codeCopyable(`randomAdditions([int])`)}, {"Applies random additions (alive cells) to the current grid. The optional arg is the randomization (0-100) - if omitted, the current randomization setting is used.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`randomChanges([int])`)}, {"Applies random changes (noise) to the current grid. The optional arg is the randomization (0-100) - if omitted, the current randomization setting is used.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`randomCull([int])`)}, {"Applies random cull to the current grid. The optional arg is the randomization (0-100) - if omitted, the current randomization setting is used.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`randomization([int])`)}, {"Gets, and optional sets, the current randomization setting.", "\nReturns: ", code(`int`), " - current randomization setting (0-100)"}},
			{{codeCopyable(`randomize([int])`)}, {"Randomizes the current grid. The optional arg is the randomization (0-100) - if omitted, the current randomization setting is used.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`randomizePopulation([int])`)}, {"Randomize the population of the current grid. The optional arg is the randomization (0-100) - if omitted, the current randomization setting is used.\n", italic("Note: This differs from "), code(`randomize()`), italic(" because that randomizes each cell in isolation - therefore it is not at all guaranteed that the final grid will have the specified % density. This function ensures that the overall grid density matches the specified randomization setting."), "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`record([bool])`)}, {"Gets, and optionally sets, whether the record instrument is on.", "\nReturns: ", code(`bool`)}},
			{{codeCopyable(`recordInstrument()`)}, {
				"Returns the ", code(`RecordInstrument`), " (or ", code(`undefined`), " if not enabled)",
				recordInstrumentObject,
			}},
			{{codeCopyable(`removeCollectedRule([string|int|Rule])`)}, {"Removes a rule from the currently ", CollectedRules.link("collected rules"), ".",
				" If no arg is supplied, the current rule is removed.",
				"\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`repeatDetect([bool])`)}, {
				"Gets, and optionally sets, whether repeat detect instrument is on.", "\nReturns: ", code(`object`), " when repeat detect is enabled (or ", code(`undefined`), " when not enabled)",
				expandable{indent: 10, rightMargin: 10, border: true, title: "object properties", content: content{
					hanging{indent: 10, gap: 8, prefix: content{code(".Examined:int")}, content: "the number of steps examined"},
					hanging{indent: 10, gap: 8, prefix: content{code(".FirstStep:int")}, content: "the first step of the repeat"},
					hanging{indent: 10, gap: 8, prefix: content{code(".Found:bool")}, content: "whether a repeat was detected"},
					hanging{indent: 10, gap: 8, prefix: content{code(".Period:int")}, content: "the repeat period"},
					hanging{indent: 10, gap: 8, prefix: content{code(".RepeatStep:int")}, content: "the step at which repeated"},
				}},
			}},
			{{codeCopyable(`repeatInstrument()`)}, {
				"Returns the ", code(`RepeatInstrument`), " (or ", code(`undefined`), " if not enabled)",
				repeatInstrumentObject,
			}},
			{{codeCopyable(`replaySnapshot()`)}, {"Replays a snapshot. Resets the grid to the last snapshot (without removing the snapshot from the stack)\n", italic("Note: If there is no current snapshot - the grid is randomized and a snapshot added"), "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`rule([string|int|Rule])`)}, {"Gets, and optionally sets, the currently active rule.",
				"\nReturns: ", code(`Rule`), " - currently active rule",
				ruleObject,
			}},
			{{codeCopyable(`ruleCollected([string|int|Rule])`)}, {
				"Gets whether the supplied rule is a currently collected rule. If no arg is supplied, the current rule is used.",
				"\nReturns: ", code(`bool`), " - whether the rule is a collected rule",
			}},
			{{codeCopyable(`run()`)}, {"Runs simulation on the current grid.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`runRecipe([string|object])`)}, {
				"Runs a ", GridRecipes.link("grid recipe"), " on the grid. Where the recipe arg is:",
				hanging{indent: 10, prefix: content{code("string")}, content: content{" - is the filename for the recipe"}},
				hanging{indent: 10, prefix: content{code("object")}, content: content{" - is the definition of the recipe ", italic("(see "), GridRecipesReference.linkItalic("reference"), italic(")")}},
				"Returns: ", code(`bool`), " - whether the recipe was successfully run (check )", code(`[user]/Documents/GoGoL/errors.log`), " if unsuccessful).",
			}},
			{{codeCopyable(`savePatternImage(Pattern[,string[,object]])`)}, {
				"Saves the pattern as an image (.png).",
				"\nThe second, string, arg is the filename - if omitted, a default filename is used.",
				" The third, optional, arg is an object for the image metadata.",
				"\nReturns: none (", code(`undefined`), ")",
			}},
			{{codeCopyable(`sleep([int])`)}, {"Sleeps for a given ", italic("ms"), " (millisecond) period. If the arg is omitted, sleeps for 1 second.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`snapshot()`)}, {"Snapshots the current grid", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`status([msg:string])`)}, {"Sets a status ", italic(`msg`), " in the ", StatusBar.link("statusbar"), ". If the ", italic(`msg`), " arg is omitted, the status message is cleared.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`step()`)}, {"Performs a single simulation step on the current grid.", "\nReturns: ", code(`int`), " - the number of cells that changed state"}},
			{{codeCopyable(`stepCount()`)}, {"Gets the current step count.", "\nReturns: ", code(`int`)}},
			{{codeCopyable(`stepAhead([int])`)}, {"Performs step ahead. If no arg is supplied, the current step ahead size setting is used.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`stepAheadSize([int])`)}, {"Gets, and optionally sets, the current step ahead size (does not perform step ahead).", "\nReturns: ", code(`int`), " - current step ahead size (>0)"}},
			{{codeCopyable(`stepBack([int])`)}, {"Performs step back. If no arg is supplied, the current step back size setting is used.", italic("\n(only available when record instrument enabled)"), "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`stepBackSize([int])`)}, {"Gets, and optionally sets, the current step back size (does not perform step back).", "\nReturns: ", code(`int`), " - current step back size (>0)"}},
			{{codeCopyable(`stepDelay([int])`)}, {"Gets, and optionally sets, the current step delay (ms).", "\nReturns: ", code(`int`), " - current step delay (0-2000)"}},
			{{codeCopyable(`stop()`)}, {"Stops simulation on the current grid.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`undoToSnapshot()`)}, {"Restores grid to last snapshot.", "\nReturns: none (", code(`undefined`), ")"}},
			{{codeCopyable(`variable(string[,any])`)}, {"Gets, and optionally sets, a specific shortcut variable. The first string arg is the name of the variable. The second, optional, arg is the value to set for the variable (converted to string).", "\nReturns: ", code(`string`), " - the value of the variable (or ", code(`undefined`), " if the variable is unknown)"}},
			{{codeCopyable(`variables()`)}, {"Gets the currently defined shortcut variables.", "\nReturns: ", code(`object`), " - keys are variable names, values are each variable value (string)"}},
			{{codeCopyable(`wrapMode([string])`)}, {"Gets, and optionally sets, the current grid wrap mode.", "\nReturns: ", code(`string`), " - current wrap mode (", code(`none`), "|", code(`horizontal`), "|", code(`vertical`), "|", code(`all`), "|", code(`toroidal`), ")"}},
			{{codeCopyable(`zoom([int])`)}, {"Gets, and optionally sets, the current zoom factor.", "\nReturns: ", code(`int`), " - current zoom factor % (10-1000)"}},
		},
	},
	"Notes:",
	hanging{indent: 10, prefix: "• ", content: content{
		"By default, no exceptions are raised by functions - any errors or invalid args are logged to ", code(`[user]/Documents/GoGoL/errors.log`),
		"\n", italic("Note: To override this default behavior and make errors raise exceptions - use "), code("exceptions = true;"),
	}},
	hanging{indent: 10, prefix: "• ", content: content{"Most arg values are coerced - for example: ", code(`sleep("10");`), " is equivalent to ", code(`sleep(10);`)}},
	hanging{indent: 10, prefix: "• ", content: content{"Shortcuts executing javascript can still be interrupted by press key ", keys{"Esc"}, " - but the interruption only occurs at the next function call."}},
	expandable{padding: 4, border: true, spaceBefore: 8, title: content{bold("console")}, content: content{
		"A limited ", code(`console`), " object is also made available - providing the following methods:",
		table{
			borders:    true,
			padding:    4,
			spaceAfter: 4,
			columns: []tableColumn{
				{
					width:  50,
					header: "Method",
				},
				{
					header: "Description",
				},
			},
			rows: []tableRow{
				{{codeCopyable(`console.log(...vals:any)`)}, {"Logs the ", italic(`vals`), " to ", code(`[user]/Documents/GoGoL/output.log`), "\nReturns: none (", code(`undefined`), ")"}},
				{{codeCopyable(`console.error(...vals:any)`)}, {"Logs the ", italic(`vals`), " as an error to ", code(`[user]/Documents/GoGoL/errors.log`), "\nReturns: none (", code(`undefined`), ")"}},
				{{codeCopyable(`console.warn(...vals:any)`)}, {"Logs the ", italic(`vals`), " as a warning to ", code(`[user]/Documents/GoGoL/errors.log`), "\nReturns: none (", code(`undefined`), ")"}},
			},
		},
	}},
	expandable{padding: 4, border: true, spaceBefore: 8, title: content{bold("files")}, content: content{
		"The ", code(`files`), " object provides rudimentary file system access:",
		table{
			borders:    true,
			padding:    4,
			spaceAfter: 4,
			columns: []tableColumn{
				{
					width:  50,
					header: "Method",
				},
				{
					header: "Description",
				},
			},
			rows: []tableRow{
				{{codeCopyable(`files.append(path:string, data:string)`)}, {
					"Appends the string ", boldItalic("data"), " to the file specified by ", boldItalic("path"), ".",
					"\nIf the file does not exist, it is created. All directories to the path are created.",
					"\nReturns: ", code(`bool`), " - success (or ", code(`undefined`), " for error)",
				}},
				{{codeCopyable(`files.info(path:string)`)}, {
					"Obtains info about a file/directory specified by ", boldItalic("path"), ".",
					"\nReturns: ", code(`object`), "(or ", code(`undefined`), " if path does not exist)",
					expandable{indent: 10, rightMargin: 10, border: true, title: "object properties", content: content{
						hanging{indent: 10, gap: 8, prefix: content{code(".name:string")}, content: "the base file/directory name"},
						hanging{indent: 10, gap: 8, prefix: content{code(".path:string")}, content: "the parent path"},
						hanging{indent: 10, gap: 8, prefix: content{code(".size:int")}, content: "the file size"},
						hanging{indent: 10, gap: 8, prefix: content{code(".isDir:bool")}, content: "whether a directory"},
						hanging{indent: 10, gap: 8, prefix: content{code(".modTime:string")}, content: "datetime the file/directory was last modified"},
					}},
				}},
				{{codeCopyable(`files.read(path:string)`)}, {
					"Reads the contents of the file specified by ", boldItalic("path"), " as a string.",
					"\nReturns: ", code(`string`), " - file contents (or ", code(`undefined`), " if path does not exist or error)",
				}},
				{{codeCopyable(`files.walk(path:string, cb:func [,bool])`)}, {
					"Walks the directory specified by ", boldItalic("path"), ".",
					"\nThe ", boldItalic("cb"), " arg is a callback function that receives an ", code(`object`), " arg as the directory entry information.",
					expandable{spaceAfter: 4, indent: 10, rightMargin: 10, border: true, title: "object properties", content: content{
						hanging{indent: 10, gap: 8, prefix: content{code(".name:string")}, content: "the base file/directory name"},
						hanging{indent: 10, gap: 8, prefix: content{code(".path:string")}, content: "the parent path"},
						hanging{indent: 10, gap: 8, prefix: content{code(".isDir:bool")}, content: "whether the entry is a directory"},
					}},
					"The callback function can optionally return a ", code(`bool`), " where ", code(`true`), " indicates:",
					hanging{indent: 10, prefix: "• ", content: "on a directory entry - that directory is not to be walked down."},
					hanging{spaceAfter: 8, indent: 10, prefix: "• ", content: "on a file entry - the entire walk is ceased."},
					"The third, optional, ", code(`bool`), " arg specifies whether directory entries should reported (default is ", code(`true`), ").",
					" Specifying ", code(`false`), " means directory entries are not reported and not walked down.",
					"\nReturns: none (", code(`undefined`), ")",
				}},
				{{codeCopyable(`files.write(path:string, data:string)`)}, {
					"Writes the string ", boldItalic("data"), " to the file specified by ", boldItalic("path"), ".",
					"\nIf the file does not exist, it is created. All directories to the path are created.",
					"\nIf the file currently exists, it is truncated prior to write.",
					"\nReturns: ", code(`bool`), " - success (or ", code(`undefined`), " for error)",
				}},
			},
		},
		"Notes:",
		hanging{indent: 20, prefix: "• ", content: content{"all ", code(`path:string`), " args are either an absolute path - or taken as a relative path from ", code(`[user]/Documents/GoGoL/`)}},
		hanging{indent: 20, prefix: "• ", content: content{boldItalic("CAUTION"), italic(" access to the filesystem is unrestricted - write files with care!")}},
	}},
	expandable{padding: 4, border: true, spaceBefore: 8, title: content{bold("require")}, content: content{
		"The ", code(`require()`), " function is supported.",
		"\nExample:",
		indent{indent: 10, spaceAfter: 4, content: content{
			italic("Shortcut javascript:"),
			codeBlock{code: `const math = require('./mathUtil.js');
console.log("REQUIRE TEST", math.add(10, 6));`},
		}},
		indent{indent: 10, spaceAfter: 8, content: content{
			italic("File: "), code(`mathUtil.js`),
			codeBlock{code: `function add(a, b) {
	return a + b;
}

function subtract(a, b) {
	return a - b;
}

function multiply(a, b) {
	return a * b;
}

module.exports = {
	add,
	subtract,
	multiply
};`},
		}},
		italic("Note: Relative paths used in "), code(`require()`), italic(" are taken as relative to shortcut javascript file being executed."),
	}},
	expandable{padding: 4, border: true, rightMargin: 4, spaceBefore: 8, title: content{bold("Example - Save Animation")}, content: content{
		"The following javascript demonstrates adding a UI dialog to save the ", Instrumentation.link("record instrument"), " animation - allowing the user to define the start frame and maximum frames (which the standard UI does not support).",
		codeBlock{indent: 8, code: `exceptions = true;
// this function is used by the dialog onChange...
function onChange(res, dlg) {
	switch (res["name"]) {
		case "file":
			dlg.Hide("err"); // clear any displayed error when file input is changed
			// update the format radio buttons according to file extension...
			const path = res["value"];
			const ext = path.slice(path.lastIndexOf('.'));
			const curr = dlg.Get("format");
			if (ext === ".gif" && curr !== "Gif") {
				dlg.Set("format", "Gif");
			}
			if (ext === ".mp4" && curr != "Mp4") {
				dlg.Set("format", "Mp4");
			}
			break;
		case "start":
			// make sure start+max frames does not exceed available frames...
			available = dlg.Get("available");
			start = res["value"];
			frames = dlg.Get("frames");
			if ((start + frames) > available) {
				dlg.Set("frames", available - start);
			}
			break;
		case "frames":
			// make sure start+max frames does not exceed available frames...
			available = dlg.Get("available");
			start = dlg.Get("start");
			frames = res["value"];
			if ((start + frames) > available) {
				dlg.Set("frames", available - start);
			}
			break;
	}
}
// dynamic dialog controls...
let framesAvailable = {
	"type": "label",
	"label": "Frames available: 0",
	"bold": true
};
let framesStart = {
	"type": "int",
	"name": "start",
	"label": "Start frame:",
	"min": 0,
	"max": 10,
	"maxLen": 6,
	"initial": 0
};
let framesMax = {
	"type": "int",
	"name": "frames",
	"label": "Max. Frames:",
	"min": 0,
	"max": 10,
	"maxLen": 6,
	"initial": 0
};
// dialog definition...
const dialog = [
	{
		"type": "init",
		"func": function (res, dlg) {
			// initialize the dialog...
			const ri = recordInstrument();
			dlg.Set("available", ri.FramesCount());
			dlg.Set("frames", ri.FramesCount());
			dlg.Set("borders", borders());
			dlg.Set("cellSize", cellSize());
		},
		"onChange": onChange, // callback for dialog control changes
		"hide": ["err"]
	},
	{
		"type": "ok",
		"label": "Save",
		"onClick": function (res, dlg) {
			if (!res["file"]) {
				dlg.Set("err", "Please select a filename");
				dlg.Show("err");
				return true;
			}
		}
	},
	framesAvailable,
	{
		"name": "format",
		"type": "radio",
		"label": "Format:",
		"options": ["Gif", "Mp4"],
		"horizontal": true,
		"initial": 0,
	},
	framesStart,
	framesMax,
	{
		"name": "file",
		"type": "file",
		"label": "File:",
		"extensions": [".gif", ".mp4"],
		"allowDirs": true,
		"maxWidth": 30,
	},
	{
		// error display label...
		"name": "err",
		"type": "label",
		"label": "err",
		"color": "#ff0000",
	},
	{
		"type": "group",
		"label": "Overrides:",
		"horizontal": true,
		"items": [
			{
				"type": "int",
				"name": "cellSize",
				"label": " Cell size:",
				"min": 1,
				"max": 16,
				"maxLen": 3
			},
			{
				"type": "check",
				"name": "borders",
				"options": ["Show Borders"],
				"horizontal": true,
			}
		]
	}
];

let ri = recordInstrument();
if (ri) {
	if (ri.FramesCount() > 0) {
		let available = ri.FramesCount();
		// setup dialog controls according to record instrument...
		framesAvailable["label"] = "Frames available: "+available;
		framesStart["max"] = available-1;
		framesMax["max"] = available;
		// show the dialog and wait for response...
		const response = promptDialog("Save Animation", dialog);
		if (response) {
			// a response indicates the dialog wasn't cancelled
			status("Saving...");
			// save the overridden settings...
			const wasFmt = animationFormat();
			const wasCellSize = cellSize();
			const wasBorders = borders();
			try {
				if (response["format"] === "Mp4") {
					animationFormat("mp4");
				} else {
					animationFormat("gif");
				}
				borders(response["borders"]);
				cellSize(response["cellSize"]);
				animationSave(response["file"], response["start"], response["frames"]);
				status("Saved");
			} catch (e) {
				status("Save error (see log)");
			} finally {
				// restore overridden settings...
				animationFormat(wasFmt);
				cellSize(wasCellSize);
				borders(wasBorders);
			}
		}
	} else {
		status("No recorded steps");
	}
} else {
	status("Record not enabled");
}`},
	}},
}

var ruleObject = expandable{indent: 10, rightMargin: 10, border: true, title: "Rule object", content: content{
	"provides the following methods:",
	hanging{indent: 10, prefix: content{codeCopyable(".BornWith():string")}},
	hanging{indent: 10, prefix: content{codeCopyable(".IsCustom():bool")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Name():string")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Permutation():int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Rle():string")}},
	hanging{indent: 10, prefix: content{codeCopyable(".SurvivesWith():string")}},
}}

var patternObject = expandable{indent: 10, rightMargin: 10, border: true, title: "Pattern object", content: content{
	"provides the following (read-only) properties:",
	hanging{indent: 10, prefix: content{codeCopyable(".Cells:array[bool]")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Comments:array[string]")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Filename:string")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Height:int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Name:string")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Origination:string")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Rule:Rule")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Width:int")}},
	"and the following methods:",
	hanging{indent: 10, prefix: content{codeCopyable(".Reflected():Pattern")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Rle():string")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Rotated(r:int):Pattern")}},
	hanging{indent: 10, prefix: content{codeCopyable(".ShiftedDown():Pattern")}},
	hanging{indent: 10, prefix: content{codeCopyable(".ShiftedLeft():Pattern")}},
	hanging{indent: 10, prefix: content{codeCopyable(".ShiftedRight():Pattern")}},
	hanging{indent: 10, prefix: content{codeCopyable(".ShiftedUp():Pattern")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Trimmed():Pattern")}},
}}

var metaRuleObject = expandable{indent: 10, rightMargin: 10, border: true, title: "MetaRule object", content: content{
	"provides the following methods:",
	hanging{indent: 10, prefix: content{codeCopyable(".List():array[Rule]")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Matches(perm:int):bool")}},
	hanging{indent: 10, prefix: content{codeCopyable(".MatchesRule(Rule):bool")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Next(perm:int):int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Previous(perm:int):int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".String():string")}},
}}

var recordInstrumentObject = expandable{indent: 10, rightMargin: 10, border: true, title: "RecordInstrument object", content: content{
	"provides the following methods:",
	hanging{indent: 10, prefix: content{codeCopyable(".ChangeLocations(offset:int,count:int):[][][2]int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".FramesCount():int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".InitialGrid():[][]bool")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Reset()")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Undo()")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Undos(n:int)")}},
}}

var repeatInstrumentObject = expandable{indent: 10, rightMargin: 10, border: true, title: "RepeatInstrument object", content: content{
	"provides the following methods:",
	hanging{indent: 10, prefix: content{codeCopyable(".Examined():int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".FirstStep():int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Found():bool")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Period():int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".RepeatStep():int")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Reset()")}},
	hanging{indent: 10, prefix: content{codeCopyable(".Step():int")}},
}}

var dialogDefinition = expandable{indent: 10, rightMargin: 10, padding: 8, border: true, title: "Dialog definition array", content: content{
	"The dialog definition ia an array - where each element is either a ", code(`string`), " or an ", code(`object`), ".",
	"\nString elements are just displayed as a label in the dialog - object elements define the control to be shown:",
	indent{indent: 8, content: content{
		h4("• Input (string)"),
		"Example:",
		codeBlock{code: `{
	"type": "string",
	"name": "str", // used as key in response object
	"label": "Rule:", // optional
	"filter": "SB/012345678", // optional - characters allowed
	"maxLen": 18, // optional - max length for input
	"initial": "S2/B23" // optional - initial value
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Input (integer)"),
		"Example:",
		codeBlock{code: `{
	"type": "integer",
	"name": "age", // used as key in response object
	"label": "Age:", // optional
	"min": 0, // minimum value
	"max": 100, // maximum value
	"maxLen": 3, // optional - max length for input
	"initial": 18 // optional - initial value
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Check boxes"),
		"Example:",
		codeBlock{code: `{
	"type": "check",
	"name": "opt-c", // used as key in response object
	"label": "Select any:", // optional
	"options": ["Foo", "Bar", "Baz"],
	"horizontal": true, // optional - defaults to false
	"initial": 7 // optional - initial options checked
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Radio buttons"),
		"Example:",
		codeBlock{code: `{
	"type": "radio",
	"name": "opt-r", // used as key in response object
	"label": "Select one:", // optional
	"options": ["Foo", "Bar", "Baz"],
	"horizontal": true, // optional - defaults to false
	"initial": 0 // optional - initial option selected
}`},
	}},
	indent{indent: 8, content: content{
		h4("• List"),
		"Example:",
		codeBlock{code: `{
	"type": "list",
	"name": "opt-l", // used as key in response object
	"label": "Select:", // optional
	"options": ["Foo", "Bar", "Baz"],
	"initial": 0 // optional - initial option selected
	"maxLen": 3 // optional - max number of items to display
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Text"),
		"Example:",
		codeBlock{code: `{
	"type": "text",
	"name": "opt-t", // used as key in response object
	"label": "Some Text:", // optional
	"initial": "initial value" // optional - initial value
	"maxLines": 3 // optional - max number of lines to display
	"maxWidth": 20 // optional - max width (in ems) for input display
}`},
	}},
	indent{indent: 8, content: content{
		h4("• File"),
		"Example:",
		codeBlock{code: `{
	"type": "file",
	"name": "file", // used as key in response object
	"label": "File:", // optional
	"maxWidth": 18, // optional - max display width for input
	"extensions": [".rle"], // optional - allowed file extensions 
	"initial": "glider.rle" // optional - initial value
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Label"),
		"Example:",
		codeBlock{code: `{
	"type": "label",
	"name": "lbl-1", // optional - not shown in response but allows addressing
	"label": "My label",
	"small": true, // optional - defaults to false
	"italic": true, // optional - defaults to false
	"bold": true, // optional - defaults to false
	"color": "#ff0000" // optional
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Button"),
		"Example:",
		codeBlock{code: `{
	"type": "button",
	"label": "My Button",
	"onClick": function(data:object,dlg:object) // optional
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Buttons"),
		"Example:",
		codeBlock{code: `{
	"type": "buttons",
	"name": "btns-1", // optional - not shown in response but allows addressing
	"label": "Button Group:", // optional
	"options": ["Foo", "Bar", "Baz"], // the actual buttons
	"onClick": function(data:object,dlg:object) // optional
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Group"),
		"Example:",
		codeBlock{code: `{
	"type": "group",
	"name": "lbl-1", // optional - not shown in response but allows addressing
	"label": "Group", // optional
	"horizontal": false, // optional - defaults to false
	"items": [...], // array of dialog items
}`},
	}},
	indent{indent: 8, content: content{
		h4("• Re-label Ok button"),
		"Example:",
		codeBlock{code: `{
	"type": "ok",
	"label": "Okay", // optional
	"onClick": function(data:object,dlg:object) // optional
}`},
		italic("Note: If the "), code(`onClick`), italic(" property is provided - the func can return a boolean, where "), code(`true`), italic(" indicates the dialog should "), boldItalic("not"), italic(" be closed."),
		italic(" This can be useful for validating the dialog data prior to closing."),
	}},
	indent{indent: 8, content: content{
		h4("• Re-label Cancel button"),
		"Example:",
		codeBlock{code: `{
	"type": "cancel",
	"label": "Nope", // optional
	"onClick": function(data:object,dlg:object) // optional
}`},
	}},
	indent{indent: 8, spaceAfter: 4, content: content{
		h4("• Initialise"),
		"Example:",
		codeBlock{code: `{
	"type": "init",
	"hide": ["ctrl-1"], // optional - initially hidden controls
	"func": function(data:object,dlg:object), // optional
	"onChange": function(data:object,dlg:object) // optional
}`},
		"Notes:",
		hanging{indent: 8, gap: 8, prefix: "•", content: content{
			"the optional ", code(`func`), " property is a callback that is called immediately before the dialog is shown.",
		}},
		hanging{indent: 8, gap: 8, prefix: "•", content: content{
			"the optional ", code(`onChange`), " property is a callback that is called each time a control value is changed.",
			" Where the ", code(`data`), " arg passed to the callback is not the full dialog values - it only contains two properties: ",
			code(`name`), " & ", code(`value`), ".",
		}},
	}},
	separator{},
	indent{indent: 8, spaceBefore: 4, content: content{
		h3("Callbacks"),
		"For dialog items that support ", code(`onClick`), " (or ", code(`init.func`), ")",
		" the function is a callback with the following signature:",
		codeBlock{indent: 8, code: `function(data:object,dlg:object)`},
		"where args are:",
		hanging{indent: 8, gap: 8, prefix: content{code(`data:object`)}, content: content{"is the current dialog data - keys are the control names and values are that control's current value."}},
		hanging{indent: 8, gap: 8, prefix: content{code(`dlg:object`)}, content: content{
			"is the dialog itself - supporting the following methods: ",
			hanging{gap: 8, prefix: content{codeCopyable(`.Cancel()`)}, content: content{}},
			hanging{gap: 8, prefix: content{codeCopyable(`.Get(name:string):any`)}, content: content{}},
			hanging{gap: 8, prefix: content{codeCopyable(`.Hide(names:string...)`)}, content: content{}},
			hanging{gap: 8, prefix: content{codeCopyable(`.IsHidden(name:string):bool`)}, content: content{}},
			hanging{gap: 8, prefix: content{codeCopyable(`.Ok()`)}, content: content{}},
			hanging{gap: 8, prefix: content{codeCopyable(`.Set(name:string, value:any)`)}, content: content{}},
			hanging{gap: 8, prefix: content{codeCopyable(`.Show(names:string...)`)}, content: content{}},
		}},
	}},
}}
