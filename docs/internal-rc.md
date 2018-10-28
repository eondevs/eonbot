### Notes:
* Arrays will be returned in ascending order. Oldest first, newest last. Exception - open orders.
* If error occurs, >= 400 HTTP status code will be returned with
JSON body containing error description:
```json
{
  "error":"action cannot be performed"
}
```
* All pairs (sent and received) must be in BASE_COUNTER format.
* All timestamps (sent and received) must be in RFC3999 format.
* To preserve precision, all floats will be returned in a string format.
* Bot uses '[Basic](https://en.wikipedia.org/wiki/Basic_access_authentication)' authentication method, so each
request must include username and password, if username/password is invalid or not present,
401 error code will be returned.
---

## Bot RC HTTP endpoints:

### Bot data endpoints:

#### Retrieving bot version:
* `GET /bot/version` - retrieves bot version code and name.     
Request parameters: none;
Request JSON body: none;    
Response JSON body:      
```json
{
  "versionCode":"2.0",
  "versionName":"Arya"
}
```

---

#### Retrieving bot's server time:
* `GET /bot/time` - retrieves bot's server time.    
Request parameters: none;
Request JSON body: none;    
Response JSON body:      
```json
{
  "time": "2006-01-02T15:04:05Z"
}
```

---

#### Retrieving bot's orders count since start:
* `GET /bot/orders/since-start` - retrieves bot's orders since start.   
Request parameters: none;
Request JSON body: none;    
Response JSON body:      
```json
{
  "count": 42
}
```

---

#### Retrieving bot's orders:
* `GET /bot/orders?pair=ETH_BTC&start=2006-01-02T15:04:05Z&end=2006-01-02T15:04:05Z` - retrieves bot's orders.    
Request parameters:
    * [optional] 'pair' - specifies which pair's orders should be returned;
    * 'start' - specifies starting timestamp;   
    * 'end' - specifies ending timestamp;   
Request JSON body: none.      
Response JSON body:
```json
{
    "ETH_BTC":[
        {
            "timeStamp": "2006-01-02T15:04:05Z",
            "orderID": "123456",
            "isFilled": true,
            "amount": "20.11",
            "rate": "10023.1234",
            "side": 1,
            "strategy":"awesomeStrat"
        },
        {
            "timeStamp": "2006-01-02T15:04:05Z",
            "orderID": "12345126",
            "isFilled": true,
            "amount": "202.111",
            "rate": "1002123.1234",
            "side": 2,
            "strategy":"awesomeStrat"
        }
    ]
}
```

---

#### Retrieving bot's orders count:
* `GET /bot/orders/count?pair=ETH_BTC` - retrieves orders count.
Request parameters:
     * [optional] 'pair' - specifies which pair's count should be returned;     
Request JSON body: none.      
Response JSON body: 
```json
{
  "count": 42
}
```

---

#### Retrieving bot's activity on hours:
* `GET /bot/orders/hours-activity?pair=ETH_BTC&days=7&end=2006-01-02T15:04:05Z` - retrieves bot's orders counts on hours.    
Request parameters:
    * [optional] 'pair' - specifies which pair's activity should be returned;
    * 'days' - specifies how many days (from the ending date) counts to return;
    * 'end' - specifies ending timestamp;   
Request JSON body: none.      
Response JSON body: 
```json
[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24]
```

---

#### Retrieving bot's cycle's snapshot:
* `GET /bot/cycles?pair=ETH_BTC&id=123` - retrieves cycle's snapshot.     
Request parameters:
    * 'pair' - specifies which pair's activity should be returned;  
    * 'id' - specifies which cycle's snapshot to return, if not specified or negative value is used returns latest;
Request JSON body: none.  
Response JSON body: specified in "Pair cycle snapshot" section below.

---

#### Retrieving saved bot's cycles ids:
* `GET /bot/cycles/ids?pair=ETH_BTC` - retrieves saved cycles count.    
Request JSON body: none.      
Response JSON body: 
```json
[
    12345,
    12346,
    12347
]
```

---

### Workflow endpoints:

