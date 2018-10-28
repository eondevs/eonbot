package strategy

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTool(t *testing.T) {
	tl, err := newToolFromJSON("test1", testTool, json.RawMessage(`{"count":10, "condsMet": true}`))
	res := &Tool{
		ID:            "test1",
		Type:          testTool,
		RawProperties: json.RawMessage(`{"count":10, "condsMet": true}`),
		Properties: &toolPropertiesMock{
			conf: toolPropertiesMockSettings{
				Count:    10,
				CondsMet: true,
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, res, tl)

	_, err = newToolFromJSON("test1", testTool, json.RawMessage(`{`))
	assert.NotNil(t, err)
}

func TestToolClone(t *testing.T) {
	tool := &Tool{
		ID:            "test1",
		Type:          testTool,
		RawProperties: json.RawMessage(`{"count":10, "condsMet": true}`),
		Properties: &toolPropertiesMock{
			conf: toolPropertiesMockSettings{
				Count:    10,
				CondsMet: true,
			},
		},
	}

	newTool, err := tool.clone()
	assert.Nil(t, err)
	assert.Equal(t, tool, newTool)

	newTool1, err := newTool.clone()
	assert.Nil(t, err)
	newTool1.ID = "test123"
	assert.NotEqual(t, newTool, newTool1)
}
