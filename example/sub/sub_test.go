package sub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaz(t *testing.T) {
	assert.Equal(t, baz(), 2)
}