#### Retrieving bot's state:
* `GET /workflow/state` - retrieves current bot's state.   
Request parameters: none.    
Request JSON body: none.      
Response JSON body: 
```json
{
  "state": "running",
  "cause": 3,
  "activationTime": "2006-01-02T15:04:05Z"
}
```
Fields explanation:
* 'state' specifies what kind of state is currently active, this field can have two options:
    * running;
    * idle;
* 'cause' specifies what caused current state activation, this field's options:
    * 0 - bot init;
    * 1 - one of the configs was modified externally;
    * 2 - there was a problem (re)loading/parsing one of the configs;
    * 3 - remote controller activated this state;
    * 4 - eon auth problems (auth code invalid, etc);
* 'activationTime' specifies when this state was activated.

---

#### Starting the bot:
* `POST /workflow/start` - starts the bot.      
Request parameters: none.   
Request JSON body:
```json
{
    "sellAll": true,
    "cancelAll": true
}
```
Response JSON body:
```json
{
    "action": "start",
    "status": 2
}
```
Fields explanation:
* Action specifies what kind of action was made.
* Status specifies action result and its option are:
    * 0 - bot is already running;
    * 1 - start action is already initialized/pending;
    * 2 - start was activated;

---

#### Stopping the bot:
* `POST /workflow/stop` - stops the bot.    
Request parameters: none.   
Request JSON body:
```json
{
    "sellAll": true,
    "cancelAll": true,
    "kill": false
}
```
* 'kill' completely stops bot process. After the stop user would have to start the process again manually.

Response JSON body:
```json
{
    "action": "stop",
    "status": 2
}
```
Fields explanation:
* Action specifies what kind of action was made.
* Status specifies action result and its option are:
    * 0 - bot is already idle;
    * 1 - stop action is already initialized/pending;
    * 2 - stop was activated;
    
---

#### Restarting the bot:
* `POST /workflow/restart` - restarts the bot.
Request parameters: none.   
Request JSON body:
```json
{
    "sellAll": true,
    "cancelAll": true
}
```
Response JSON body: none.


---

### Configs endpoints:

#### Retrieving configs summary:
* `GET /configs/summary` - retrieves configs summary.   
Request parameters: none.   
Request JSON body: none.   
Response JSON body:
```json
{
    "auth": true,
    "main": true,
    "remote": true,
    "subConfigs":["array", "of", "subConfigs", "file", "names"],
    "strategies":["array", "of", "strategies", "file", "names"]
}
```

#### Updating main config:
* `PUT /configs/main` - updates main config.    
Request parameters: none.   
Request JSON body: basic main config json object.       
Response JSON body: none.    

---

#### Retrieving main config:
* `GET /configs/main` - retrieves main config.  
Request parameters: none.   
Request JSON body: none.    
Response JSON body: basic main config json object.

---

#### Updating remote config:
* `PUT /configs/remote` - updates remote config.    
Request parameters: none.   
Request JSON body: basic remote config json object.       
Response JSON body: none.    

---

#### Retrieving remote config:
* `GET /configs/remote` - retrieves remote config.  
Request parameters: none.   
Request JSON body: none.    
Response JSON body: basic remote config json object.

---

#### Updating sub config:
* `PUT /configs/sub` - updates sub config.  
Request parameters: none.   
Request JSON body (config field contains basic sub-config json object):      
```json
{
    "fileName": "subConfig123",
    "config":{}
}
```
Response JSON body: none.

---

#### Retrieving sub config:
* `GET /configs/sub?fileName=sub123` - retrieves sub config.    

Request parameters: 
* 'fileName' specifies sub config file name (without file extension);   

Request JSON body: none.        
Response JSON body: basic sub-config json object.

---

#### Removing sub config:
* `DELETE /configs/sub?fileName=sub123` - removes sub config.   

Request parameters: 
* 'fileName' specifies sub config file name (without file extension);   
    
Request JSON body: none.  
Response JSON body: none.  

---

#### Updating strategy:
* `PUT /configs/strategy` - updates strategy.  
Request parameters: none.   
Request JSON body (config field contains basic strategy json object):      
```json
{
    "fileName": "strategy123",
    "config":{}
}
```
Response JSON body: none.

---

