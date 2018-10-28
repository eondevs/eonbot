package integration_tests

import (
	"eonbot/pkg/exchange"
	"eonbot/pkg/strategy"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestStrategiesWithOneTool(t *testing.T) {
	type cycleRes struct {
		Res       bool
		ShouldErr bool
		Data      exchange.Data
	}

	tests := []struct {
		Name          string
		JSON          string
		JSONErr       bool
		Cycles        []cycleRes
		CandlesNeeded int
	}{
		{
			Name: "Strategy with simple change tool that uses LastPrice as change object with fixed price",
			JSON: `{
		            "name":"strategy#1+test1",
		            "seq":"test1",
		            "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
		            "tools":{
		                "test1":{
		                    "type":"simpleChange",
		                    "properties":{
		                        "obj":"ASK",
		                        "cond": "equal",
		                        "calcType": "FIxed",
		                        "shiftVal": 100
		                    }
		                }
		            }
		        }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Ticker: exchange.TickerData{AskPrice: decimal.New(100, 0)},
					},
				},
				{ // to confirm that tool was not reset
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Ticker: exchange.TickerData{AskPrice: decimal.New(100, 0)},
					},
				},
			},
			CandlesNeeded: 0,
		},
		{
			Name: "Strategy with simple change tool that uses OpenPrice as change object and waits for exact match",
			JSON: `{
		            "name":"strategy#2+-/hey",
		            "seq":"test1#",
		            "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
		            "tools":{
		                "test1#":{
		                    "type":"SimpleChange",
		                    "properties":{
		                        "obj":"OPEN",
		                        "cond": "equal",
		                        "calcType": "units",
		                        "shiftVal": -100
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
							{Open: decimal.New(200, 0)},
						},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Open: decimal.New(100, 0)},
						},
					},
				},
				{ // to confirm that tool was not reset
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Open: decimal.New(100, 0)},
						},
					},
				},
			},
			CandlesNeeded: 1,
		},
		{
			Name: "Strategy with simple change tool that uses SMA as change object and waits for increase",
			JSON: `{
		            "name":"strategy.",
		            "seq":"test1+",
		            "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
		            "tools":{
		                "test1+":{
		                    "type":"simpleChange",
		                    "properties":{
		                        "obj":"SMA",
		                        "objConf":{
		                            "period": 5,
		                            "price": "Close"
		                        },
		                        "cond": "aboveOrEQUAL",
		                        "calcType": "percent",
		                        "shiftVal": 0.5
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
							{Close: decimal.RequireFromString("1.5")},
							{Close: decimal.RequireFromString("2")},
							{Close: decimal.RequireFromString("3")},
							{Close: decimal.RequireFromString("2.5")},
							{Close: decimal.RequireFromString("3.5")},
						},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.RequireFromString("2")},
							{Close: decimal.RequireFromString("3")},
							{Close: decimal.RequireFromString("2.5")},
							{Close: decimal.RequireFromString("3.5")},
							{Close: decimal.RequireFromString("4")},
						},
					},
				},
				{ // to confirm that tool was not reset
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.RequireFromString("2")},
							{Close: decimal.RequireFromString("3")},
							{Close: decimal.RequireFromString("2.5")},
							{Close: decimal.RequireFromString("3.5")},
							{Close: decimal.RequireFromString("4")},
						},
					},
				},
			},
			CandlesNeeded: 5,
		},
		{
			Name: "Strategy with simple change tool that uses EMA as change object and waits for decrease",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"SIMPLEchange",
                            "properties":{
                                "obj":"EMa",
                                "objConf":{
                                    "period": 3,
                                    "price": "high"
                                },
                                "cond": "BELOW",
                                "calcType": "percent",
                                "shiftVal": -1
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
							{High: decimal.New(10, 0)},
							{High: decimal.New(11, 0)},
							{High: decimal.New(12, 0)},
							{High: decimal.New(13, 0)},
							{High: decimal.New(14, 0)},
							{High: decimal.New(15, 0)},
						},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{High: decimal.New(10, 0)},
							{High: decimal.New(11, 0)},
							{High: decimal.New(11, 0)},
							{High: decimal.New(10, 0)},
							{High: decimal.New(9, 0)},
							{High: decimal.New(11, 0)},
						},
					},
				},
				{ // to confirm that tool was not reset
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{High: decimal.New(10, 0)},
							{High: decimal.New(11, 0)},
							{High: decimal.New(11, 0)},
							{High: decimal.New(10, 0)},
							{High: decimal.New(9, 0)},
							{High: decimal.New(11, 0)},
						},
					},
				},
			},
			CandlesNeeded: 6,
		},
		{
			Name: "Strategy with simple change tool that uses WMA as change object and waits for exact match",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"simplechange",
                            "properties":{
                                "obj":"wma",
                                "objConf":{
                                    "period": 3,
                                    "price": "low"
                                },
                                "cond": "equAL",
                                "calcType": "percent",
                                "shiftVal": -50
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
							{Low: decimal.New(10, 0)},
							{Low: decimal.New(40, 0)},
							{Low: decimal.New(30, 0)},
						},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Low: decimal.New(5, 0)},
							{Low: decimal.New(5, 0)},
							{Low: decimal.New(25, 0)},
						},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Low: decimal.New(5, 0)},
							{Low: decimal.New(5, 0)},
							{Low: decimal.New(25, 0)},
						},
					},
				},
				{ // to confirm that tool was not reset
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Low: decimal.New(5, 0)},
							{Low: decimal.New(5, 0)},
							{Low: decimal.New(25, 0)},
						},
					},
				},
			},
			CandlesNeeded: 3,
		},
		{
			Name: "Strategy with rollercoaster tool that uses WMA as change object and waits for exact match",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"rollerCoaster",
                            "properties":{
                                "pointType":"HIGHest",
                                "obj":"wma",
                                "objConf":{
                                    "period": 3,
                                    "price": "high"
                                },
                                "calcType": "percent",
                                "shiftVal": -50
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
							{High: decimal.New(10, 0)},
							{High: decimal.New(40, 0)},
							{High: decimal.New(30, 0)},
						},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{High: decimal.New(5, 0)},
							{High: decimal.New(5, 0)},
							{High: decimal.New(25, 0)},
						},
					},
				},
				{ // to confirm that tool was not reset
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{High: decimal.New(5, 0)},
							{High: decimal.New(5, 0)},
							{High: decimal.New(25, 0)},
						},
					},
				},
			},
			CandlesNeeded: 3,
		},
		{
			Name: "Strategy with rollercoaster tool that uses ClosePrice as change object and waits for exact match",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"rollerCoaster",
                            "properties":{
                                "pointType":"LOWEST",
                                "obj":"close",
                                "calcType": "UNITS",
                                "shiftVal": 10
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
						},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(20, 0)},
						},
					},
				},
				{ // to confirm that tool was not reset
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(20, 0)},
						},
					},
				},
			},
			CandlesNeeded: 1,
		},
		{
			Name: "Strategy with rollercoaster tool that uses AskPrice as change object and waits for exact match",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"rollercoaster",
                            "properties":{
                                "pointType":"lowest",
                                "obj":"ask",
                                "calcType": "PERCENT",
                                "shiftVal": 10
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
						Ticker: exchange.TickerData{AskPrice: decimal.New(10, 0)},
					},
				},
				{
					Res:       false,
					ShouldErr: false,
					Data: exchange.Data{
						Ticker: exchange.TickerData{AskPrice: decimal.New(9, 0)},
					},
				},
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Ticker: exchange.TickerData{AskPrice: decimal.RequireFromString("9.9")},
					},
				},
				{ // to confirm that tool was not reset
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Ticker: exchange.TickerData{AskPrice: decimal.RequireFromString("9.9")},
					},
				},
			},
			CandlesNeeded: 0,
		},
		{
			Name: "Strategy with macd tool",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"MACD",
                            "properties":{
                                "diff": 1.5,
                                "ema1Period":3,
                                "ema2Period":4,
                                "signalPeriod":2,
                                "price": "close",
                                "calcType": "units",
                                "cond": "belowORequal"
                            }
                        }
                    }
                }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.RequireFromString("10")},
							{Close: decimal.RequireFromString("11")},
							{Close: decimal.RequireFromString("12")},
							{Close: decimal.RequireFromString("13")},
							{Close: decimal.RequireFromString("14")},
							{Close: decimal.RequireFromString("15")},
							{Close: decimal.RequireFromString("10")},
							{Close: decimal.RequireFromString("11")},
							{Close: decimal.RequireFromString("12")},
							{Close: decimal.RequireFromString("13")},
							{Close: decimal.RequireFromString("14")},
							{Close: decimal.RequireFromString("15")},
						},
					},
				},
			},
			CandlesNeeded: 12,
		},
		{
			Name: "Strategy with rsi tool",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"RSI",
                            "properties":{
                                "period": 14,
                                "price": "LOW",
                                "levelVal": 42,
                                "cond": "below"
                            }
                        }
                    }
                }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Low: decimal.RequireFromString("44.34")},
							{Low: decimal.RequireFromString("44.09")},
							{Low: decimal.RequireFromString("44.15")},
							{Low: decimal.RequireFromString("43.61")},
							{Low: decimal.RequireFromString("44.33")},
							{Low: decimal.RequireFromString("44.83")},
							{Low: decimal.RequireFromString("45.10")},
							{Low: decimal.RequireFromString("45.42")},
							{Low: decimal.RequireFromString("45.84")},
							{Low: decimal.RequireFromString("46.08")},
							{Low: decimal.RequireFromString("45.89")},
							{Low: decimal.RequireFromString("46.03")},
							{Low: decimal.RequireFromString("45.61")},
							{Low: decimal.RequireFromString("46.28")},
							{Low: decimal.RequireFromString("46.28")},
							{Low: decimal.RequireFromString("46")},
							{Low: decimal.RequireFromString("46.03")},
							{Low: decimal.RequireFromString("46.41")},
							{Low: decimal.RequireFromString("46.22")},
							{Low: decimal.RequireFromString("45.64")},
							{Low: decimal.RequireFromString("46.21")},
							{Low: decimal.RequireFromString("46.25")},
							{Low: decimal.RequireFromString("45.71")},
							{Low: decimal.RequireFromString("46.45")},
							{Low: decimal.RequireFromString("45.78")},
							{Low: decimal.RequireFromString("45.35")},
							{Low: decimal.RequireFromString("44.03")},
							{Low: decimal.RequireFromString("44.18")},
						},
					},
				},
			},
			CandlesNeeded: 28,
		},
		{
			Name: "Strategy with stoch tool",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"Stoch",
                            "properties":{
                                "KPeriod": 2,
                                "DPeriod": 3,
                                "levelVal": 56,
                                "cond": "below"
                            }
                        }
                    }
                }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{
								High:  decimal.RequireFromString("12"),
								Low:   decimal.RequireFromString("11"),
								Close: decimal.RequireFromString("9"),
							},
							{
								High:  decimal.RequireFromString("9"),
								Low:   decimal.RequireFromString("8"),
								Close: decimal.RequireFromString("10"),
							},
							{
								High:  decimal.RequireFromString("13"),
								Low:   decimal.RequireFromString("7"),
								Close: decimal.RequireFromString("12"),
							},
							{
								High:  decimal.RequireFromString("10"),
								Low:   decimal.RequireFromString("8"),
								Close: decimal.RequireFromString("9"),
							},
						},
					},
				},
			},
			CandlesNeeded: 4,
		},
		{
			Name: "Strategy with trailing trends tool that uses SMA",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"trailingTrends",
                            "properties":{
                                "backIndex": 2,
                                "diff": 2,
                                "obj": "sma",
                                "objConf":{
                                    "period": 3,
                                    "price":"close"
                                },
                                "cond": "equal",
                                "calcType": "units"
                            }
                        }
                    }
                }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.RequireFromString("1")},
							{Close: decimal.RequireFromString("2")},
							{Close: decimal.RequireFromString("3")},
							{Close: decimal.RequireFromString("4")},
							{Close: decimal.RequireFromString("5")},
						},
					},
				},
			},
			CandlesNeeded: 5,
		},
		{
			Name: "Strategy with trailing trends tool that uses candle HighPrice",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"trailingTrends",
                            "properties":{
                                "backIndex": 6,
                                "diff": 600,
                                "obj": "HIGH",
                                "cond": "equal",
                                "calcType": "percent"
                            }
                        }
                    }
                }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{High: decimal.RequireFromString("1")},
							{High: decimal.RequireFromString("2")},
							{High: decimal.RequireFromString("3")},
							{High: decimal.RequireFromString("4")},
							{High: decimal.RequireFromString("5")},
							{High: decimal.RequireFromString("6")},
							{High: decimal.RequireFromString("7")},
						},
					},
				},
			},
			CandlesNeeded: 7,
		},
		{
			Name: "Strategy with BB tool that uses SMA as middle band",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"bb",
                            "properties":{
                                "band": "upper",
                                "period": 3,
                                "stdev": 2,
                                "price": "low",
                                "maType": "SMA",
                                "cond": "above",
                                "calcType": "UNITS",
                                "shiftVal": -1,
                                "obj": "last"
                            }
                        }
                    }
                }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Low: decimal.New(1, 0)},
							{Low: decimal.New(2, 0)},
							{Low: decimal.New(3, 0)},
						},
						Ticker: exchange.TickerData{LastPrice: decimal.RequireFromString("3.6")},
					},
				},
			},
			CandlesNeeded: 3,
		},
		{
			Name: "Strategy with MASpread tool that uses SMAs",
			JSON: `{
                    "name":"strategy",
                    "seq":"test1",
                    "outcomes":[
                        {
                            "type":"sandbox"
                        }
                    ],
                    "tools":{
                        "test1":{
                            "type":"maspread",
                            "properties":{
                                "spread": 0.5,
                                "baseMA": 2,
                                "ma1":{
                                    "maType": "sma",
                                    "period": 3,
                                    "price": "close"
                                },
                                "ma2":{
                                    "maType": "sma",
                                    "period": 4,
                                    "price": "CLOSE"
                                },
                                "cond": "equal",
                                "calcType": "units"
                            }
                        }
                    }
                }`,
			JSONErr: false,
			Cycles: []cycleRes{
				{
					Res:       true,
					ShouldErr: false,
					Data: exchange.Data{
						Candles: []exchange.Candle{
							{Close: decimal.New(1, 0)},
							{Close: decimal.New(2, 0)},
							{Close: decimal.New(3, 0)},
							{Close: decimal.New(4, 0)},
						},
					},
				},
			},
			CandlesNeeded: 4,
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
			}
		})
	}
}
