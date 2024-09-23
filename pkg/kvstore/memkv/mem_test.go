package memkv_test

import (
	"fmt"
	"testing"

	"github.com/openyard/eventstore/pkg/kvstore/memkv"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryKVS(t *testing.T) {
	sut := memkv.NewMemoryKVS("_index", "_meta", "_content")
	assert.NoError(t, sut.AssertBucket("_index"))
	assert.NoError(t, sut.AssertBucket("_meta"))
	assert.NoError(t, sut.AssertBucket("_content"))

	v, err := sut.Get("_index", "foo")
	assert.Empty(t, v)
	assert.Equal(t, fmt.Errorf("key (_index:foo) not found"), err)

	v, err = sut.Get("foo", "bar")
	assert.Empty(t, v)
	assert.Equal(t, fmt.Errorf("bucket (foo) not found"), err)
	err = sut.Put("foo", "bar", []byte("baz"))
	assert.Equal(t, fmt.Errorf("bucket (foo) not found"), err)

	assert.NoError(t, sut.Put("_index", "foo", []byte("bar")))
	v, err = sut.Get("_index", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", string(v))
}
