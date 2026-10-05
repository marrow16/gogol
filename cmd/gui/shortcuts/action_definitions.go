package shortcuts

type Action int

const (
	Unknown Action = iota
	NoAction
	// grid settings...
	BoundaryMode
	CellSize
	GridHeight
	GridSize
	GridWidth
	StepDelay
	WrapMode
	// grid colors...
	Borders
	CellColorAlive
	CellColorBorder
	CellColorDead
	// grid drawing...
	Clear
	MaxAdjacents
	RandomAdditions
	RandomChanges
	RandomCull
	Randomization
	Randomize
	RandomizePopulation
	// repeat detection...
	RepeatDetect
	RepeatDetectSave
	// recording...
	AnimationFormat
	AnimationSave
	Record
	// heat mapping...
	HeatMap
	HeatMapColors
	HeatMapReveal
	HeatMapSave
	// snapshots...
	ReplaySnapshot
	Snapshot
	UndoToSnapshot
	// running/stepping...
	Run
	Step
	StepBack
	StepAhead
	StepAheadBy
	StepBackBy
	Stop
	// rule settings...
	BornWith
	Rule
	RuleInt
	RulePerm
	SurvivesWith
	// meta rules...
	IterateMetaRule
	NextMetaRule
	PreviousMetaRule
	// collected rules...
	AddCollectedRule
	IterateCollectedRules
	NextCollectedRule
	PreviousCollectedRule
	RemoveCollectedRule
	// misc...
	CollectFiles
	Export
	ExportImage
	Log
	Status
	Name
	NameReset
	RunRecipe
	Sleep
	// flow control...
	If
	While
	BreakIf
	StopIf
	Call
	Repeat
	Variable
	ClearVariable
	ClearVariables
	maxAction
)

func (a Action) String() string {
	switch a {
	case BoundaryMode:
		return "boundary-mode"
	case CellSize:
		return "cell-size"
	case GridHeight:
		return "grid-height"
	case GridSize:
		return "grid-size"
	case GridWidth:
		return "grid-width"
	case StepDelay:
		return "step-delay"
	case WrapMode:
		return "wrap-mode"
	case Borders:
		return "borders"
	case CellColorAlive:
		return "cell-color-alive"
	case CellColorBorder:
		return "cell-color-border"
	case CellColorDead:
		return "cell-color-dead"
	case Clear:
		return "clear"
	case MaxAdjacents:
		return "max-adjacents"
	case RandomAdditions:
		return "random-additions"
	case RandomChanges:
		return "random-changes"
	case RandomCull:
		return "random-cull"
	case Randomization:
		return "randomization"
	case Randomize:
		return "randomize"
	case RandomizePopulation:
		return "randomize-population"
	case RepeatDetect:
		return "repeat-detect"
	case RepeatDetectSave:
		return "repeat-detect-save"
	case AnimationFormat:
		return "record-animation-format"
	case AnimationSave:
		return "record-animation-save"
	case Record:
		return "record"
	case HeatMap:
		return "heat-map"
	case HeatMapColors:
		return "heat-map-colors"
	case HeatMapReveal:
		return "heat-map-reveal"
	case HeatMapSave:
		return "heat-map-save"
	case ReplaySnapshot:
		return "replay-snapshot"
	case Snapshot:
		return "snapshot"
	case UndoToSnapshot:
		return "undo-to-snapshot"
	case Run:
		return "run"
	case Step:
		return "step"
	case StepBack:
		return "step-back"
	case StepAhead:
		return "step-ahead"
	case StepAheadBy:
		return "step-ahead-by"
	case StepBackBy:
		return "step-back-by"
	case Stop:
		return "stop"
	case BornWith:
		return "rule-born-with"
	case Rule:
		return "rule"
	case RuleInt:
		return "rule-int"
	case RulePerm:
		return "rule-perm"
	case SurvivesWith:
		return "rule-survives-with"
	case IterateMetaRule:
		return "iterate-meta-rule"
	case NextMetaRule:
		return "next-meta-rule"
	case PreviousMetaRule:
		return "previous-meta-rule"
	case AddCollectedRule:
		return "add-collected-rule"
	case IterateCollectedRules:
		return "iterate-collected-rules"
	case NextCollectedRule:
		return "next-collected-rule"
	case PreviousCollectedRule:
		return "previous-collected-rule"
	case RemoveCollectedRule:
		return "remove-collected-rule"
	case CollectFiles:
		return "collect-files"
	case Export:
		return "export"
	case ExportImage:
		return "export-image"
	case Log:
		return "log"
	case Status:
		return "status"
	case Name:
		return "name"
	case NameReset:
		return "name-reset"
	case RunRecipe:
		return "run-recipe"
	case Sleep:
		return "sleep"
	case If:
		return "if"
	case While:
		return "while"
	case BreakIf:
		return "break-if"
	case StopIf:
		return "stop-if"
	case Call:
		return "call"
	case Repeat:
		return "repeat"
	case Variable:
		return "variable"
	case ClearVariable:
		return "clear-variable"
	case ClearVariables:
		return "clear-variables"
	}
	return ""
}

