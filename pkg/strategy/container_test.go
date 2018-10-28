package strategy

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerTypesChecks(t *testing.T) {
	// sequence type
	cont := container{seq: &sequence{}}
	assert.True(t, cont.isSeq())
	assert.False(t, cont.isTool())
	assert.False(t, cont.isUndefined())
	assert.False(t, cont.isBoth())

	// tool type
	cont = container{tool: &Tool{}}
	assert.True(t, cont.isTool())
	assert.False(t, cont.isSeq())
	assert.False(t, cont.isUndefined())
	assert.False(t, cont.isBoth())

	// undefined type
	cont = container{}
	assert.True(t, cont.isUndefined())
	assert.NotNil(t, cont.isUndefinedErr())
	assert.False(t, cont.isTool())
	assert.False(t, cont.isSeq())
	assert.False(t, cont.isBoth())

	// both types
	cont = container{tool: &Tool{}, seq: &sequence{}}
	assert.True(t, cont.isTool())
	assert.True(t, cont.isSeq())
	assert.True(t, cont.isBoth())
	assert.NotNil(t, cont.isBothErr())
	assert.False(t, cont.isUndefined())
}

func TestContainerClone(t *testing.T) {
	cont := &container{seq: &sequence{
		elems: []*seqElem{
			{
				cont: &container{
					tool: &Tool{
						Type: "test",
						RawProperties: []byte(`{
                            "count":10
                        }`),
						Properties: &toolPropertiesMock{
							conf: toolPropertiesMockSettings{
								Count: 10,
							},
						},
					},
				},
			},
		},
	},
		tool: &Tool{
			Type: "test",
			RawProperties: []byte(`{
                            "count":10
                        }`),
			Properties: &toolPropertiesMock{
				conf: toolPropertiesMockSettings{
					Count: 10,
				},
			},
		},
	}

	newCont, err := cont.clone()
	assert.Nil(t, err)
	assert.Equal(t, cont, newCont)

	newCont1, err := newCont.clone()
	newCont1.tool.Type = "test123"
	assert.Nil(t, err)
	assert.NotEqual(t, newCont, newCont1)
}

func TestContainerCandlesCount(t *testing.T) {
	// tool type
	cont := container{tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{Count: 5}}}}
	assert.Equal(t, 5, cont.candlesCount())

	// sequence type
	cont = container{seq: &sequence{
		elems: []*seqElem{
			{
				cont: &container{
					tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{Count: 5}}},
				},
			},
			{
				cont: &container{
					tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{Count: 7}}},
				},
			},
		},
	}}
	assert.Equal(t, 7, cont.candlesCount())

	// undefined type
	cont = container{}
	assert.Equal(t, 0, cont.candlesCount())
}

func TestContainerValidate(t *testing.T) {
	// undefined type
	cont := container{}
	assert.NotNil(t, cont.validate())

	// both types
	cont = container{seq: &sequence{}, tool: &Tool{}}
	assert.NotNil(t, cont.validate())

	// sequence type
	cont = container{
		seq: &sequence{
			elems: []*seqElem{
				{
					cont: &container{
						tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{Err: "test err"}}},
					},
				},
			},
		},
	}
	assert.NotNil(t, cont.validate())

	// non erroring sequence
	cont = container{
		seq: &sequence{
			elems: []*seqElem{
				{
					cont: &container{
						tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{}}},
					},
				},
			},
		},
	}
	assert.Nil(t, cont.validate())

	// tool type
	cont = container{tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{Err: "test err"}}}}
	assert.NotNil(t, cont.validate())
}

func TestContainerSnapshot(t *testing.T) {
	cont := &container{
		seq: &sequence{
			elems: []*seqElem{
				{
					cont: &container{
						tool: &Tool{
							Type: testTool,
							ID:   "test1",
							Properties: &toolPropertiesMock{
								conf: toolPropertiesMockSettings{
									Snap: tools.Snapshot{
										CondsMet: true,
										Data:     123,
									},
								},
							},
						},
					},
				},
				{
					cont: &container{
						tool: &Tool{
							Type: testTool,
							ID:   "test2",
							Properties: &toolPropertiesMock{
								conf: toolPropertiesMockSettings{
									Snap: tools.Snapshot{
										CondsMet: true,
										Data:     123,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	res := map[string]tools.FullSnapshot{
		"test1": tools.FullSnapshot{
			Type: testTool,
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data:     123,
			},
		},
		"test2": tools.FullSnapshot{
			Type: testTool,
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data:     123,
			},
		},
	}

	assert.Equal(t, res, cont.snapshot())
}

func TestContainerConditionsMet(t *testing.T) {
	// undefined type
	cont := container{}
	ok, err := cont.conditionsMet(exchange.Data{})
	assert.False(t, ok)
	assert.NotNil(t, err)

	// both types
	cont = container{seq: &sequence{}, tool: &Tool{}}
	ok, err = cont.conditionsMet(exchange.Data{})
	assert.False(t, ok)
	assert.NotNil(t, err)

	// sequence type
	cont = container{
		seq: &sequence{
			elems: []*seqElem{
				{
					cont: &container{
						tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{CondsMet: true}}},
					},
				},
			},
		},
	}
	ok, err = cont.conditionsMet(exchange.Data{})
	assert.True(t, ok)
	assert.Nil(t, err)

	// tool type
	cont = container{tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{CondsMet: true}}}}
	ok, err = cont.conditionsMet(exchange.Data{})
	assert.True(t, ok)
	assert.Nil(t, err)
}

func TestContainerReset(t *testing.T) {
	// sequence type
	cont := container{
		seq: &sequence{
			elems: []*seqElem{
				{
					cont: &container{
						tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{}}},
					},
				},
				{
					cont: &container{
						tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{}}},
					},
				},
			},
		},
	}
	cont.reset()
	assert.True(t, cont.seq.elems[0].cont.tool.Properties.(*toolPropertiesMock).conf.IsReset)
	assert.True(t, cont.seq.elems[1].cont.tool.Properties.(*toolPropertiesMock).conf.IsReset)

	// tool type
	cont = container{tool: &Tool{Properties: &toolPropertiesMock{conf: toolPropertiesMockSettings{}}}}
	cont.reset()
	assert.True(t, cont.tool.Properties.(*toolPropertiesMock).conf.IsReset)
}
