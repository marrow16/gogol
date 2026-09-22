package logic

type StepInstrumentation interface {
	Instrument(step uint64, locations []int)
}

type StepStopInstrumentation interface {
	InstrumentStop(step uint64, locations []int) bool
}

type DualUseInstrumentation interface {
	StepInstrumentation
	StepStopInstrumentation
}

type StopReason int

const (
	StepCompleted StopReason = iota
	NoChangesDetected
	InstrumentStopped
)

func (r StopReason) String() string {
	switch r {
	case StepCompleted:
		return "Step Completed"
	case NoChangesDetected:
		return "No Changes Detected"
	case InstrumentStopped:
		return "Instrument Stopped"
	}
	return "Unknown"
}
