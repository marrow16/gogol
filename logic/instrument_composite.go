package logic

type CompositeInstrument []DualUseInstrumentation

var _ StepInstrumentation = (CompositeInstrument)(nil)
var _ StepStopInstrumentation = (CompositeInstrument)(nil)
var _ DualUseInstrumentation = (CompositeInstrument)(nil)

func (c CompositeInstrument) Instrument(step uint64, locations []int) {
	for _, inst := range c {
		inst.Instrument(step, locations)
	}
}

func (c CompositeInstrument) InstrumentStop(step uint64, locations []int) bool {
	for _, inst := range c {
		if inst.InstrumentStop(step, locations) {
			return true
		}
	}
	return false
}
