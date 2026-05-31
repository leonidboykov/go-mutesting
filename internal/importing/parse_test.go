package importing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAndTypeCheckFileTypeCheckWholePackage(t *testing.T) {
	_, _, err := ParseAndTypeCheckFile(t.Context(), "../astutil/create.go")
	assert.Nil(t, err)
}
