## Config files

Note: users will be able to specify configs/sub configs location when starting the bot, though by default all configs should be placed in `configs` directory which should be in the same directory as bot executable (`sub-configs` directory should be in `configs` directory).
Same goes with strategies, but strategies should reside in their own `strategies` directory which is also in the same directory as bot executable.

---

### Main config:
File name: `main.json`

#### Settings:
* Bot config (JSON:"botConfig", custom object):
    * Cycle delay (JSON: "cycleDelay", int) - specifies the amount of time (in seconds) needed to wait between cycles. Cannot be less than 5.
    * Active pairs (JSON:"activePairs", array of strings) specifies pairs to be used by bot. String format: BASE_COUNTER. Cannot be empty.
    * [Advanced] Stream count (JSON:"streamCount", int) specifies how many concurrent streams should be used. Cannot be less than 1. When in doubt use 6.
    * Side task restarts (JSON:"sideTaskRestarts", int) specifies how many times sellAll/cancelAll tasks should be restarted if error occurs during their execution. Cannot be less than 1.
* Pairs config (JSON:"pairsConfig", custom object):
    * Candle interval (JSON:"candleInterval", int) specifies candle interval in minutes.
    * Order history day count (JSON:"orderHistoryDayCount", int) specifies how many days of order history to retrieve from exchange (calculated from the current day). Cannot be less than 1.
    * Strategies (JSON:"strategies", array of strings) specifies names of strategies that should be used by pair that will use this config. Values must be strategies names (found inside strategies files), not strategies *file* names. Cannot be empty.
    * Cancel open orders (JSON:"cancelOpenOrders", bool) specifies whether the open orders should be canceled after specified time or not.
    * Open orders lifespan (JSON:"openOrdersLifespan", int) specifies how long should the bot wait (in seconds) until it should cancel an open order. Each open order will have their separate lifespan i.e. open orders won't be closed all at once. 'Cancel open orders' must be set to true.

Example:
```json
{
    "botConfig":{
        "cycleDelay": 15,
        "activePairs":["ETH_BTC", "DGB_BTC"],
        "streamCount": 6,
        "sideTaskRestarts": 3
    },
    "pairsConfig": {
        "candleInterval": 300,
        "orderHistoryDayCount": 30,
        "strategies": ["moonLamboMagnet", "panicSell"],
        "cancelOpenOrders": true,
        "openOrdersLifespan": 60
    }
}
```

---

### Remote config:
File name: `remote.json`

#### Settings:
* Exchange driver address (JSON:"exchangeDriverAddress", string) specifies exchange driver address.
* Internal (JSON:"internal", custom object):
    * Username (JSON:"username", string) specifies internal EonBot remote controller username used to authenticate remote connections.
    * Password (JSON:"password", string) specifies internal EonBot remote controller passwrod used to authenticate remote connections.
* Telegram (JSON:"telegram", custom object):
    * Enable (JSON:"enable", bool) specifies whether to enable telegram remote controller or not.
    * Token (JSON:"token", string) specifies Telegram bot token used to authorize EonBot on Telegram.
    * Owner (JSON:"owner", string) specifies EonBot owner's Telegram username, so that only he/she could interact with the bot on Telegram.

Example:
```json
{
    "exchangeDriverAddress":"http://localhost:3000/",
    "internal": {
        "username": "name123",
        "password": "pass123"
    },
    "telegram": {
        "enable": true,
        "token": "telegramToken123",
        "owner": "telegramUser123"
    }
}
```

---

### Sub config:
File directory and name: `sub-configs/yourCustomName-sub.json`

#### Settings:
* Active (JSON:"active", bool) specifies whether this sub config
should be used or not. Useful when you want to have multiple sub configs for one specific pair, but only one is allowed.
* Pairs (JSON:"pairs", array of strings) specifies which pairs should use this sub config. String format: BASE_COUNTER. Cannot be empty.
* Pairs config (JSON:"pairsConfig", custom object):
    * Candle interval (JSON:"candleInterval", int) specifies candle interval in minutes.
    * Order history day count (JSON:"orderHistoryDayCount", int) specifies how many days of order history to retrieve from exchange (calculated from the current day). Cannot be less than 1.
    * Strategies (JSON:"strategies", array of strings) specifies names of strategies that should be used by pair that will use this config. Values must be strategies names (found inside strategies files), not strategies *file* names. Cannot be empty.
    * Cancel open orders (JSON:"cancelOpenOrders", bool) specifies whether the open orders should be canceled after specified time or not.
    * Open orders lifespan (JSON:"openOrdersLifespan", int) specifies how long should the bot wait (in seconds) until it should cancel an open order. Each open order will have their separate lifespan i.e. open orders won't be closed all at once. 'Cancel open orders' must be set to true.

Example:
```json
{
    "active": true,
    "pairs": ["ETH_BTC", "DGB_BTC"],
    "pairsConfig": {
        "candleInterval": 300,
        "orderHistoryDayCount": 20,
        "strategies": ["allInOne", "awesomeDCA"],
        "cancelOpenOrders": true,
        "openOrdersLifespan": 30
    }
}
```
