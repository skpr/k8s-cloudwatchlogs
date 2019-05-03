package metadata

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFile is used for mocking os.FileInfo in the test suite.
type TestFile struct {
	name      string
	directory bool
	os.FileInfo
}

// Name mock.
func (f TestFile) Name() string {
	return f.name
}

// IsDir mock.
func (f TestFile) IsDir() bool {
	return f.directory
}

func TestExtract(t *testing.T) {
	file := TestFile{
		name:      "POD_NAMESPACE_CONTAINER-ID.log",
		directory: false,
	}

	namespace, pod, container, err := Extract(file)
	assert.Nil(t, err)

	assert.Equal(t, "NAMESPACE", namespace)
	assert.Equal(t, "POD", pod)
	assert.Equal(t, "CONTAINER", container)
}

func TestValidate(t *testing.T) {
	// We don't allow directories.
	assert.False(t, Validate(TestFile{
		name:      "POD_NAMESPACE_CONTAINER-ID.log",
		directory: true,
	}))

	// We want a specific format.
	assert.False(t, Validate(TestFile{
		name:      "POD_NAMESPACE_CONTAINER_____ID.log",
		directory: false,
	}))
}
