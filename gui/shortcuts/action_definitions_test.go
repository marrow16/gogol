package shortcuts

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestActionsRegistered(t *testing.T) {
	for action := NoAction + 1; action < maxAction; action++ {
		name := action.String()
		assert.NotEmpty(t, name, "action %d has no name", action)
		ar, ok := actionsRegistry[name]
		assert.True(t, ok, "action %q is not registered", name)
		assert.Equal(t, action, ar.action)
	}
}