func (a Action) Syntax() string {
	syntax := a.String()
	if af, ok := actionsRegistry[syntax]; ok && af.operand != none {
		if af.operandSyntax != "" {
			return syntax + ":" + af.operandSyntax
		} else {
			return syntax + ":operand"
		}
	}
	return syntax
}

// AltSyntax
// n = 0 - default
// n = 1 - no operand
// n = 2 - inc
// n = -2 - dec
func (a Action) AltSyntax(n int) string {
	if n == 0 {
		return a.Syntax()
	}
	syntax := a.String()
	if af, ok := actionsRegistry[syntax]; ok {
		switch {
		case n == 2 && af.operand.incDecAllowed():
			return syntax + "++"
		case n == -2 && af.operand.incDecAllowed():
			return syntax + "--"
		}
	}
	return syntax
}

type operandMode int

const (
	none operandMode = iota
	required
	requiredIncDec
	optional
	optionalIncDec
)

func (o operandMode) required() bool {
	return o == required || o == requiredIncDec
}

func (o operandMode) incDecAllowed() bool {
	return o == requiredIncDec || o == optionalIncDec
}

type registryItem struct {
	action        Action
	operand       operandMode
	operandSyntax string
	subActions    bool
}

var actionsRegistry = map[string]registryItem{
	Repeat.String(): {
		action:        Repeat,
		operand:       required,
		subActions:    true,
		operandSyntax: "n,action,...",
	},
	Name.String(): {
		action:        Name,
		operand:       required,
		operandSyntax: "format",
	},
	NameReset.String(): {
		action:  NameReset,
		operand: required,
	},
	CollectFiles.String(): {
		action: CollectFiles,
	},
	Run.String(): {
		action: Run,
	},
	Stop.String(): {
		action: Stop,
	},
	Export.String(): {
		action: Export,
	},
	ExportImage.String(): {
		action:        ExportImage,
		operand:       optional,
		operandSyntax: "metadata",
	},
	Clear.String(): {
		action: Clear,
	},
	Snapshot.String(): {
		action: Snapshot,
	},
	UndoToSnapshot.String(): {
		action: UndoToSnapshot,
	},
	ReplaySnapshot.String(): {
		action: ReplaySnapshot,
	},
	Step.String(): {
		action: Step,
	},
	StepBack.String(): {
		action: StepBack,
	},
	StepAhead.String(): {
		action:        StepAhead,
		operand:       optionalIncDec,
		operandSyntax: "n",
	},
	Randomize.String(): {
		action:        Randomize,
		operand:       optionalIncDec,
		operandSyntax: "n",
	},
	Randomization.String(): {
		action:        Randomization,
		operand:       requiredIncDec,
		operandSyntax: "n",
	},
	RandomChanges.String(): {
		action:        RandomChanges,
		operand:       optional,
		operandSyntax: "n",
	},
	RandomAdditions.String(): {
		action:        RandomAdditions,
		operand:       optional,
		operandSyntax: "n",
	},
	RandomCull.String(): {
		action:        RandomCull,
		operand:       optional,
		operandSyntax: "n",
	},
	RandomizePopulation.String(): {
		action:        RandomizePopulation,
		operand:       optional,
		operandSyntax: "n",
	},
	MaxAdjacents.String(): {
		action:        MaxAdjacents,
		operand:       required,
		operandSyntax: "n",
	},
	StepDelay.String(): {
		action:        StepDelay,
		operand:       requiredIncDec,
		operandSyntax: "ms",
	},
	RulePerm.String(): {
		action:        RulePerm,
		operand:       requiredIncDec,
		operandSyntax: "n",
	},
	RuleInt.String(): {
		action:        RuleInt,
		operand:       requiredIncDec,
		operandSyntax: "n",
	},
	Sleep.String(): {
		action:        Sleep,
		operand:       optional,
		operandSyntax: "ms",
	},
	RunRecipe.String(): {
		action:        RunRecipe,
		operand:       optional,
		operandSyntax: "filename",
	},
	WrapMode.String(): {
		action:        WrapMode,
		operand:       required,
		operandSyntax: "mode",
	},
	BoundaryMode.String(): {
		action:        BoundaryMode,
		operand:       required,
		operandSyntax: "mode",
	},
	StepAheadBy.String(): {
		action:        StepAheadBy,
		operand:       required,
		operandSyntax: "n",
	},
	StepBackBy.String(): {
		action:        StepBackBy,
		operand:       required,
		operandSyntax: "n",
	},
	Rule.String(): {
		action:        Rule,
		operand:       required,
		operandSyntax: "name|rle",
	},
	BornWith.String(): {
		action:        BornWith,
		operand:       requiredIncDec,
		operandSyntax: "x",
	},
	SurvivesWith.String(): {
		action:        SurvivesWith,
		operand:       requiredIncDec,
		operandSyntax: "x",
	},
	GridWidth.String(): {
		action:        GridWidth,
		operand:       required,
		operandSyntax: "n",
	},
	GridHeight.String(): {
		action:        GridHeight,
		operand:       required,
		operandSyntax: "n",
	},
	GridSize.String(): {
		action:        GridSize,
		operand:       required,
		operandSyntax: "wXh",
	},
	Record.String(): {
		action:        Record,
		operand:       optional,
		operandSyntax: "bool",
	},
	RepeatDetect.String(): {
		action:        RepeatDetect,
		operand:       optional,
		operandSyntax: "bool",
	},
	RepeatDetectSave.String(): {
		action: RepeatDetectSave,
	},
	HeatMap.String(): {
		action:        HeatMap,
		operand:       optional,
		operandSyntax: "type",
	},
	HeatMapColors.String(): {
		action:        HeatMapColors,
		operand:       required,
		operandSyntax: "colors",
	},
	HeatMapSave.String(): {
		action:        HeatMapSave,
		operand:       optional,
		operandSyntax: "metadata",
	},
	HeatMapReveal.String(): {
		action: HeatMapReveal,
	},
	NextMetaRule.String(): {
		action:        NextMetaRule,
		operand:       required,
		operandSyntax: "meta",
	},
	PreviousMetaRule.String(): {
		action:        PreviousMetaRule,
		operand:       required,
		operandSyntax: "meta",
	},
	IterateMetaRule.String(): {
		action:        IterateMetaRule,
		operand:       required,
		subActions:    true,
		operandSyntax: "meta,action,...",
	},
	Log.String(): {
		action:        Log,
		operand:       required,
		operandSyntax: "message",
	},
	Status.String(): {
		action:        Status,
		operand:       optional,
		operandSyntax: "msg",
	},
	AddCollectedRule.String(): {
		action: AddCollectedRule,
	},
	RemoveCollectedRule.String(): {
		action: RemoveCollectedRule,
	},
	PreviousCollectedRule.String(): {
		action: PreviousCollectedRule,
	},
	NextCollectedRule.String(): {
		action: NextCollectedRule,
	},
	IterateCollectedRules.String(): {
		action:        IterateCollectedRules,
		operand:       required,
		subActions:    true,
		operandSyntax: "action,...",
	},
	Borders.String(): {
		action:        Borders,
		operand:       required,
		operandSyntax: "bool",
	},
	CellSize.String(): {
		action:        CellSize,
		operand:       required,
		operandSyntax: "n",
	},
	CellColorAlive.String(): {
		action:        CellColorAlive,
		operand:       required,
		operandSyntax: "color",
	},
	CellColorDead.String(): {
		action:        CellColorDead,
		operand:       required,
		operandSyntax: "color",
	},
	CellColorBorder.String(): {
		action:        CellColorBorder,
		operand:       required,
		operandSyntax: "color",
	},
	AnimationSave.String(): {
		action: AnimationSave,
	},
	AnimationFormat.String(): {
		action:        AnimationFormat,
		operand:       required,
		operandSyntax: "gif|mp4",
	},
	If.String(): {
		action:        If,
		operand:       required,
		subActions:    true,
		operandSyntax: "condition,action,...",
	},
	While.String(): {
		action:        While,
		operand:       required,
		subActions:    true,
		operandSyntax: "condition,action,...",
	},
	BreakIf.String(): {
		action:        BreakIf,
		operand:       required,
		operandSyntax: "condition",
	},
	StopIf.String(): {
		action:        StopIf,
		operand:       required,
		operandSyntax: "condition",
	},
	"call-shortcut": {
		action:        Call,
		operand:       required,
		operandSyntax: "name",
	},
	Call.String(): {
		action:        Call,
		operand:       required,
		operandSyntax: "name",
	},
	Variable.String(): {
		action:        Variable,
		operand:       required,
		operandSyntax: "name[op]",
	},
	"var": {
		action:        Variable,
		operand:       required,
		operandSyntax: "name[op]",
	},
	ClearVariable.String(): {
		action:        ClearVariable,
		operand:       required,
		operandSyntax: "name",
	},
	"clear-var": {
		action:        ClearVariable,
		operand:       required,
		operandSyntax: "name",
	},
	ClearVariables.String(): {
		action:        ClearVariables,
		operand:       optional,
		operandSyntax: "name,...",
	},
	"clear-vars": {
		action:        ClearVariables,
		operand:       optional,
		operandSyntax: "name,...",
	},
}
