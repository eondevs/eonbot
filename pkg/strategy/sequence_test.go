package strategy

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/tools"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsJointValid(t *testing.T) {
	tests := []struct {
		Name  string
		Joint int
	}{
		{
			Name:  "Valid AND joint",
			Joint: containerJoint_AND,
		},
		{
			Name:  "Valid OR joint",
			Joint: containerJoint_OR,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, isJointValid(v.Joint))
		})
	}

	assert.False(t, isJointValid(-1))
}

func TestNewRootSequence(t *testing.T) {
	seq, err := newRootSequence("test1 and { test2 or test3 }", map[string]*Tool{
		"test1": {ID: "test1"},
		"test2": {ID: "test2"},
		"test3": {ID: "test3"},
	})
	res := &sequence{
		elems: []*seqElem{
			{
				cont: &container{
					tool: &Tool{ID: "test1", assigned: true},
				},
				joinNextWith: containerJoint_AND,
			},
			{
				cont: &container{
					seq: &sequence{
						elems: []*seqElem{
							{
								cont: &container{
									tool: &Tool{ID: "test2", assigned: true},
								},
								joinNextWith: containerJoint_OR,
							},
							{
								cont: &container{
									tool: &Tool{ID: "test3", assigned: true},
								},
							},
						},
					},
				},
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, res, seq)
}

func TestSequenceClone(t *testing.T) {
	seq := &sequence{
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
	}

	newSeq, err := seq.clone()
	assert.Nil(t, err)
	assert.Equal(t, seq, newSeq)

	newSeq1, err := seq.clone()
	assert.Nil(t, err)

	newSeq1.elems[0].cont.tool.Type = "test123"
	assert.NotEqual(t, newSeq, newSeq1)
}

func TestSequenceValidate(t *testing.T) {
	seq := sequence{
		elems: []*seqElem{
			{
				cont: &container{
					tool: &Tool{
						Properties: &toolPropertiesMock{
							conf: toolPropertiesMockSettings{
								Err: "test",
							},
						},
					},
				},
			},
		},
	}
	assert.NotNil(t, seq.validate())

	// success
	seq = sequence{
		elems: []*seqElem{
			{
				cont: &container{
					tool: &Tool{
						Properties: &toolPropertiesMock{},
					},
				},
			},
		},
	}
	assert.Nil(t, seq.validate())
}

func TestSequenceCandlesCount(t *testing.T) {
	seq := sequence{
		elems: []*seqElem{
			{
				cont: &container{
					seq: &sequence{
						elems: []*seqElem{
							{
								cont: &container{
									tool: &Tool{
										Properties: &toolPropertiesMock{
											conf: toolPropertiesMockSettings{
												Count: 20,
											},
										},
									},
								},
							},
							{
								cont: &container{
									tool: &Tool{
										Properties: &toolPropertiesMock{
											conf: toolPropertiesMockSettings{
												Count: 23,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				cont: &container{
					seq: &sequence{
						elems: []*seqElem{
							{
								cont: &container{
									tool: &Tool{
										Properties: &toolPropertiesMock{
											conf: toolPropertiesMockSettings{
												Count: 19,
											},
										},
									},
								},
							},
							{
								cont: &container{
									tool: &Tool{
										Properties: &toolPropertiesMock{
											conf: toolPropertiesMockSettings{
												Count: 38,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, 38, seq.candlesCount())
}

func TestSequenceSnapshot(t *testing.T) {
	seq := sequence{
		elems: []*seqElem{
			{
				cont: &container{
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
				},
			},
			{
				cont: &container{
					seq: &sequence{
						elems: []*seqElem{
							{
								cont: &container{
									tool: &Tool{
										Type: testTool,
										ID:   "test3",
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
										ID:   "test4",
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
		"test3": tools.FullSnapshot{
			Type: testTool,
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data:     123,
			},
		},
		"test4": tools.FullSnapshot{
			Type: testTool,
			Snapshot: tools.Snapshot{
				CondsMet: true,
				Data:     123,
			},
		},
	}

	assert.Equal(t, res, seq.snapshot())
}

func TestSeqElemFilled(t *testing.T) {
	elem := seqElem{
		cont:         newContTool(&Tool{}),
		joinNextWith: containerJoint_AND,
	}

	assert.True(t, elem.filled())
}

func TestSeqElemOnlyCont(t *testing.T) {
	elem := seqElem{
		cont: newContTool(&Tool{}),
	}

	assert.True(t, elem.onlyCont())
}

func TestBracketAdd(t *testing.T) {
	brc := bracketInfo{}
	brc.add("test1")
	assert.Equal(t, "test1", brc.content.String())

	brc.add("test2")
	assert.Equal(t, "test1 test2", brc.content.String())
}

func TestBracketReset(t *testing.T) {
	brc := bracketInfo{}
	brc.add("test")
	brc.skip = 5
	brc.inside = true
	brc.reset()
	assert.Equal(t, bracketInfo{}, brc)
}

func TestSeqFromString(t *testing.T) {
	tests := []struct {
		Name      string
		Seq       string
		Tools     map[string]*Tool
		Res       *sequence
		ShouldErr bool
	}{
		{
			Name:      "Empty sequence string results in error",
			Seq:       "",
			Tools:     make(map[string]*Tool),
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Empty tools list results in error",
			Seq:       "test1",
			Tools:     nil,
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Unexpected closing bracket results in error",
			Seq:       "}",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Opening bracket right after previous ID or container and not joint results in error",
			Seq:       "test1 { test2 and test3 }",
			Tools:     map[string]*Tool{"test1": {}, "test2": {}, "test3": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence starting with AND results in error",
			Seq:       "AND test1",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence containing two AND keywords one after another results in error",
			Seq:       "test1 AND and",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence starting with OR results in error",
			Seq:       "OR test1",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence containing two OR keywords one after another results in error",
			Seq:       "test1 OR or",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence tool ID containing invalid symbol results in error",
			Seq:       "test1!!!",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence tool ID pointing to a tool that does not exist results in error",
			Seq:       "test",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Usage of already assigned tool in sequence results in error",
			Seq:       "test1 AND test1",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Two tools IDs in a row results in error",
			Seq:       "test1 test2",
			Tools:     map[string]*Tool{"test1": {}, "test2": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence ending with a reserved keyword results in error",
			Seq:       "test1 AND",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence inner logic block ending with a reserved keyword results in error",
			Seq:       "test1 AND { test2 AND }",
			Tools:     map[string]*Tool{"test1": {}, "test2": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Sequence ending without closing bracket results in error",
			Seq:       "{ test1",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:      "Empty sequence inner logic block results in error",
			Seq:       "test1 and { }",
			Tools:     map[string]*Tool{"test1": {}},
			Res:       nil,
			ShouldErr: true,
		},
		{
			Name:  "Successful serialization of a sequence that is inside a pair of brackets",
			Seq:   "{ test1 and test2 }",
			Tools: map[string]*Tool{"test1": {ID: "test1"}, "test2": {ID: "test2"}},
			Res: &sequence{
				elems: []*seqElem{
					{
						cont: &container{
							seq: &sequence{
								elems: []*seqElem{
									{
										cont: &container{
											tool: &Tool{ID: "test1", assigned: true},
										},
										joinNextWith: containerJoint_AND,
									},
									{
										cont: &container{
											tool: &Tool{ID: "test2", assigned: true},
										},
									},
								},
							},
						},
					},
				},
			},
			ShouldErr: false,
		},
		{
			Name:  "Successful serialization of a sequence with one inner logic block",
			Seq:   "test1 AND { test2 or test3 }",
			Tools: map[string]*Tool{"test1": {ID: "test1"}, "test2": {ID: "test2"}, "test3": {ID: "test3"}},
			Res: &sequence{
				elems: []*seqElem{
					{
						cont: &container{
							tool: &Tool{ID: "test1", assigned: true},
						},
						joinNextWith: containerJoint_AND,
					},
					{
						cont: &container{
							seq: &sequence{
								elems: []*seqElem{
									{
										cont: &container{
											tool: &Tool{ID: "test2", assigned: true},
										},
										joinNextWith: containerJoint_OR,
									},
									{
										cont: &container{
											tool: &Tool{ID: "test3", assigned: true},
										},
									},
								},
							},
						},
					},
				},
			},
			ShouldErr: false,
		},
		{
			Name: "Successful serialization of a sequence with an AND keyword inside of 5 containers",
			Seq:  "{ { { { { test1 and test2 } } } } }",
			Tools: map[string]*Tool{
				"test1": {ID: "test1"},
				"test2": {ID: "test2"},
			},
			Res: &sequence{
				elems: []*seqElem{
					{
						cont: &container{
							seq: &sequence{
								elems: []*seqElem{
									{
										cont: &container{
											seq: &sequence{
												elems: []*seqElem{
													{
														cont: &container{
															seq: &sequence{
																elems: []*seqElem{
																	{
																		cont: &container{
																			seq: &sequence{
																				elems: []*seqElem{
																					{
																						cont: &container{
																							seq: &sequence{
																								elems: []*seqElem{
																									{
																										cont: &container{
																											tool: &Tool{ID: "test1", assigned: true},
																										},
																										joinNextWith: containerJoint_AND,
																									},
																									{
																										cont: &container{
																											tool: &Tool{ID: "test2", assigned: true},
																										},
																									},
																								},
																							},
																						},
																					},
																				},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			ShouldErr: false,
		},
		{
			Name: "Successful serialization of a sequence with more than one multi level inner logic blocks",
			Seq:  "test1 & { { { test2 && test3 } AND test4 } | { { test5 } && test6 } } & { test7 or test8 || test9 }",
			Tools: map[string]*Tool{
				"test1": {ID: "test1"},
				"test2": {ID: "test2"},
				"test3": {ID: "test3"},
				"test4": {ID: "test4"},
				"test5": {ID: "test5"},
				"test6": {ID: "test6"},
				"test7": {ID: "test7"},
				"test8": {ID: "test8"},
				"test9": {ID: "test9"},
			},
			Res: &sequence{
				elems: []*seqElem{
					{
						cont: &container{
							tool: &Tool{ID: "test1", assigned: true},
						},
						joinNextWith: containerJoint_AND,
					},
					{
						cont: &container{
							seq: &sequence{
								elems: []*seqElem{
									{
										cont: &container{
											seq: &sequence{
												elems: []*seqElem{
													{
														cont: &container{
															seq: &sequence{
																elems: []*seqElem{
																	{
																		cont: &container{
																			tool: &Tool{ID: "test2", assigned: true},
																		},
																		joinNextWith: containerJoint_AND,
																	},
																	{
																		cont: &container{
																			tool: &Tool{ID: "test3", assigned: true},
																		},
																	},
																},
															},
														},
														joinNextWith: containerJoint_AND,
													},
													{
														cont: &container{
															tool: &Tool{ID: "test4", assigned: true},
														},
													},
												},
											},
										},
										joinNextWith: containerJoint_OR,
									},
									{
										cont: &container{
											seq: &sequence{
												elems: []*seqElem{
													{
														cont: &container{
															seq: &sequence{
																elems: []*seqElem{
																	{
																		cont: &container{
																			tool: &Tool{ID: "test5", assigned: true},
																		},
																	},
																},
															},
														},
														joinNextWith: containerJoint_AND,
													},
													{
														cont: &container{
															tool: &Tool{ID: "test6", assigned: true},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						joinNextWith: containerJoint_AND,
					},
					{
						cont: &container{
							seq: &sequence{
								elems: []*seqElem{
									{
										cont: &container{
											tool: &Tool{ID: "test7", assigned: true},
										},
										joinNextWith: containerJoint_OR,
									},
									{
										cont: &container{
											tool: &Tool{ID: "test8", assigned: true},
										},
										joinNextWith: containerJoint_OR,
									},
									{
										cont: &container{
											tool: &Tool{ID: "test9", assigned: true},
										},
									},
								},
							},
						},
					},
				},
			},
			ShouldErr: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := seqFromString(v.Seq, v.Tools, true)
			if v.ShouldErr {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, v.Res, res)
				assert.Nil(t, err)
			}
		})
	}
}

func TestSequenceConditionsMet(t *testing.T) {
	tests := []struct {
		Name      string
		Seq       *sequence
		Res       bool
		ShouldErr bool
	}{
		{
			Name:      "Empty sequence slice results in error",
			Seq:       &sequence{},
			Res:       false,
			ShouldErr: true,
		},
		{
			Name: "Last sequence slice element without STOP joint results in error",
			Seq: &sequence{elems: []*seqElem{
				{
					cont:         &container{tool: &Tool{}},
					joinNextWith: containerJoint_AND,
				},
				{
					cont:         &container{tool: &Tool{}},
					joinNextWith: containerJoint_AND,
				},
			},
			},
			Res:       false,
			ShouldErr: true,
		},
		{
			Name: "Conditions not met when one of the tools returns error",
			Seq: &sequence{
				elems: []*seqElem{
					{
						cont: &container{
							tool: &Tool{
								Properties: &toolPropertiesMock{
									conf: toolPropertiesMockSettings{
										CondsMet: false,
									},
								},
							},
						},
						joinNextWith: containerJoint_OR,
					},
					{
						cont: &container{
							tool: &Tool{
								Properties: &toolPropertiesMock{
									conf: toolPropertiesMockSettings{
										Err: "test",
									},
								},
							},
						},
					},
				},
			},
			Res:       false,
			ShouldErr: true,
		},
		{
			Name: "Conditions not met when one of the tools returns false",
			Seq: &sequence{
				elems: []*seqElem{
					{
						cont: &container{
							tool: &Tool{
								Properties: &toolPropertiesMock{
									conf: toolPropertiesMockSettings{
										CondsMet: true,
									},
								},
							},
						},
						joinNextWith: containerJoint_AND,
					},
					{
						cont: &container{
							tool: &Tool{
								Properties: &toolPropertiesMock{
									conf: toolPropertiesMockSettings{
										CondsMet: false,
									},
								},
							},
						},
					},
				},
			},
			Res:       false,
			ShouldErr: false,
		},
		{
			Name: "Successful sequence when one of the OR cases return true",
			Seq: &sequence{
				elems: []*seqElem{
					{
						cont: &container{
							tool: &Tool{
								Properties: &toolPropertiesMock{
									conf: toolPropertiesMockSettings{
										CondsMet: false,
									},
								},
							},
						},
						joinNextWith: containerJoint_OR,
					},
					{
						cont: &container{
							tool: &Tool{
								Properties: &toolPropertiesMock{
									conf: toolPropertiesMockSettings{
										CondsMet: true,
									},
								},
							},
						},
					},
				},
			},
			Res:       true,
			ShouldErr: false,
		},
		{
			Name: "Successful sequence with more than one inner multi level logic block check",
			Seq: &sequence{ // test1 and { test2 && { test3 or test4} | test5 } and test6
				elems: []*seqElem{
					{
						cont: &container{
							tool: &Tool{
								Properties: &toolPropertiesMock{
									conf: toolPropertiesMockSettings{
										CondsMet: true,
									},
								},
							},
						},
						joinNextWith: containerJoint_AND,
					},
					{
						cont: &container{
							seq: &sequence{
								elems: []*seqElem{
									{
										cont: &container{
											tool: &Tool{
												Properties: &toolPropertiesMock{
													conf: toolPropertiesMockSettings{
														CondsMet: true,
													},
												},
											},
										},
										joinNextWith: containerJoint_AND,
									},
									{
										cont: &container{
											seq: &sequence{
												elems: []*seqElem{
													{
														cont: &container{
															tool: &Tool{
																Properties: &toolPropertiesMock{
																	conf: toolPropertiesMockSettings{
																		CondsMet: false,
																	},
																},
															},
														},
														joinNextWith: containerJoint_OR,
													},
													{
														cont: &container{
															tool: &Tool{
																Properties: &toolPropertiesMock{
																	conf: toolPropertiesMockSettings{
																		CondsMet: false,
																	},
																},
															},
														},
													},
												},
											},
										},
										joinNextWith: containerJoint_OR,
									},
									{
										cont: &container{
											tool: &Tool{
												Properties: &toolPropertiesMock{
													conf: toolPropertiesMockSettings{
														CondsMet: true,
													},
												},
											},
										},
									},
								},
							},
						},
						joinNextWith: containerJoint_AND,
					},
					{
						cont: &container{
							tool: &Tool{
								Properties: &toolPropertiesMock{
									conf: toolPropertiesMockSettings{
										CondsMet: true,
									},
								},
							},
						},
					},
				},
			},
			Res:       true,
			ShouldErr: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			res, err := v.Seq.conditionsMet(exchange.Data{})
			if v.ShouldErr {
				assert.NotNil(t, err)
			} else {
				assert.Equal(t, v.Res, res)
				assert.Nil(t, err)
			}
		})
	}
}

func TestSequenceReset(t *testing.T) {
	seq := sequence{
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
	}
	seq.reset()
	assert.True(t, seq.elems[0].cont.tool.Properties.(*toolPropertiesMock).conf.IsReset)
	assert.True(t, seq.elems[1].cont.tool.Properties.(*toolPropertiesMock).conf.IsReset)
}
