package recursivegenerationwithsubpkgexclude_test

import (
	"os"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubpkg2NotExist(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	subpkg2MockFile := pathlib.NewPath(wd).Join("subpkg2", "mocks.go")
	exists, err := subpkg2MockFile.Exists()
	require.NoError(t, err)
	assert.False(t, exists, "subpkg2 mocks.go file exists when it shouldn't")
}

func TestSubpkg1Exists(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	subpkg2MockFile := pathlib.NewPath(wd).Join("subpkg1", "mocks.go")
	exists, err := subpkg2MockFile.Exists()
	require.NoError(t, err)
	assert.True(t, exists, "subpkg1 mocks.go file doesn't exist when it should")
}
