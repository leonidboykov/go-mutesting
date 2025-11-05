package importing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseAndTypeCheckFileTypeCheckWholePackage(t *testing.T) {
	_, _, err := ParseAndTypeCheckFile("../astutil/create.go")
	assert.Nil(t, err)
}
