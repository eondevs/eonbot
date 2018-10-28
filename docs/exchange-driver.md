## Exchange driver specifications:

### Notes:
* Arrays should be returned in ascending order. Oldest first, newest last. Exception - open orders.
* If error occurs, >= 400 HTTP status code must be returned with
JSON body containing error description:
```json
{
  "error":"action cannot be performed"
}
```
* All pairs (received and returned) must be in BASE_COUNTER format.
* All timestamps (received and returned) must be in RFC3999 format.
* All intervals (received and returned) must be integers that 
represent seconds. For months use 31 days worth of seconds.
* To preserve precision, all floats will be sent to the exchange driver in a string format. Returned floats can be in either float or string format.
* Orders' "side" field can only have "buy" or "sell" values.
* Candles endpoint have an optional 'limit' field, if this field is not available
in the request sent from bot, use the biggest limit value allowed by exchange.
* Candles and order history endpoints have optional 'end' parameter, if not specified, send current time as the end time value to the exchange to retrieve the most
latest info, then, if exchange supports websockets for this type of data, cache initial data (returned from HTTP endpoint) and update it with websockets data.
* All HTTP endpoints must be implemented as shown below, otherwise bot
won't be able to function with the driver properly.

### REST HTTP Endpoints

#### Updating API keys/secrets:
* `POST /api-info` - updates API credentials list
with the provided array.    
Request parameters: none;     
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
* `GET /api-info` - retrieves API credentials list.  
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

#### Connection testing (aka pinging):
* `GET /ping` - pings exchange. Nothing much should be done here, if exchange responds successfully,
just respond with 200 status code. 
Request parameters: none;   
Request JSON body: none;  
Response JSON body: none;   

---

#### Retrieving cooldown state info:
* `GET /cooldown-info` - retrieves cooldown state info.  
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
If "active" field is false, "until" field can be omitted.   

---

#### Retrieving candles intervals:
* `GET /intervals` - retrieves candles intervals.
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

---

#### Retrieving asset pairs:
* `GET /pairs` - retrieves asset pairs.     
Request parameters: none;   
Request JSON body: none;  
Response JSON body: 
```json
{
  "ETH_BTC":{
    "basePrecision": 8,
    "counterPrecision": 8,
    "minValue": 0.0001,
    "minRate": 0.001,
    "maxRate": 0.002,
    "rateStep": 0.0005,
    "minAmount": 0.0001,
    "maxAmount": 10,
    "amountStep": 0.005
  },
  "DGB_BTC":{
    "basePrecision": 8,
    "counterPrecision": 8,
    "minValue": 0.001,
    "minRate": 0.01,
    "maxRate": 0.02,
    "rateStep": 0.005,
    "minAmount": 0.001,
    "maxAmount": 200,
    "amountStep": 0.05
  }
}
```

If exchange does not provide info for one or more of these fields (e.g. basePrecision), 
don't include it in the response or set it to 0 / negative value (only difined, positive values will be used by eonbot).

---

#### Retrieving ticker:
* `GET /ticker?pair=ETH_BTC` - retrieves ticker of specific pair (or all pairs).
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

#### Retrieving candlestick data:
* `GET /candles?pair=ETH_BTC&interval=300&end=2006-01-02T15:04:04Z` - retrieves candlestick info of specific pair.  
Request parameters:
    * 'pair' - specifies which pair's candlestick info should be returned;
    * 'interval' - specifies which interval should be used;
    * [optional ] 'end' - specifies ending timestamp;
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
* `GET /balances` - retrieves balances.     
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

#### Placing buy order:
* `POST /buy` - places buy order.   
Request parameters: none;   
Request JSON body: 
```json
{
  "pair": "ETH_BTC",
  "rate": "0.00132",
  "amount": "0.12"
}
```
Response JSON body: 
```json
{
  "id":"asd12345"
}
```

---

#### Placing sell order:
* `POST /sell` - places sell order.
Request parameters: none;   
Request JSON body: 
```json
{
  "pair": "ETH_BTC",
  "rate": "0.00132",
  "amount": "0.12"
}
```
Response JSON body: 
```json
{
  "id":"asd12345"
}
```

---

#### Cancelling open order:
* `POST /cancel` - cancels open order.
Request parameters: none;   
Request JSON body: 
```json
{
  "pair": "ETH_BTC",
  "id": "asd12345"
}
```
Response JSON body: none;   

---

#### Retrieving order:
* `GET /order?pair=ETH_BTC&id=asd12345` - retrieves specific order of specific pair.        
Request parameters:
    * 'pair' - specifies which pair's order should be returned;
    * 'id' - specifies which order to return.
    
Request JSON body: none;  
Response JSON body:     
```json
{
  "timestamp": "2006-01-02T15:04:04Z",
  "orderID": "asd12345",
  "isFilled": true,
  "amount": 0.2,
  "rate": 332.1,
  "side": "buy"
}
```

** If order does not exist, it's essential to return 404 status code **

---

#### Retrieving open orders:
* `GET /open-orders?pair=ETH_BTC` - retrieves open orders of specific pair.     
Request parameters:
    * 'pair' - specifies which pair's open orders should be returned;

Request JSON body: none;  
Response JSON body:     
```json
{
  "timestamp: "2006-01-02T15:04:04Z",
  "orderID": "asd12345",
  "isFilled": false,
  "amount": 0.2,
  "rate": 332.1,
  "side": "buy"
}
```

---

#### Retrieving order history:
* `GET /order-history?pair=ETH_BTC&start=2006-01-02T15:04:04Z&end=2006-01-02T15:04:04Z` - retrieves order history of specific pair.     
Request parameters:
    * 'pair' - specifies which pair's order history should be returned;
    * [optional] 'start' - specifies starting timestamp, if not specified, max amount of orders should be returned;
    * [optional] 'end' - specifies ending timestamp;

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