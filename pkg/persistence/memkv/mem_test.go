package memkv_test

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/openyard/eventstore/pkg/persistence/memkv"
	"github.com/stretchr/testify/assert"
)

func TestNewMemoryKVS(t *testing.T) {
	sut := memkv.NewMemoryKVS("_index", "_meta", "_content")

	v, err := sut.Get("_index", "foo")
	assert.Zero(t, binary.BigEndian.Uint64(v))
	assert.Equal(t, fmt.Errorf("key (_index:foo) not found"), err)

	v, err = sut.Get("foo", "bar")
	assert.Zero(t, binary.BigEndian.Uint64(v))
	assert.Equal(t, fmt.Errorf("bucket (foo) not found"), err)
	err = sut.Put("foo", "bar", []byte("baz"))
	assert.Equal(t, fmt.Errorf("bucket (foo) not found"), err)

	assert.NoError(t, sut.Put("_index", "foo", []byte("bar")))
	v, err = sut.Get("_index", "foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", string(v))
}
