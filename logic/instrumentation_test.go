package logic

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStopReason_String(t *testing.T) {
	assert.Equal(t, "Step Completed", StepCompleted.String())
	assert.Equal(t, "No Changes Detected", NoChangesDetected.String())
	assert.Equal(t, "Instrument Stopped", InstrumentStopped.String())
	assert.Equal(t, "Unknown", StopReason(-1).String())
}
