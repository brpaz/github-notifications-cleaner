package version_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/github-notifications-cleaner/cmd/version"
)

func TestVersionCmd(t *testing.T) {
	cmd := version.NewCmd()
	assert.NotNil(t, cmd)

	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	err := cmd.Execute()
	assert.NoError(t, err)

	out, err := io.ReadAll(b)
	assert.NoError(t, err)
	assert.Contains(t, string(out), "Build date:")
}
