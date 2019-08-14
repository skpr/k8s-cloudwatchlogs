package cloudwatchlogs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGroupName(t *testing.T) {
	assert.Equal(t, "prefix-name", getGroupName("prefix", "name"))
	assert.Equal(t, "name", getGroupName("", "name"))
}