#### Retrieving strategy:
* `GET /configs/strategy?fileName=strategy123` - retrieves strategy.    

Request parameters: 
* 'fileName' specifies strategy file name (without file extension);

Request JSON body: none.        
Response JSON body: basic strategy json object.

---

#### Removing strategy:
* `DELETE /configs/strategy?fileName=sub123` - removes strategy.   

Request parameters: 
* 'fileName' specifies strategy file name (without file extension);

Request JSON body: none.  
Response JSON body: none.  

---

### Exchange endpoints:

#### Updating API keys/secrets:
* `POST /exchange/api-info` - updates API credentials list
with the provided array.    
Request parameters: none.     
Request JSON body:  
```json
[
  {
    "key":"key1",
    "secret":"secret1"
  },
  {
    "key":"key2",
    "secret":"secret2"
  },
  {
    "key":"key3",
    "secret":"secret3"
  }
]
```
Response JSON body: none;

---

#### Retrieving API keys/secrets:
* `GET /exchange/api-info` - retrieves API credentials list.  
Request parameters: none;
Request JSON body: none;    
Response JSON body:      
```json
[
  {
    "key":"key1",
    "secret":"secret1"
  },
  {
    "key":"key2",
    "secret":"secret2"
  },
  {
    "key":"key3",
    "secret":"secret3"
  }
]
```

---

#### Retrieving cooldown state info:
* `GET /exchange/cooldown-info` - retrieves cooldown state info.  
Request parameters: none;   
Request JSON body: none;  
Response JSON body: 
```json
{
  "active": true,
  "start": "2006-01-02T15:04:04Z",
  "end": "2006-01-02T15:04:04Z"
}
```

---

#### Retrieving asset pairs:
* `GET /exchange/pairs` - retrieves asset pairs.     
Request parameters: none;   
Request JSON body: none;  
Response JSON body: 
```json
[
    "ETH_BTC",
    "DGB_BTC",
    "LTC_BTC"
]
```

---

#### Retrieving candles intervals:
* `GET /exchange/intervals` - retrieves candles intervals.  
Request parameters: none;   
Request JSON body: none;  
Response JSON body: 
```json
[
  60,
  300,
  900
]
```

#### Retrieving ticker:
* `GET /exchange/ticker?pair=ETH_BTC` - retrieves ticker of specific pair (or all pairs).
Request parameters:
    * [optional] 'pair' - specifies which pair's ticker info should be returned, if not specified all pairs' tickers should be returned;

Request JSON body: none;  
Response JSON body (if pair is specified):     
```json
{
  "lastPrice": 3312.01,
  "askPrice": 3321.03,
  "bidPrice": 3309.1,
  "baseVolume": 874.7,
  "counterVolume": 13.2,
  "dayPercentChange": 3.9
}
```
Response JSON body (if pair is NOT specified):    

```json
{
  "ETH_BTC":{
    "lastPrice": 3312.01,
    "askPrice": 3321.03,
    "bidPrice": 3309.1,
    "baseVolume": 874.7,
    "counterVolume": 13.2,
    "dayPercentChange": 3.9
  },
  "DGB_BTC"{
    "lastPrice": 3312.01,
    "askPrice": 3321.03,
    "bidPrice": 3309.1,
    "baseVolume": 874.7,
    "counterVolume": 13.2,
    "dayPercentChange": 3.9
  }
}
```

---

#### Retrieving candles data:
* `GET /exchange/candles?pair=ETH_BTC&interval=300&end=2006-01-02T15:04:04Z` - retrieves candlestick info of specific pair.  
Request parameters:
    * 'pair' - specifies which pair's candlestick info should be returned;
    * 'interval' - specifies which interval should be used;
    * [optional] 'end' - specifies ending timestamp. If you need latest info, don't use end parameter at all, use it only when you need info at a specific (not latest) point in time;
    * [optional] 'limit' - specifies the minimum amount of candles that should be returned;

