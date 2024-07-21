package anyi

import (
	"testing"

	"github.com/jieliu2000/anyi/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestNewFlow(t *testing.T) {

	client := test.MockClient{}
	flow := NewFlow(&client, "flow1")

	assert.NotNil(t, flow)
	assert.Equal(t, "flow1", flow.Name)
	assert.Equal(t, &client, flow.clientImpl)

}
