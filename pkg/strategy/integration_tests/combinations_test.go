package integration_tests

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestStrategiesWithMoreThanOneTool(t *testing.T) {
	type cycleRes struct {
		Res       bool
		ShouldErr bool
		Data      exchange.Data
		CallReset bool
	}

	tests := []struct {
		Name          string
		JSON          string
		JSONErr       bool
		Cycles        []cycleRes
		CandlesNeeded int
	}{
		{
			Name: "Strategy with simple change, bb, rollercoaster and buyprice tools (all separated by ANDs)",
			JSON: `{
		            "name":"strategy",
		            "seq":"test1 and test2 and test3 and test4",
		            "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
		            "tools":{
                        "test1":{
                            "type":"simpleChange",
                            "properties":{
                                "obj":"ask",
                                "cond": "equal",
                                "calcType": "FIXED",
                                "shiftVal": 80
                            }
                        },
                        "test2":{
                            "type":"bb",
                            "properties":{
                                "band": "UPPER",
                                "period": 3,
                                "stdev": 2,
                                "price": "close",
                                "maType": "sma",
                                "cond": "above",
                                "calcType": "units",
                                "shiftVal": -10,
                                "obj": "last"
                            }
                        },
                        "test3":{
                            "type":"rollerCoaster",
                            "properties":{
                                "pointType":"lowest",
                                "obj":"close",
                                "calcType": "units",
                                "shiftVal": 5
                            }
						},
						"test4":{
							"type":"buyprice",
							"properties":{
								"obj":"last",
								"cond":"above",
								"calcType": "units",
								"shiftVal": 2
							}
						}
		            }
		        }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(14, 0)},
							{Close: decimal.New(23, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(10, 0),
						},
						BuyPrice: decimal.RequireFromString("65.7"),
					},
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(20, 0)},
							{Close: decimal.New(30, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(70, 0),
							LastPrice: decimal.New(70, 0),
						},
						BuyPrice: decimal.RequireFromString("65.7"),
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(20, 0)},
							{Close: decimal.New(30, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(80, 0),
							LastPrice: decimal.New(70, 0),
						},
						BuyPrice: decimal.RequireFromString("65.7"),
					},
					CallReset: true,
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(20, 0)},
							{Close: decimal.New(30, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(80, 0),
							LastPrice: decimal.New(70, 0),
						},
						BuyPrice: decimal.RequireFromString("65.7"),
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(20, 0)},
							{Close: decimal.New(35, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(80, 0),
							LastPrice: decimal.New(70, 0),
						},
						BuyPrice: decimal.RequireFromString("65.7"),
					},
				},
			},
			CandlesNeeded: 3,
		},
		{
			Name: "Strategy with simple change, bb, rollercoaster and buyprice tools (all separated by ORs except for the buyprice)",
			JSON: `{
		            "name":"strategy",
		            "seq":"{ test1 or test2 or test3 } and test4",
		            "outcomes":[
                        {"type":"sandbox"}
                    ],
		            "tools":{
                        "test1":{
                            "type":"simpleChange",
                            "properties":{
                                "obj":"ask",
                                "cond": "EQUAL",
                                "calcType": "fixed",
                                "shiftVal": 80
                            }
                        },
                        "test2":{
                            "type":"bb",
                            "properties":{
                                "band": "UPPER",
                                "period": 3,
                                "stdev": 2,
                                "price": "Close",
                                "maType": "sma",
                                "cond": "Above",
                                "calcType": "units",
                                "shiftVal": -10,
                                "obj": "last"
                            }
                        },
                        "test3":{
                            "type":"rollercoaster",
                            "properties":{
                                "pointType":"lowest",
                                "obj":"close",
                                "calcType": "units",
                                "shiftVal": 5
                            }
						},
						"test4":{
							"type":"buyprice",
							"properties":{
								"obj":"last",
								"cond":"above",
								"calcType":"units",
								"shiftVal":2
							}
						}
		            }
		        }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(14, 0)},
							{Close: decimal.New(23, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(10, 0),
						},
						BuyPrice: decimal.RequireFromString("7.2"),
					},
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(20, 0)},
							{Close: decimal.New(27, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(10, 0),
						},
						BuyPrice: decimal.RequireFromString("7.2"),
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(20, 0)},
							{Close: decimal.New(30, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(10, 0),
						},
						BuyPrice: decimal.RequireFromString("7.2"),
					},
					CallReset: true,
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(20, 0)},
							{Close: decimal.New(30, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(10, 0),
						},
						BuyPrice: decimal.RequireFromString("7.2"),
					},
				},
			},
			CandlesNeeded: 3,
		},
		{
			Name: "Strategy with simple change, bb, rollercoaster and buyprice tools with one inner logic block",
			JSON: `{
		            "name":"strategy",
		            "seq":"test1 and test2 and { test3 or test4 }",
		            "outcomes":[
                        {"type":"sandbox"}
                    ],
		            "tools":{
                        "test1":{
                            "type":"simplechange",
                            "properties":{
                                "obj":"ask",
                                "cond": "equal",
                                "calcType": "Fixed",
                                "shiftVal": 30
                            }
						},
						"test2":{
							"type":"buyprice",
							"properties":{
								"obj":"last",
								"cond":"below",
								"calcType":"units",
								"shiftVal":5
							}
						},
                        "test3":{
                            "type":"rollercoaster",
                            "properties":{
                                "pointType":"lowest",
                                "obj":"close",
                                "calcType": "units",
                                "shiftVal": 5
                            }
                        },
                        "test4":{
                            "type":"bb",
                            "properties":{
                                "band": "upper",
                                "period": 3,
                                "stdev": 2,
                                "price": "close",
                                "maType": "sma",
                                "cond": "ABOVE",
                                "calcType": "units",
                                "shiftVal": 1,
                                "obj": "last"
                            }
                        }
		            }
		        }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(14, 0)},
							{Close: decimal.New(23, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(31, 0),
							LastPrice: decimal.New(25, 0),
						},
						BuyPrice: decimal.RequireFromString("45.111"),
					},
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(13, 0)},
							{Close: decimal.New(15, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(11, 0),
						},
						BuyPrice: decimal.RequireFromString("45.111"),
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(13, 0)},
							{Close: decimal.New(28, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(11, 0),
						},
						BuyPrice: decimal.RequireFromString("45.111"),
					},
					CallReset: true,
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(13, 0)},
							{Close: decimal.New(28, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(11, 0),
						},
						BuyPrice: decimal.RequireFromString("45.111"),
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(13, 0)},
							{Close: decimal.New(33, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(30, 0),
							LastPrice: decimal.New(40, 0),
						},
						BuyPrice: decimal.RequireFromString("45.111"),
					},
				},
			},
			CandlesNeeded: 3,
		},
		{
			Name: "Strategy with simple change, bb and rollercoaster tools with two inner logic blocks",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1 and { test2 or { test3 and test4 } }",
                    "outcomes":[
                        {"type":"sandbox"}
                    ],
                    "tools":{
                        "test1":{
                            "type":"simpleChange",
                            "properties":{
                                "obj":"ask",
                                "cond": "equal",
                                "calcType": "Fixed",
                                "shiftVal": 35
                            }
                        },
                        "test2":{
                            "type":"rollercoaster",
                            "properties":{
                                "pointType":"lowest",
                                "obj":"close",
                                "calcType": "units",
                                "shiftVal": 5
                            }
                        },
                        "test3":{
                            "type":"bb",
                            "properties":{
                                "band": "upper",
                                "period": 3,
                                "stdev": 2,
                                "price": "close",
                                "maType": "sma",
                                "cond": "above",
                                "calcType": "units",
                                "shiftVal": 1,
                                "obj": "last"
                            }
                        },
                        "test4":{
                            "type":"simplechange",
                            "properties":{
                                "obj":"last",
                                "cond": "equal",
                                "calcType": "units",
                                "shiftVal": 5
                            }
                        }
                    }
                }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(14, 0)},
							{Close: decimal.New(23, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(31, 0),
							LastPrice: decimal.New(25, 0),
						},
					},
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(13, 0)},
							{Close: decimal.New(15, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(31, 0),
							LastPrice: decimal.New(30, 0),
						},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(13, 0)},
							{Close: decimal.New(15, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(35, 0),
							LastPrice: decimal.New(30, 0),
						},
					},
					CallReset: true,
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(10, 0)},
							{Close: decimal.New(13, 0)},
							{Close: decimal.New(20, 0)},
						},
						Ticker: exchange.TickerData{
							AskPrice:  decimal.New(35, 0),
							LastPrice: decimal.New(30, 0),
						},
					},
				},
			},
			CandlesNeeded: 3,
		},
	}

	for _, v := range tests {
		v := v
		t.Run(v.Name, func(t *testing.T) {
			t.Parallel()
			strat := strategy.Strategy{}
			err := strat.UnmarshalJSON([]byte(v.JSON))
			if v.JSONErr {
				assert.NotNil(t, err)
				return
			} else {
				if !assert.Nil(t, err) {
					return
				}
			}

			assert.Equal(t, v.CandlesNeeded, strat.CandlesNeeded())
			for _, cyc := range v.Cycles {
				ok, err := strat.ReadyToAct(cyc.Data)
				if cyc.ShouldErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
				assert.Equal(t, cyc.Res, ok)
				if cyc.CallReset {
					strat.Reset(false)
				}
			}
		})
	}
}