Request JSON body: none;  
Response JSON body:     
```json
[
  {
    "timestamp": "2006-01-02T15:04:04Z",
    "open": 230.01,
    "high": 240.1,
    "low": 220.1,
    "close": 235.8,
    "baseVolume": 342.1,
    "counterVolume": 34.5
  },
  {
      "timestamp": "2006-01-02T15:04:04Z",
      "open": 233.01,
      "high": 250.1,
      "low": 230.1,
      "close": 238.8,
      "baseVolume": 362.1,
      "counterVolume": 38.5
    }
]
```

---

#### Retrieving balances:
* `GET /exchange/balances` - retrieves balances.     
Request parameters: none;   
Request JSON body: none;  
Response JSON body: 
```json
{
  "BTC":0.01,
  "ETH":0.1,
  "DGB":231.4,
  "XMR": 21.1
}
```

---

#### Retrieving open orders:
* `GET /exchange/open-orders?pair=ETH_BTC` - retrieves open orders of specific pair.     
Request parameters:
    * 'pair' - specifies which pair's open orders should be returned;

Request JSON body: none;  
Response JSON body:     
```json
{
  "orderID": "asd12345",
  "isFilled": false,
  "amount": 0.2,
  "rate": 332.1,
  "side": "buy"
}
```

---

#### Retrieving order history:
* `GET /exchange/order-history?pair=ETH_BTC&start=2006-01-02T15:04:04Z&end=2006-01-02T15:04:04Z` - retrieves order history of specific pair.     
Request parameters:
    * 'pair' - specifies which pair's order history should be returned;
    * [optional] 'start' - specifies starting timestamp. If not specified, returns max amount of orders;
    * [optional] 'end' - specifies ending timestamp. If you need latest info, don't use end parameter at all, use it only when you need info at a specific (not latest) point in time;

Request JSON body: none;  
Response JSON body:     
```json
[
  {
    "timestamp": "2006-01-02T15:04:04Z",
    "orderID": "asd12345",
    "isFilled": true,
    "amount": 0.2,
    "rate": 332.1,
    "side": "buy"
  },
  {
    "timestamp": "2006-01-02T15:04:04Z",
    "orderID": "asd12345",
    "isFilled": true,
    "amount": 0.2,
    "rate": 332.1,
    "side": "sell"
  }
]
```

---

## WebSockets:
WebSockets are only used to publish events from the bot.    
* `/ws` - connect to websocket.
Events that will be published by the bot:
    * State change (updated state info can be retrieved from `/workflow/state`):
    ```json
    {
        "event":"state-update"
    }
    ```
    * Cooldown activation (updated cooldown info can be retrieved from `/exchange/cooldown-info`):
    ```json
    {
        "event":"cooldown-activation"
    }
    ```
    * Pair cycle end (latest cycle has been added to the database and can be retrieved from `/bot/cycle`):
    ```json
    {
        "event":"pair-cycle-end"
    }
    ```

---

## Pair cycle snapshot:
Pair cycle snapshot will have timestamps that specify when the cycle
was started and completed. 'isSuccessful' field specifies whether the
cycle was successful, if true - 'result' field will be non empty 
(and error empty), if false - 'error' field will be non empty 
(and result empty).       
Example:        
Unsuccessful:       
```json
{
    "startedAt": "2006-01-02T15:04:04Z",
    "completedAt": "2006-01-02T15:04:04Z",
    "isSuccessful": false,
    "error": "exchange prevented order placing"
}
```
Successful:
```json
{
    "startedAt": "2006-01-02T15:04:04Z",
    "completedAt": "2006-01-02T15:04:04Z",
    "isSuccessful": true,
    "result":{
        "type":"open-orders",
        "openOrders":{
            "total":13,
            "open": 3,
            "cancelled": 10
        }
    }
}
```

Result field can have different contents and type:
* Type 'open-orders' - specifies total open orders count, how many are left open and how many were cancelled during that cycle, example of result field with 'open-orders' type:
```json
{
    "type":"open-orders",
    "total": 10,
    "open": 4,
    "cancelled": 6
}
```
* Type 'strategies' - specifies strategies snapshots of that cycle, example of result field with 'strategies' type:
```json
{
    "type":"strategies",
    "snapshots":{
        "awesomeStrat": {
            "condsMet": true,
            "seq": "test1 and test2",
            "tools": {
                "test1":{},
                "test2":{}
            }
        },
        "awesomeStrat2": {
            "condsMet": true,
            "seq": "test1 or test2",
            "tools": {
                "test1":{},
                "test2":{}
            }
        }
    }
}
```

