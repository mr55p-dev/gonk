package gonk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTag(t *testing.T) {
	assert := assert.New(t)
	config := "path.segment.key"
	tag := parseConfigTag(config)
	assert.Equal("key", tag.Key())
	assert.Equal([]any{"path", "segment", "key"}, tag.path)
	assert.False(tag.options.optional)
}

func TestParseTagOptional(t *testing.T) {
	assert := assert.New(t)
	config := "path.segment.key,optional"
	tag := parseConfigTag(config)
	assert.Equal("key", tag.Key())
	assert.Equal([]any{"path", "segment", "key"}, tag.path)
	assert.True(tag.options.optional)
}
