package strategy

import (
	"encoding/json"
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy/outcome"
	"eonbot/pkg/strategy/tools"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestStrategyClone(t *testing.T) {
	strat := &Strategy{
		name: "test",
		outcomes: []*outcome.Outcome{{
			Type: "dca",
			Raw: []byte(`{
                "type": "dca",
                "properties": {
                    "price": "bid",
                    "repeat": 3,
                    "amount": 0.1,
                    "calcType": "counterpercent"
                }
            }`),
			Conf: &outcome.DCA{
				Repeat: 3,
				Buy: outcome.Buy{
					Price:       exchange.BidPrice,
					Amount:      decimal.RequireFromString("0.1"),
					Calc:        "counterpercent",
					BasePercent: true,
				},
				StateIndex: 0,
			},
		}},
		minCandles: 12,
		origSeq:    "test1",
		seq: &sequence{
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
	}

	newStrat, err := strat.Clone()
	assert.Nil(t, err)
	assert.Equal(t, strat, newStrat)

	newStrat1, err := strat.Clone()
	assert.Nil(t, err)
	newStrat1.outcomes[0].Conf.(*outcome.DCA).StateIndex = 1
	assert.NotEqual(t, newStrat, newStrat1)
}

func TestStrategyGetters(t *testing.T) {
	strat := Strategy{
		name:       "test",
		outcomes:   []*outcome.Outcome{{}},
		minCandles: 12,
	}
	assert.Equal(t, "test", strat.Name())
	assert.Equal(t, []*outcome.Outcome{{}}, strat.Outcomes())
	assert.Equal(t, 12, strat.CandlesNeeded())
}

func TestStrategyReset(t *testing.T) {
	strat := Strategy{
		outcomes: []*outcome.Outcome{
			{
				Type: outcome.DCAOutcome,
				Conf: &outcome.DCA{
					StateIndex: 2,
				},
			},
		},
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

	strat.Reset(true)
	assert.Equal(t, 0, strat.outcomes[0].Conf.(*outcome.DCA).StateIndex)
	assert.True(t, strat.seq.elems[0].cont.tool.Properties.(*toolPropertiesMock).conf.IsReset)
}

func TestStrategyReadyToAct(t *testing.T) {
	strat := Strategy{
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
				},
			},
		},
	}

	ok, err := strat.ReadyToAct(exchange.Data{})
	assert.True(t, ok)
	assert.Nil(t, err)

	// panic error
	strat = Strategy{
		seq: &sequence{
			elems: []*seqElem{
				{
					cont: &container{
						tool: &Tool{
							Properties: &toolPropertiesMock{
								conf: toolPropertiesMockSettings{
									Err:   "test",
									Panic: true,
								},
							},
						},
					},
				},
			},
		},
	}

	ok, err = strat.ReadyToAct(exchange.Data{})
	assert.False(t, ok)
	assert.NotNil(t, err)

	// normal error
	strat = Strategy{
		seq: &sequence{
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
		},
	}

	ok, err = strat.ReadyToAct(exchange.Data{})
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestStrategySnapshot(t *testing.T) {
	strat := Strategy{
		origSeq: "test1 and test2",
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

	res := Snapshot{
		Seq: "test1 and test2",
		Tools: map[string]tools.FullSnapshot{
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
		},
	}

	assert.Equal(t, res, strat.Snapshot())
}

func TestStrategyUnmarshalJSON(t *testing.T) {
	tests := []struct {
		Name      string
		JSON      string
		Res       Strategy
		ShouldErr bool
	}{
		{
			Name:      "Invalid JSON results in error",
			JSON:      "{",
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name:      "Empty strategy name results in error",
			JSON:      `{"name":""}`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name:      "Invalid strategy name results in error",
			JSON:      `{"name":"^!!$#"}`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name:      "Empty tools list results in error",
			JSON:      `{"name":"test"}`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name: "Invalid tool ID results in error",
			JSON: `{
		            "name":"test1",
		            "seq":"test2 and test3",
		            "outcomes":[],
		            "tools":{
		                "test)2":{
		                    "type": "test"
		                },
		                "test3":{
		                    "type": "test"
		                }
		            }
		        }`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name: "Tool ID equal to strategy name results in error",
			JSON: `{
		            "name":"test2",
		            "seq":"test2 and test3",
		            "outcomes":[],
		            "tools":{
		                "test2":{
		                    "type": "test"
		                },
		                "test3":{
		                    "type": "test"
		                }
		            }
		        }`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name: "Invalid tool properties results in error",
			JSON: `{
		            "name":"test3",
		            "seq":"test4",
		            "outcomes":[],
		            "tools":{
		                "test4":{
		                    "type": "test",
		                    "properties":{
		                        "count":true
		                    }
		                }
		            }
		        }`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name: "Tool ID usage in a sequence more than once results in error",
			JSON: `{
		            "name":"test4",
		            "seq":"test5 and test5",
		            "outcomes":[],
		            "tools":{
		                "test5":{
		                    "type": "test",
		                    "properties":{
		                        "count":10
		                    }
		                }
		            }
		        }`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name: "Sequence validation returns error",
			JSON: `{
		            "name":"test5",
		            "seq":"test6",
		            "outcomes":[],
		            "tools":{
		                "test6":{
		                    "type": "test",
		                    "properties":{
		                        "err":"test"
		                    }
		                }
		            }
		        }`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name: "Sequence validation panics",
			JSON: `{
		            "name":"test5",
		            "seq":"test6",
		            "outcomes":[],
		            "tools":{
		                "test6":{
		                    "type": "test",
		                    "properties":{
		                        "err":"test",
                                "panic":true
		                    }
		                }
		            }
		        }`,
			Res:       Strategy{},
			ShouldErr: true,
		},
		{
			Name: "Successful Strategy JSON unmarshal",
			JSON: `{
		            "name":"test6#+_-/",
		            "seq":"test7 and { test8 && { test9 | test10 } } OR test11",
		            "outcomes":[
                        {
                            "type": "buy",
                            "properties": {
                                "price": "ask",
                                "amount": 50.5,
                                "calcType": "counterunits"
                            }
                        }
                    ],
		            "tools":{
		                "test7":{
		                    "type": "test",
                            "properties":{"condsMet": true,"count":1}
		                },
                        "test8":{
		                    "type": "test",
                            "properties":{"condsMet": true,"count":9}
		                },
                        "test9":{
		                    "type": "test",
                            "properties":{"condsMet": true,"count":3}
		                },
                        "test10":{
		                    "type": "test",
                            "properties":{"condsMet": true,"count":5}
		                },
                        "test11":{
		                    "type": "test",
                            "properties":{"condsMet": true,"count":10}
		                }
		            }
		        }`,
			Res: Strategy{
				origSeq: "test7 and { test8 && { test9 | test10 } } OR test11",
				seq: &sequence{
					elems: []*seqElem{
						{
							cont: &container{
								tool: &Tool{
									ID:            "test7",
									Type:          testTool,
									RawProperties: json.RawMessage(`{"condsMet": true,"count":1}`),
									Properties: &toolPropertiesMock{
										conf: toolPropertiesMockSettings{
											CondsMet: true,
											Count:    1,
										},
									},
									assigned: true,
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
													ID:            "test8",
													Type:          testTool,
													RawProperties: json.RawMessage(`{"condsMet": true,"count":9}`),
													Properties: &toolPropertiesMock{
														conf: toolPropertiesMockSettings{
															CondsMet: true,
															Count:    9,
														},
													},
													assigned: true,
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
																	ID:            "test9",
																	Type:          testTool,
																	RawProperties: json.RawMessage(`{"condsMet": true,"count":3}`),
																	Properties: &toolPropertiesMock{
																		conf: toolPropertiesMockSettings{
																			CondsMet: true,
																			Count:    3,
																		},
																	},
																	assigned: true,
																},
															},
															joinNextWith: containerJoint_OR,
														},
														{
															cont: &container{
																tool: &Tool{
																	ID:            "test10",
																	Type:          testTool,
																	RawProperties: json.RawMessage(`{"condsMet": true,"count":5}`),
																	Properties: &toolPropertiesMock{
																		conf: toolPropertiesMockSettings{
																			CondsMet: true,
																			Count:    5,
																		},
																	},
																	assigned: true,
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
							joinNextWith: containerJoint_OR,
						},
						{
							cont: &container{
								tool: &Tool{
									ID:            "test11",
									Type:          testTool,
									RawProperties: json.RawMessage(`{"condsMet": true,"count":10}`),
									Properties: &toolPropertiesMock{
										conf: toolPropertiesMockSettings{
											CondsMet: true,
											Count:    10,
										},
									},
									assigned: true,
								},
							},
						},
					},
				},
				outcomes: []*outcome.Outcome{
					{
						Type: outcome.BuyOutcome,
						Raw: []byte(`{
                            "type": "buy",
                            "properties": {
                                "price": "ask",
                                "amount": 50.5,
                                "calcType": "counterunits"
                            }
                        }`),
						Conf: &outcome.Buy{
							Price:  exchange.AskPrice,
							Amount: decimal.RequireFromString("50.5"),
							Calc:   outcome.CalcCounterUnits,
						},
					},
				},
				minCandles: 10,
				stratType:  BuyModeStrat,
			},
			ShouldErr: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			strat := Strategy{}
			err := strat.UnmarshalJSON([]byte(v.JSON))
			if v.ShouldErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, v.Res, strat)
			}
		})
	}
}

func TestDetermineType(t *testing.T) {
	tests := []struct {
		Name        string
		Outcomes    []*outcome.Outcome
		StratType   string
		ShouldError bool
	}{
		{
			Name:        "Unsuccessful type determination when outcomes list is empty",
			Outcomes:    []*outcome.Outcome{},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when outcome has invalid type",
			Outcomes: []*outcome.Outcome{
				{
					Type: "test",
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when two buy outcomes are present",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.BuyOutcome,
					Conf: &outcome.Buy{
						Price:  exchange.AskPrice,
						Amount: decimal.New(100, 0),
						Calc:   outcome.CalcCounterPercent,
					},
				},
				{
					Type: outcome.BuyOutcome,
					Conf: &outcome.Buy{
						Price:  exchange.AskPrice,
						Amount: decimal.New(100, 0),
						Calc:   outcome.CalcCounterPercent,
					},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when buy outcome is used after sell",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.SellOutcome,
					Conf: &outcome.Sell{
						Price: exchange.AskPrice,
					},
				},
				{
					Type: outcome.BuyOutcome,
					Conf: &outcome.Buy{
						Price:  exchange.AskPrice,
						Amount: decimal.New(100, 0),
						Calc:   outcome.CalcCounterPercent,
					},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when buy outcome is used after dca",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.DCAOutcome,
					Conf: &outcome.DCA{},
				},
				{
					Type: outcome.BuyOutcome,
					Conf: &outcome.Buy{
						Price:  exchange.AskPrice,
						Amount: decimal.New(100, 0),
						Calc:   outcome.CalcCounterPercent,
					},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when two sell outcomes are present",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.SellOutcome,
					Conf: &outcome.Sell{
						Price: exchange.AskPrice,
					},
				},
				{
					Type: outcome.SellOutcome,
					Conf: &outcome.Sell{
						Price: exchange.AskPrice,
					},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when sell outcome is used after buy",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.BuyOutcome,
					Conf: &outcome.Buy{
						Price:  exchange.AskPrice,
						Amount: decimal.New(100, 0),
						Calc:   outcome.CalcCounterPercent,
					},
				},
				{
					Type: outcome.SellOutcome,
					Conf: &outcome.Sell{
						Price: exchange.AskPrice,
					},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when sell outcome is used after dca",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.DCAOutcome,
					Conf: &outcome.DCA{},
				},
				{
					Type: outcome.SellOutcome,
					Conf: &outcome.Sell{
						Price: exchange.AskPrice,
					},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when two dca outcomes are present",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.DCAOutcome,
					Conf: &outcome.DCA{},
				},
				{
					Type: outcome.DCAOutcome,
					Conf: &outcome.DCA{},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when dca outcome is used after buy",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.BuyOutcome,
					Conf: &outcome.Buy{
						Price:  exchange.AskPrice,
						Amount: decimal.New(100, 0),
						Calc:   outcome.CalcCounterPercent,
					},
				},
				{
					Type: outcome.DCAOutcome,
					Conf: &outcome.DCA{},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when dca outcome is used after sell",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.SellOutcome,
					Conf: &outcome.Sell{
						Price: exchange.AskPrice,
					},
				},
				{
					Type: outcome.DCAOutcome,
					Conf: &outcome.DCA{},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Successful buy strategy type determination",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.BuyOutcome,
					Conf: &outcome.Buy{
						Price:  exchange.AskPrice,
						Amount: decimal.New(100, 0),
						Calc:   outcome.CalcCounterPercent,
					},
				},
				{
					Type: outcome.TelegramOutcome,
					Conf: &outcome.Telegram{
						Selection: outcome.RotatingSelection,
						Messages:  []string{"test"},
					},
				},
			},
			StratType:   BuyModeStrat,
			ShouldError: false,
		},
		{
			Name: "Successful sell strategy type determination",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.SellOutcome,
					Conf: &outcome.Sell{
						Price: exchange.AskPrice,
					},
				},
				{
					Type: outcome.TelegramOutcome,
					Conf: &outcome.Telegram{
						Selection: outcome.RotatingSelection,
						Messages:  []string{"test"},
					},
				},
			},
			StratType:   SellModeStrat,
			ShouldError: false,
		},
		{
			Name: "Successful sell strategy type determination",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.DCAOutcome,
					Conf: &outcome.DCA{},
				},
				{
					Type: outcome.TelegramOutcome,
					Conf: &outcome.Telegram{
						Selection: outcome.RotatingSelection,
						Messages:  []string{"test"},
					},
				},
			},
			StratType:   SellModeStrat,
			ShouldError: false,
		},
		{
			Name: "Unsuccessful type determination when two telegram outcomes are present",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.TelegramOutcome,
					Conf: &outcome.Telegram{
						Selection: outcome.RotatingSelection,
						Messages:  []string{"test"},
					},
				},
				{
					Type: outcome.TelegramOutcome,
					Conf: &outcome.Telegram{
						Selection: outcome.RotatingSelection,
						Messages:  []string{"test"},
					},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Successful any mode with telegram outcom strategy type determination",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.TelegramOutcome,
					Conf: &outcome.Telegram{
						Selection: outcome.RotatingSelection,
						Messages:  []string{"test"},
					},
				},
			},
			StratType:   AnyModeStrat,
			ShouldError: false,
		},
		{
			Name: "Unsuccessful type determination when two sandbox outcomes are present",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.SandboxOutcome,
				},
				{
					Type: outcome.SandboxOutcome,
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Unsuccessful type determination when sell outcome goes after sandbox outcome",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.SandboxOutcome,
				},
				{
					Type: outcome.SellOutcome,
					Conf: &outcome.Sell{
						Price: exchange.AskPrice,
					},
				},
			},
			StratType:   "",
			ShouldError: true,
		},
		{
			Name: "Successful any mode with sandbox strategy type determination",
			Outcomes: []*outcome.Outcome{
				{
					Type: outcome.SandboxOutcome,
				},
			},
			StratType:   AnyModeStrat,
			ShouldError: false,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			strat := Strategy{outcomes: v.Outcomes}
			err := strat.determineType()
			if v.ShouldError {
				assert.NotNil(t, err)
				assert.Empty(t, strat.Type())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, v.StratType, strat.Type())
			}
		})
	}
}