Strategy ('strategies' field one element) fields explanation:
* 'condsMet' specifies whether all conditions were met and strategy executed its outcomes.
* 'seq' specifies strategy's tools sequence.
* 'tools' specifies every tool used in the sequence configuration, snapshot and result (i.e. whether it returned true or not - 'condsMet'). Tool example:     
```json
{
    "type": "simpleChange",
    "properties": {},
    "snapshot":{
        "condsMet": true,
        "data":{
            "changeVal": "123.123",
            "objVal": "120.333"
        }
    }
}
```
* 'type' field specifies tool type, just like in strategy config.
* 'properties' field specifies tool configuration properties (from strategies file, this specific tool section).
* 'snapshot' specifies tool snapshot:
    * 'condsMet' specifies whether tool all conditions were met (true = met).
    * 'data' specifies tool specific snapshot data. Different tools will have different 'data' object fields, 'Tools snapshots data' section explains each tool's snapshot data.

---

#### Tools snapshots data:
1. BuyPrice:
    ```json
    {
        "buyPrice": "123.123",
        "shiftedBuyPrice": "133.123",
        "objVal": "120.333"
    }
    ```
    * 'buyPrice' specifies original buy price returned/averaged from exchange.        
    * 'shiftedBuyPrice' specifies buy price with applied shift calculations.      
    * 'objVal' specifies ticker price which is used to compare with shifted buy price.

2. SimpleChange:
    ```json
    {
        "changeVal": "123.123",
        "objVal": "120.333"
    }
    ```   
    * 'changeVal' specifies cached change object value with applied shift calculations.      
    * 'objVal' specifies latest change object value which is used to compare with shifted cached change object value.

3. RollerCoaster:
    ```json
    {
        "pointVal": "123.123",
        "shiftedPointVal": "133.123",
        "objVal": "120.333"
    }
    ```   
    * 'pointVal' specifies cached lowest/highest point.  
    * 'shiftedPointVal' specifies cached lowest/highest point with applied shift calculations.    
    * 'objVal' specifies latest change object value which is used to compare with shifted point value.

4. RSI:
    ```json
    {
        "rsiVal": "48.123"
    }
    ```   
    * 'rsiVal' specifies current RSI value.

5. MACD:
    ```json
    {
        "diffVal": "23.43",
        "macdLine": "98.123",
        "signalLine": "99.123"
    }
    ```   
    * 'diffVal' specifies difference between current MACD and Signal lines.
    * 'macdLine' - specifies latest MACD line point value.
    * 'signalLine' - specifies latest Signal line point value.

6. Stoch:
    ```json
    {
        "K": "73.12",
        "D": "75.222"
    }
    ```   
    * 'K' specifies current %K value.
    * 'D' specifies current %D value.

7. BollingerBands:
    ```json
    {
        "shiftedBand": "1234.444",
        "objVal": "1230.33",
        "upper": "1233.22",
        "middle": "1200.1",
        "lower": "1194.3"
    }
    ```     
    * 'shiftedBand' specifies upper/lower band value with applied shift calculations.
    * 'objVal' specifies change object value that is used to compare with band with applied shift calculations value.
    * 'upper' specifies current upper band value.
    * 'middle' specifies current middle band value.
    * 'lower' specifies current lower band value.

8. MA Spread:
    ```json
    {
        "spreadVal": "3.2",
        "ma1Val": "1230.02",
        "ma2Val": "1233.22"
    }
    ```   
    * 'spread' specifies spread between MAs value.
    * 'ma1Val' specifies MA1 current value.
    * 'ma2Val' specifies MA2 current value.

9. TrailingTrends:
    ```json
    {
        "diffVal": "3.2",
        "leadObjVal": "1230.02",
        "backObjVal": "1238.02"
    }
    ```  
    * 'diffVal' specifies difference between leading object (latest candle) and object/candle in the back.
    * 'leadObjVal' specifies current leading object / latest candle value.
    * 'backObjVal' specifies current back object / x candles before latest candle value.
