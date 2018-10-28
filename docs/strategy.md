## Strategies configuration
In order to properly work, strategy's configuration file must have:
* Unique file name with a '-strat' suffix;
* Sequence of tools' unique IDs with valid separation keywords/signs (JSON:"seq", string);
* Outcome(s) (JSON:"outcomes", array of custom objects);
* Tool(s) (JSON:"tools", map of custom objects);

#### Strategy's JSON file structure:
```json
{
    "seq": "tool1 and tool2",
    "outcomes": [],
    "tools": {
        "tool1": {},
        "tool2": {}
    }
}
```

### General strategy rules:
* Only one outcome of Buy, Sell, DCA group can be used per strategy (to avoid orders collisions);
* All outcomes can be used only once per strategy (Buy, Sell, DCA, Telegram, etc), multiple combinations are allowed though (e.g. Buy + Telegram);
* All tools IDs and strategy name in the same strategy file must be unique (tools IDs and strategy name are checked at the same level);
* All strategies names across all strategies files must be unique;
* All tools IDs and strategies names must match the following regexp: `/^[a-zA-Z0-9./#+-]*$/` (lower and upper cased alphanumerics and symbols:  . / # + -);

### Outcomes configuration:
Possible outcomes:
* Buy ("buy")
* Sell ("sell")
* DCA ("dca")
* Telegram ("telegram")
* Sandbox ("sandbox")

Each outcome (in strategy's outcomes' array) must have:
* Type (JSON:"type", string) specifies which type of outcome is the properties field for;
* Properties (JSON:"properties", custom object) contains specified outcome data.

#### Outcome JSON structure:
```json
{
    "type": "buy",
    "properties": {}
}
```

### Outcomes types and their properties:
Which outcome properties can be used depends on the type of the tool. Some outcomes
might have similar properties some might not.

##### List of outcomes and their properties' structures:
1. Buy outcome ("buy") allows the bot to place a buy order when all strategy's tools return true. Can be used only in buy mode.

    * ##### Outcome properties:
        * Price (JSON:"price", string) specifies which price should be used when placing order. Possible options:
            * last, ask, bid (all of these values will be taken from **the latest ticker**);
        * Calc type (JSON:"calcType", string) specifies how amount field should be interpreted. Possible options:
            * counterPercent - calculates amount from counter asset by using amount field as percent value (how much percent of counter asset to deduct from counter asset balance);
            * counterUnits - uses amount field as counter asset amount (how much of counter asset to deduct from counter asset balance);
            * baseUnits - uses amount field as base asset amount (how much base asset to buy);
        * Amount (JSON:"amount", float) specifies value that will be used to calculate amount.

Buy outcome JSON example:
```json
{
    "type": "buy",
    "properties": {
        "price":"ask",
        "calcType": "counterPercent",
        "amount": 10
    }
}
```

2. Sell outcome ("sell") allows the bot to place a sell order when all strategy's tools return true. Can be used only in sell mode. Will empty all balance.

    * ##### Outcome properties:
        * Price (JSON:"price", string) specifies which price should be used when placing order. Possible options:
            * last, ask, bid (all of these values will be taken from **the latest ticker**);

Sell outcome JSON example:
```json
{
    "type": "sell",
    "properties": {
        "price":"bid"
    }
}
```

3. DCA outcome ("dca") allows the bot to place a 	
recurrent buy orders when all strategy's tools return true. Can be used only in sell mode (buy strategy/outcome must be used before DCA).

    * ##### Outcome properties:
        * Repeat (JSON:"repeat", int) specifies how many times should this outcome be activated (each time strategy must return true).
        * Price (JSON:"price", string) specifies which price should be used when placing order. Possible options:
            * last, ask, bid (all of these values will be taken from **the latest ticker**);
        * Calc type (JSON:"calcType", string) specifies how amount field should be interpreted. Possible options:
            * counterPercent - calculates amount from counter asset balance by using amount field as percent value (how much percent of counter asset to deduct from counter asset balance);
            * counterUnits - uses amount field as counter asset amount (how much of counter asset to deduct from counter asset balance);
            * basePercent - calculates amount from base asset balance by using amount field as percent value (how much more percent of base asset to buy);
            * baseUnits - uses amount field as base asset amount (how much base asset to buy);
        * Amount (JSON:"amount", float) specifies value that will be used to calculate amount.

DCA outcome JSON example:
```json
{
    "type": "dca",
    "properties": {
        "repeat":5,
        "price":"ask",
        "calcType": "counterPercent",
        "amount": 3
    }
}
```

4. Telegram outcome ("telegram") allows the bot to send notification to Telegram.

    * ##### Outcome properties:
        * Selection type (JSON:"selection", string) specifies how should the messages be selected. Possible options:
            * rotating - uses all messages one by one (restarts when end is reached);
            * random - random message selection;
        * Messages (JSON:"messages", array of strings) specifies messages that should be sent to Telegram.

Telegram outcome JSON example:
```json
{
    "type": "telegram",
    "properties": {
        "selection":"rotating",
        "messages":["hello", "what's up"]
    }
}
```

5. Sandbox outcome ("sandbox") allows the bot to run and check strategies with real data without making any external actions. When sandbox outcome is used,
no other outcomes can be used. This outcome doesn't have any properties.

Sandbox outcome JSON example:
```json
{
    "type": "sandbox"
}
```

### Tools configuration:
Possible tools:
* BuyPrice ("buyPrice")
* Simple Change ("simpleChange");
* RollerCoaster ("rollercoaster");
* RSI ("rsi");
* MACD ("macd");
* Stochastic ("stoch");
* Bollinger Bands ("bb");
* Moving Averages Spread ("maSpread");
* Trailing Trends ("trailingTrends");

Each tool (in the strategy's tools' map) must have:
* Type (JSON:"type", string) specifies which type of tool is the properties field for;
* Properties (JSON: "properties", custom object, depends on type field) contains specified
    tool actions data;

#### Tool's JSON structure:
```json
{
    "type": "simpleChange",
    "properties": {}
}
```

### Tools types and their properties:
Which tool properties can be used depends on the type of the tool. Some tools
might have similar properties some might not.

##### List of tools and their properties' structures:
1. Buy price checking tool ("buyPrice") waits until [averaged] buy price matches specified conditions with one of the ticker values (you can check how much has the price dropped/increased). Note: if buy price does not exist i.e. asset was not bought yet, tool will always return true.
    * ##### Tool properties that specify which exchange/data values to   follow:
        * Change object (JSON:"obj", string) specifies the value type that needs to be compared with buy price. Possible options:
            * last, ask, bid (all of these values will be taken from ** the latest ticker**);

    * ##### Tool properties that specify how the ticker price should have changed from buy price to allow the bot to act:
        * Shift value (JSON:"shiftVal", float) specifies how much should the buy price be 'shifted' to create the new point that needs to be later on reached by the ticker price. Positive values   increase cached value, negative - decrease.
        * Calc type (JSON:"calcType") specifies how the shift should be made.
        Possible options:
            * percent - increases/decreases the buy price by x (x in this case is shift value) percent;
            * units - increases/decreases the buy price by x (x in this case is shift value) units (simple addition/subtraction);

    * ##### Tool conditions:
        * Cond (JSON:"cond", string) specifies what type of condition should be formed. Possible options:
            * equal - ticker price and buy price (with shift calculations) should be exactly the same;
            * above - ticker price should be above buy price (with shift calculations);
            * aboveOrEqual - ticker price should be above or equal to buy price (with shift calculations);
            * below - ticker price should be below buy price (with shift calculations);
            * belowOrEqual - ticker price should be below or equal to buy price (with shift calculations);
            * aboveOrBelow - ticker price should be above or below to buy price (with shift calculations);

Buy price tool JSON example:
```json
{
    "type": "buyPrice",
    "properties": {
        "obj": "last",
        "calcType": "units",
        "shiftVal": 5.5,
        "cond": "equal"
    }
}
```

In this example 'last' ticker price must be exactly 5.5 units below buy price.

2. Simple change tool ("simpleChange") caches specified value at the start of strategy's usage and waits until the latest value matches the specified conditions.

    * ##### Tool properties that specify which exchange/data values to   follow:
        * Change object (JSON:"obj", string) specifies the value type that needs to be cached and later on checked. Possible options:
            * open, high, low, close (all of these values will be taken from ** the latest candle**);
            * last, ask, bid, 24hrPercent, baseVolume, counterVolume (all of these values will be taken from ** the latest ticker**);
            * sma, ema, wma;
        * Change object config (JSON:"objConf", custom object) specifies the configuration properties to properly use the specified change object. **Only needed when change object is one of the moving averages (SMA, EMA, WMA)**. Change object config (when one of the MAs is used):
            * Period (JSON: "period", int) specifies how many candles should be used to calculate specified MA;
            * Price (JSON: "price", string) specifies which candle price value should be used. Possible options: open, high, low, close;

    * ##### Tool properties that specify how the initial value should have changed to allow the bot to act:
        * Shift value (JSON:"shiftVal", float) specifies how much should the cached value be 'shifted' to create the new point that needs to be later on reached by a new change object value. Positive values   increase cached value, negative - decrease.
        * Calc type (JSON:"calcType") specifies how the shift should be made.
        Possible options:
            * percent - increases/decreases the cached change object value by x (x in this case is shift value) percent;
            * units - increases/decreases the cached change object value by x (x in this case is shift value) units (simple addition/subtraction);
            * fixed - will not change the initial change object value, instead, uses shift value as a 'cached value'. **Can be used to achieve PingPong strategy (from v1.x.x versions) results**;

    * ##### Tool conditions:
        * Cond (JSON:"cond", string) specifies what type of condition should be formed. Possible options:
            * equal - latest change object value and cached value (with shift calculations) should be exactly the same;
            * above - latest change object value should be above cached value (with shift calculations);
            * aboveOrEqual - latest change object value should be above or equal to cached value (with shift calculations);
            * below - latest change object value should be below cached value (with shift calculations);
            * belowOrEqual - latest change object value should be below or equal to cached value (with shift calculations);
            * aboveOrBelow - latest change object value should be above or below to cached value (with shift calculations);

Simple change tool JSON example:
```json
{
    "type": "simpleChange",
    "properties": {
        "obj": "sma",
        "objConf": {
            "period": 3,
            "price": "close"
        },
        "calcType": "units",
        "shiftVal": 5.5,
        "cond": "aboveOrEqual"
    }
}
```

 With the config above, let's say the latest 3 candles' closing prices are: 100, 110, 120 - SMA of 3 candles period would be 110. At this point the new SMA value is cached and the bot waits until the new SMA value will be 115.5 or above (cached value 110 + shift value 5.5).

---

3. RollerCoaster tool ("rollercoaster") looks for the lowest/highest point of the specified change object, caches it and waits until the latest change object value increases/decreases by specified percent/units amount.

    * ##### Tool properties that specify which exchange/data values to  follow:
        * Change object (JSON:"obj", string) specifies the value type that needs to be cached and later on checked. Possible options:
            * open, high, low, close (all of these values will be taken from ** the latest candle**);
            * last, ask, bid, 24hrPercent, baseVolume, counterVolume (all of these values will be taken from ** the latest ticker**);
            * sma, ema, wma;
        * Change object config (JSON:"objConf", custom object) specifies the configuration properties to properly use the specified change object. **Only needed when change object is one of the moving averages (SMA, EMA, WMA)**. Change object config (when one of the MAs is used):
            * Period (JSON: "period", int) specifies how many candles should be used to calculate specified MA;
            * Price (JSON: "price", string) specifies which candle price value should be used. Possible options: open, high, low, close;
        * Point type (JSON:"pointType", string) specifies whether the bot should look for highest or lowest point. Possible options:
            * highest - if used, shift value must be negative i.e. bot should wait for value drop;
            * lowest - if used, shift value must be positive i.e. bot should wait for value rise;

    * ##### Tool properties that specify how the initial value should have changed to allow the bot to act:
        * Shift value (JSON:"shiftVal", float) specifies how much should the cached value be 'shifted' to create the new point that needs to be later on reached by a new change object value. Positive values   increase cached value, negative - decrease;
        * Calc type (JSON:"calcType") specifies how the shift should be made.
        Possible options:
            * percent - increases/decreases the cached change object value by x (x in this case is shift value) percent;
            * units - increases/decreases the cached change object value by x (x in this case is shift value) units (simple addition/subtraction);

RollerCoaster tool JSON example:
```json
{
    "type": "rollercoaster",
    "properties": {
        "obj": "sma",
        "objConf": {
            "period": 3,
            "price": "close"
        },
        "pointType": "lowest",
        "calcType": "units",
        "shiftVal": 5.5
    }
}
```

With the config above, let's say the latest 3 candles' closing prices are: 100, 110, 120 - SMA of 3 candles period would be 110. At this point the SMA value is cached as the lowest point and the bot waits until the new SMA will be calculated. If the new SMA will be lower than current cached one, new SMA will be cached as the lowest point, if it will be above current lowest point by 5.5 (115.5) or more the tool returns true.

---

4. RSI tool ("rsi") waits until RSI value matches specified conditions.

    * ##### Tool properties that specify which exchange/data values to  follow:
        * Period (JSON:"period", int) specifies how many candles should be used to calculate specified RSI;
        * Price (JSON:"price", string) specifies which candle price value should be used. Possible options: open, high, low, close;

    * ##### RSI level:
        * Level value (JSON:"levelVal", float) specifies value from 0 to 100 that needs that will be used in conditions with RSI value;

    * ##### Tool conditions:
        * Cond (JSON:"cond", string) specifies what type of condition should be formed. Possible options:
            * equal - user specified level value and RSI should be exactly the same;
            * above - RSI should be above user specified level value;
            * aboveOrEqual - RSI should be above or equal to user specified level value;
            * below - RSI should be below user specified level value;
            * belowOrEqual - RSI should be below or equal to user specified level value;
            * aboveOrBelow - RSI should be above or below to user specified level value;

RSI tool JSON example:
```json
{
    "type": "rsi",
    "properties": {
        "period": 14,
        "price":"close",
        "levelVal": 45,
        "cond": "below"
    }
}
```
This tool will return true when the RSI value of 14 candles is below 45.

---

5. MACD tool ("macd") waits until MACD line and Signal line values spread/difference matches specified conditions. Signal line is used as base when calculating difference with MACD line.

    * ##### Tool properties that specify which exchange/data values to  follow:
        * EMA1Period (JSON:"ema1Period", int) specifies how many candles should be used to calculate first EMA;
        * EMA2Period (JSON:"ema2Period", int) specifies how many candles should be used to calculate second EMA;
        * SignalPeriod (JSON:"signalPeriod", int) specifies how many MACD line values should be used to calculate Signal line;
        * Price (JSON:"price", string) specifies which candle price value should be used. Possible options: open, high, low, close;

    * ##### MACD and Signal lines difference calculation:
        * Difference value (JSON:"diff", float) specifies value that will be used in conditions when checking lines differences.
        Positive values show how much MACD line is above, negative below Signal line.
        * Calc type (JSON:"calcType", string) specifies whether the user specified diff value should be expressed in percent or units format. Possible options:
            * percent;
            * units;

    * ##### Tool conditions:
        * Cond (JSON:"cond", string) specifies what type of condition should be formed. Possible options:
            * equal - user specified diff and bot calculated one (MACD and Signal lines) should be exactly the same;
            * above - bot calculated diff (MACD and Signal lines) should be above user specified diff;
            * aboveOrEqual - bot calculated diff (MACD and Signal lines) should be above or equal to user specified diff;
            * below - bot calculated diff (MACD and Signal lines) should be below user specified diff;
            * belowOrEqual - bot calculated diff (MACD and Signal lines) should be below or equal to user specified diff;
            * aboveOrBelow - bot calculated diff (MACD and Signal lines) should be above or below to user specified diff;

MACD tool JSON example:
```json
{
    "type": "macd",
    "properties": {
        "ema1Period": 3,
        "ema2Period": 2,
        "signalPeriod": 3,
        "price": "close",
        "diff": 3.5,
        "calcType": "percent",
        "cond": "above"
    }
}
```
This tool will return true when MACD line will be more than 3.5 percent above Signal line. MACD line calculation: fast EMA (fewer candles) - slow EMA (more candles).

---

6. Stochastic tool ("stoch") waits until Stoch %D value matches specified conditions.

    * ##### Tool properties that specify which exchange/data values to  follow:
        * KPeriod (JSON:"KPeriod", int) specifies how many candles should be used to calculate K line;
        * DPeriod (JSON:"DPeriod", int) specifies how many candles should be used to calculate D line;

    * ##### Stoch level:
        * Level value (JSON:"levelVal", float) specifies value from 0 to 100 that needs that will be used in conditions with Stoch %D value;

    * ##### Tool conditions:
        * Cond (JSON:"cond", string) specifies what type of condition should be formed. Possible options:
            * equal - Stoch %D value and specified level value should be exactly the same;
            * above - Stoch %D value should be above specified level value;
            * aboveOrEqual - Stoch %D value should be above or equal specified level value;
            * below - Stoch %D value should be below specified level value;
            * belowOrEqual - Stoch %D value should be below or equal to specified level value;
            * aboveOrBelow  - Stoch %D value should be above or below specified level value;

Stoch tool JSON example:
```json
{
    "type": "stoch",
    "properties": {
        "KPeriod": 14,
        "DPeriod": 3,
        "levelVal": 30,
        "cond": "below"
    }
}
```
This tool will return true when Stoch %D value will be below 30.

---

7. BollingerBands tool ("bb") waits until the specified ticker data value matches the conditions with one of bands.
    * ##### Tool properties that specify which exchange/data values to  follow:
        * Ticker data object (JSON:"obj", string) specifies the value type that needs to be checked/compared with one of the bands. Possible options:
            * last, ask, bid (all of these values will be taken from ** the latest ticker**);
        * Band (JSON:"band", string) specifies which BB band should be used in conditions check. Possible options:
            * lower;
            * upper;
        * Period (JSON:"period", int) specifies how many candles should be used to calculate BB;
        * STDEV (JSON:"stdev", float) specifies value that will be used
    	to multiply standard deviation when calculating
    	upper/lower bands. Use 2.0 when in doubt.
        * Price (JSON:"price", string) specifies which candle price value should be used. Possible options: open, high, low, close;
        * MA type (JSON:"maType", string) specifies which MA should
        be used as the middle band. Possible value: sma, ema, wma.

    * ##### Tool properties that specify how the band value should be changed for the bot to act:
        * Shift value (JSON:"shiftVal", float) specifies how much should the band value be 'shifted' to create the new point that needs to be reached by a ticker data value. Positive values increase band value, negative - decrease.
        * Calc type (JSON:"calcType") specifies how the shift should be made.
        Possible options:
            * percent - increases/decreases the band value by x (x in this case is shift value) percent;
            * units - increases/decreases the band value by x (x in this case is shift value) units (simple addition/subtraction);

    * ##### Tool conditions:
        * Cond (JSON:"cond", string) specifies what type of condition should be formed. Possible options:
            * equal - ticker value and specified band (with shift calculation) should be exactly the same;
            * above - ticker value should be above specified band (with shift calculation);
            * aboveOrEqual - ticker value should be above or equal specified band (with shift calculation);
            * below - ticker value should be below specified band (with shift calculation);
            * belowOrEqual - ticker value should be below or equal to specified band (with shift calculation);
            * aboveOrBelow  - ticker value should be above or below specified band (with shift calculation);

BB tool JSON example:
```json
{
    "type": "bb",
    "properties": {
        "obj": "last",
        "band":"lower",
        "period": 3,
        "stdev": 2.0,
        "price":"close",
        "maType":"sma",
        "shiftVal": 10,
        "calcType": "units",
        "cond": "below"
    }
}
```

 With the config above, let's say the lower band value is 100. After applying shift calculations the lower band is 'lifted' to 102. Tool will return true when ticker's last price will be below 102.

---

8. MA Spread tool ("maSpread") waits until the spread/difference between base and non-base MAs meets conditions. ** MAs of the same type cannot have same period values ** .
    * ##### Tool properties that specify which exchange/data values to  follow:
        * BaseMA (JSON:"baseMA", int) specifies which MA (possible values: 1 or 2) should be used as the base one when calculating the spread.
        * MA1 (JSON:"ma1", custom object):
            * MA type (JSON:"maType", string) specifies what type of MA should be used. Possible value: sma, ema, wma.
            * Price (JSON:"price", string) specifies which candle price value should be used. Possible options: open, high, low, close;
            * Period (JSON:"period", int) specifies how many candles should be used to calculate BB;
        * MA2 (JSON:"ma2", custom object):
            * MA type (JSON:"maType", string) specifies what type of MA should be used. Possible value: sma, ema, wma.
            * Price (JSON:"price", string) specifies which candle price value should be used. Possible options: open, high, low, close;
            * Period (JSON:"period", int) specifies how many candles should be used to calculate BB;

    * ##### MAs spread calculation:
        * Spread value (JSON:"spread", float) specifies value that will be used in conditions when checking MAs spread/differences.
        Positive values show how much non-base MA is above, negative below base MA.
        * Calc type (JSON:"calcType", string) specifies whether the user specified spread value should be expressed in percent or units format. Possible options:
            * percent;
            * units;

    * ##### Tool conditions:
        * Cond (JSON:"cond", string) specifies what type of condition should be formed. Possible options:
            * equal - user specified spread and bot calculated one should be exactly the same;
            * above - bot calculated spread should be above user specified spread;
            * aboveOrEqual - bot calculated spread should be above or equal to user specified spread;
            * below - bot calculated spread should be below user specified spread;
            * belowOrEqual - bot calculated spread should be below or equal to user specified spread;
            * aboveOrBelow - bot calculated spread should be above or below to user specified spread;

MA Spread tool JSON example:
```json
{
    "type": "maSpread",
    "properties": {
        "baseMA": 2,
        "ma1":{
            "maType":"sma",
            "period":3,
            "price":"close"
        },
        "ma2":{
            "maType":"sma",
            "period":4,
            "price":"close"
        },
        "spread": 3.0,
        "calcType": "percent",
        "cond": "above"
    }
}
```

This tool will return true when non-base MA will be above base MA and their difference will be 3 percent.

---

9. TrailingTrends tool ("trailingTrends") waits until the value x candles back matches the conditions with the latest candle. Back value is used as base when calculating difference with the latest value.
    * ##### Tool properties that specify which exchange/data values to  follow:
        * Data object (JSON:"obj", string) specifies the value type that needs to be checked/compared. Will be used with both latest and back values. Possible options:
            * open, high, low, close (all of these values will be taken from ** the latest candle**);
            * sma, ema, wma;
        * Data object config (JSON:"objConf", custom object) specifies the configuration properties to properly use the specified data object. **Only needed when data object is one of the moving averages (SMA, EMA, WMA)**. Data object config (when one of the MAs is used):
            * Period (JSON: "period", int) specifies how many candles should be used to calculate specified MA;
            * Price (JSON: "price", string) specifies which candle price value should be used. Possible options: open, high, low, close;
        * Back index (JSON:"backIndex", int) specifies how many candles **before** the latest candles, should the back candle be. First candle's, before the latest one, index is 1.

    * ##### Latest and back value difference calculation:
        * Difference value (JSON:"diff", float) specifies value that will be used in conditions when checking latest and back values differences.
        Positive values show how much latest value is above, negative below back value.
        * Calc type (JSON:"calcType", string) specifies whether the user specified diff value should be expressed in percent or units format. Possible options:
            * percent;
            * units;

    * ##### Tool conditions:
        * Cond (JSON:"cond", string) specifies what type of condition should be formed. Possible options:
            * equal - user specified diff and bot calculated one should be exactly the same;
            * above - bot calculated diff should be above user specified spread;
            * aboveOrEqual - bot calculated diff should be above or equal to user specified diff;
            * below - bot calculated diff should be below user specified spread;
            * belowOrEqual - bot calculated diff should be below or equal to user specified diff;
            * aboveOrBelow - bot calculated diff should be above or below to user specified diff;

TrailingTrends tool JSON example:
```json
{
    "type": "trailingTrends",
    "properties": {
        "obj": "sma",
        "objConf": {
            "period": 3,
            "price": "close"
        },
        "backIndex": 2,
        "diff": 15,
        "calcType": "units",
        "cond": "above"
    }
}
```
 With the config above, let's say the latest 3 candles' SMA values are: 300, 310, 320. The back SMA value would be 300 (back index is 2; 0 is the latest value), and the latest value is 320. The tool would return true because the calculated diff is 20 (above specified 15).
