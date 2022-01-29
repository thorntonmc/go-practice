package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var goodJSONString = `{"name":"michael"}`
var badJSONString = `{"name":"michael","address": "1234 Shady Lane Boston, MA"}`

func TestMarshal(t *testing.T) {
	var err error
	tj := &testJSON{}

	err = marshal([]byte(goodJSONString), tj)
	assert.NoError(t, err)
	assert.Equal(t, "michael", tj.Name)

	err = marshal([]byte(badJSONString), tj)
	assert.Error(t, err)
}
