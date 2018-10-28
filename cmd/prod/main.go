package main

import (
	"eonbot/pkg"
	"eonbot/pkg/asset"
	"eonbot/pkg/bot"
	"eonbot/pkg/exchange"
	"eonbot/pkg/file"
	"eonbot/pkg/settings"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ebCMD       = kingpin.New("eonbot", "")
	botCMD      = ebCMD.Command("bot", "Assets analysis and trading bot.").Default()
	exchangeCMD = ebCMD.Command("exchange", "Exchange driver testing tool.")
)

func init() {
	ebCMD.HelpFlag.Short('h')
	ebCMD.Version(pkg.FullVersion())
	ebCMD.VersionFlag.Short('v')
}

func main() {
	switch kingpin.MustParse(ebCMD.Parse(os.Args[1:])) {
	case botCMD.FullCommand():
		botCommand()
	case exchangeCMD.FullCommand():
		exchangeCommand()
	}
}

var (
	/*
		bot commands
	*/

	botVerbose = botCMD.Flag("verbose", "Enable verbose / debug level logging.").
			Short('V').Default("false").Bool()

	botStdout = botCMD.Flag("stdout", "Enable logs printing to stdout.").
			Short('s').Default("false").Bool()

	botLogsDir = botCMD.Flag("logs-dir", "Specify logs directory.").
			PlaceHolder("<dir>").String()

	botJSONLogs = botCMD.Flag("json-logs", "Enable logging in JSON format.").
			Short('J').Default("false").Bool()

	botConfDir = botCMD.Flag("conf-dir", "Specify configs directory.").
			PlaceHolder("<dir>").String()

	botSubsDir = botCMD.Flag("subs-dir", "Specify sub-configs directory.").
			PlaceHolder("<dir>").String()

	botStratDir = botCMD.Flag("strat-dir", "Specify strategies directory.").
			PlaceHolder("<dir>").String()

	botCancelAll = botCMD.Flag("cancel-all", "Cancel all open orders on bot start. Disabled when 'auto-start' is not used.").
			Default("false").Bool()

	botSellAll = botCMD.Flag("sell-all", "Sell all coins on bot start. Disabled when 'auto-start' is not used.").
			Default("false").Bool()

	botAutoStart = botCMD.Flag("auto-start", "Immediately start the bot and not wait for RC to start it. Enables 'sell-all' and 'cancel-all' flags. NOTE: all mandatory configs must be present.").
			Short('a').Default("false").Bool()

	botReloadStop = botCMD.Flag("reload-stop", "Stop the bot if new config contains an error. NOTE: if config files will be removed when the bot is running, the bot will stop (if it's initial cycle, the bot will stop the whole process and print out the error), doesn't matter if this flag is used.").
			Short('r').Default("false").Bool()

	botRCPort = botCMD.Flag("port", "Port to use for internal remote controller.").
			Short('p').Default("8080").Int()

	botHTTPTimeout = botCMD.Flag("http-timeout", "Specify the max amount of time (in seconds) the request should take to go to the exchange driver and receive response.").
			Default("30").Int64()
)

func botCommand() {
	execDir, err := file.ExecDir()
	if err != nil {
		ebCMD.Fatalf("%s", err)
	}

	if *botLogsDir == "" {
		*botLogsDir = path.Join(execDir, "logs")
	}

	if *botConfDir == "" {
		*botConfDir = path.Join(execDir, "configs")
	}

	if *botSubsDir == "" {
		*botSubsDir = path.Join(*botConfDir, "sub-configs")
	}

	if *botStratDir == "" {
		*botStratDir = path.Join(execDir, "strategies")
	}

	conf := settings.Exec{
		Verbose:       *botVerbose,
		Stdout:        *botStdout,
		LogsDir:       *botLogsDir,
		LogsJSON:      *botJSONLogs,
		ConfigsDir:    *botConfDir,
		SubsDir:       *botSubsDir,
		StrategiesDir: *botStratDir,
		CancelAll:     *botCancelAll,
		SellAll:       *botSellAll,
		ReloadStop:    *botReloadStop,
		AutoStart:     *botAutoStart,
		RCPort:        *botRCPort,
		HTTPTimeout:   *botHTTPTimeout,
	}

	if err := bot.Launch(conf); err != nil {
		ebCMD.Fatalf("%s", err)
	}
}

var (
	/*
		exchange commands
	*/

	exchTimeout = exchangeCMD.Flag("timeout", "Exchange driver request timeout in seconds.").
			Default("60").Int64()

	exchAPIInfoLoc = exchangeCMD.Flag("api-loc", "Specify API credentials file location (applies for both downloading and uploading).").
			Short('L').PlaceHolder("<path-to-json-file>").String()

	exchUploadAPI = exchangeCMD.Flag("upload-api", "Upload API credentials from local JSON file.").
			Short('U').Default("false").Bool()

	exchDownloadAPI = exchangeCMD.Flag("download-api", "Download API credentials and save them locally.").
			Short('D').Default("false").Bool()

	exchCooldown = exchangeCMD.Flag("cooldown", "Retrieve cooldown state info.").
			Short('C').Default("false").Bool()

	exchPair = exchangeCMD.Flag("pair", "Specify action pair. Format: BASE_COUNTER.").
			Default("ETH_BTC").String()

	exchRetrieveAll = exchangeCMD.Flag("retrieve-all", "Retrieve info from all endpoints (except API credentials and single order).").
			Short('R').Default("false").Bool()

	exchPairInfo = exchangeCMD.Flag("pair-info", "Retrieve pair info.").
			Short('p').Default("false").Bool()

	exchIntervals = exchangeCMD.Flag("intervals", "Retrieve intervals.").
			Short('i').Default("false").Bool()

	exchTicker = exchangeCMD.Flag("ticker", "Retrieve pair ticker info.").
			Short('t').Default("false").Bool()

	exchCandle = exchangeCMD.Flag("candle", "Retrieve pair one candle info.").
			Short('c').Default("false").Bool()

	exchCandleIndex = exchangeCMD.Flag("candle-index", "Specify candle index, from right to left (0 = latest candle, 1 = candle before the latest, etc).").
			Default("0").Int()

	exchCandleInterval = exchangeCMD.Flag("candle-interval", "Specify candle interval.").
				Default("300").Int()

	exchBalances = exchangeCMD.Flag("balances", "Retrieve all non-zero balances. Overrides '--pair-balance' flag.").
			Short('B').Default("false").Bool()

	exchBalance = exchangeCMD.Flag("pair-balance", "Retrieve pair assets balances.").
			Short('b').Default("false").Bool()

	exchRate = exchangeCMD.Flag("rate", "Specify buy/sell order rate.").
			Default("0.01").Float64()

	exchBuy = exchangeCMD.Flag("buy", "Place a buy order.").
		PlaceHolder("amount").String()

	exchSell = exchangeCMD.Flag("sell", "Place a sell order.").
			PlaceHolder("amount").String()

	exchCancelOrder = exchangeCMD.Flag("cancel", "Cancel specific open order.").
			PlaceHolder("orderID123").String()

	exchOrder = exchangeCMD.Flag("order", "Retrieve specific order.").
			Short('o').PlaceHolder("orderID123").String()

	exchOpenOrders = exchangeCMD.Flag("open-orders", "Retrieve all open orders of a pair. NOTE: only first 5 orders will be displayed.").
			Default("false").Bool()

	exchOrderHistory = exchangeCMD.Flag("hist", "Retrieve order history of a pair. NOTE: only latest 5 orders will be displayed.").
				Short('H').Default("false").Bool()

	exchOrderHistoryDays = exchangeCMD.Flag("hist-days", "Specify order history days count (from the current)").
				Default("7").Int64()

	exchAddress = exchangeCMD.Arg("address", "Exchange driver address.").Required().String()
)

func exchangeCommand() {
	exch := exchange.New(*exchTimeout)

	if err := exch.SetAddress(*exchAddress); err != nil {
		ebCMD.Fatalf("%s", err)
	}

	pair, err := asset.PairFromString(*exchPair)
	if err != nil {
		ebCMD.Fatalf("%s", err)
	}

	var b strings.Builder

	if err := exch.Ping(); err != nil {
		ebCMD.Fatalf("%s", err)
	}
	if b.Len() > 0 {
		b.WriteString("----------\n")
	}
	b.WriteString("Exchange driver successfully responded to a ping request\n")

	if *exchAPIInfoLoc == "" {
		execDir, err := file.ExecDir()
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		*exchAPIInfoLoc = file.JSONPath(execDir, "api-info")
	}

	if *exchUploadAPI {
		var apiInfo []exchange.APIInfo
		if err := file.LoadJSON(*exchAPIInfoLoc, &apiInfo); err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if err := exch.SetAPIInfo(apiInfo); err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}
		b.WriteString(fmt.Sprintf("API info successfully uploaded from %s\n", *exchAPIInfoLoc))
	}

	if *exchDownloadAPI {
		apiInfo, err := exch.GetAPIInfo()
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if err := file.SaveJSON(*exchAPIInfoLoc, apiInfo); err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}
		b.WriteString(fmt.Sprintf("API info successfully downloaded to %s\n", *exchAPIInfoLoc))
	}

	if *exchCooldown || *exchRetrieveAll {
		cooldown, err := exch.GetCooldownInfo()
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		b.WriteString("Cooldown info:\n")
		b.WriteString(fmt.Sprintf(" - Is active: %v\n", cooldown.Active))
		if cooldown.Active {
			b.WriteString(fmt.Sprintf(" - Actived on: %s\n", cooldown.Start.Format(time.RFC1123)))
			b.WriteString(fmt.Sprintf(" - Active until: %s\n", cooldown.End.Format(time.RFC1123)))
		}
	}

	if *exchPairInfo || *exchRetrieveAll {
		pp, err := exch.GetPairs()
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		for _, p := range pp {
			if p.Equal(pair) {
				b.WriteString(fmt.Sprintf("%s pair info:\n", p.String()))
				b.WriteString(fmt.Sprintf(" - Base asset precision: %d\n", p.BasePrecision))
				b.WriteString(fmt.Sprintf(" - Counter asset precision: %d\n", p.CounterPrecision))
				b.WriteString(fmt.Sprintf(" - Minimum rate: %s\n", p.MinRate.String()))
				b.WriteString(fmt.Sprintf(" - Maximum rate: %s\n", p.MaxRate.String()))
				b.WriteString(fmt.Sprintf(" - Rate step: %s\n", p.RateStep.String()))
				b.WriteString(fmt.Sprintf(" - Minimum amount: %s\n", p.MinAmount.String()))
				b.WriteString(fmt.Sprintf(" - Maximum amount: %s\n", p.MaxAmount.String()))
				b.WriteString(fmt.Sprintf(" - Amount step: %s\n", p.AmountStep.String()))
				b.WriteString(fmt.Sprintf(" - Minimum value (rate * amount): %s\n", p.MinValue.String()))
				break
			}
		}
	}

	if *exchIntervals || *exchRetrieveAll {
		intervals, err := exch.GetIntervals()
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		b.WriteString("Intervals:\n")
		for _, interval := range intervals {
			b.WriteString(fmt.Sprintf(" - %d\n", interval))
		}
	}

	if *exchTicker || *exchRetrieveAll {
		ticker, err := exch.GetTicker(pair)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		b.WriteString(fmt.Sprintf("%s pair ticker info:\n", pair.String()))
		b.WriteString(fmt.Sprintf(" - Last price: %s\n", ticker.LastPrice.String()))
		b.WriteString(fmt.Sprintf(" - Ask price: %s\n", ticker.AskPrice.String()))
		b.WriteString(fmt.Sprintf(" - Bid price: %s\n", ticker.BidPrice.String()))
		b.WriteString(fmt.Sprintf(" - 24hr price change: %s%%\n", ticker.DayPercentChange.String()))
		b.WriteString(fmt.Sprintf(" - Base volume: %s\n", ticker.BaseVolume.String()))
		b.WriteString(fmt.Sprintf(" - Counter volume: %s\n", ticker.CounterVolume.String()))
	}

	if *exchCandle || *exchRetrieveAll {
		candles, err := exch.GetCandles(pair, *exchCandleInterval, time.Now().Add(-time.Second*time.Duration(*exchCandleInterval)*time.Duration(*exchCandleIndex)), 1)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		if len(candles) == 0 {
			b.WriteString("Candle info was not returned\n")
		} else {
			candle := candles[0]

			b.WriteString(fmt.Sprintf("%s pair candle info:\n", pair.String()))
			b.WriteString(fmt.Sprintf(" - Timestamp: %s\n", candle.Timestamp.Format(time.RFC1123)))
			b.WriteString(fmt.Sprintf(" - Open price: %s\n", candle.Open.String()))
			b.WriteString(fmt.Sprintf(" - High price: %s\n", candle.High.String()))
			b.WriteString(fmt.Sprintf(" - Low price: %s\n", candle.Low.String()))
			b.WriteString(fmt.Sprintf(" - Close price: %s\n", candle.Close.String()))
			b.WriteString(fmt.Sprintf(" - Base volume: %s\n", candle.BaseVolume.String()))
			b.WriteString(fmt.Sprintf(" - Counter volume: %s\n", candle.CounterVolume.String()))
		}
	}

	if *exchBalances || *exchRetrieveAll {
		balances, err := exch.GetBalances()
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		b.WriteString("All non-zero balances:\n")
		var nonZeroExists bool
		for k, bal := range balances {
			if bal.GreaterThan(decimal.Zero) {
				b.WriteString(fmt.Sprintf(" - %s: %s\n", k, bal.String()))
				if !nonZeroExists {
					nonZeroExists = true
				}
			}
		}

		if !nonZeroExists {
			b.WriteString("Non-zero balances don't exist\n")
		}
	} else if *exchBalance {
		balances, err := exch.GetBalances()
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		baseBalance := balances[string(pair.Base)]
		counterBalance := balances[string(pair.Counter)]
		b.WriteString(fmt.Sprintf("%s pair assets balances:\n", pair.String()))
		b.WriteString(fmt.Sprintf(" - %s: %s\n", pair.Base, baseBalance))
		b.WriteString(fmt.Sprintf(" - %s: %s\n", pair.Counter, counterBalance))
	}

	rate := decimal.NewFromFloat(*exchRate)
	if err != nil {
		ebCMD.Fatalf("%s", err)
	}

	if *exchBuy != "" {
		amount, err := decimal.NewFromString(*exchBuy)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		id, err := exch.Buy(pair, rate, amount)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		b.WriteString(fmt.Sprintf("Successfully placed %s pair buy order:\n", pair.String()))
		b.WriteString(fmt.Sprintf(" - Order ID: %s\n", id))
		b.WriteString(fmt.Sprintf(" - Rate: %s\n", rate.String()))
		b.WriteString(fmt.Sprintf(" - Amount: %s\n", amount.String()))
	}

	if *exchSell != "" {
		amount, err := decimal.NewFromString(*exchSell)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		id, err := exch.Sell(pair, rate, amount)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		b.WriteString(fmt.Sprintf("Successfully placed %s pair sell order:\n", pair.String()))
		b.WriteString(fmt.Sprintf(" - Order ID: %s\n", id))
		b.WriteString(fmt.Sprintf(" - Rate: %s\n", rate.String()))
		b.WriteString(fmt.Sprintf(" - Amount: %s\n", amount.String()))
	}

	if *exchCancelOrder != "" {
		err := exch.CancelOrder(pair, *exchCancelOrder)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		b.WriteString(fmt.Sprintf("Successfully cancelled %s order\n", *exchCancelOrder))
	}

	if *exchOrder != "" {
		order, err := exch.GetOrder(pair, *exchOrder)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		b.WriteString("Order info:\n")
		b.WriteString(fmt.Sprintf(" - Timestamp: %s\n", order.Timestamp.Format(time.RFC1123)))
		b.WriteString(fmt.Sprintf(" - ID: %s\n", order.ID))
		b.WriteString(fmt.Sprintf(" - Is filled: %v\n", order.IsFilled))
		b.WriteString(fmt.Sprintf(" - Amount: %s\n", order.Amount.String()))
		b.WriteString(fmt.Sprintf(" - Rate: %s\n", order.Rate.String()))
		b.WriteString(fmt.Sprintf(" - Order side: %s\n", order.Side))
	}

	if *exchOpenOrders || *exchRetrieveAll {
		openOrders, err := exch.GetOpenOrders(pair)
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		if len(openOrders) == 0 {
			b.WriteString(fmt.Sprintf("%s pair has no open orders\n", pair.String()))
		} else if len(openOrders) > 5 {
			b.WriteString(fmt.Sprintf("%s pair open orders (first 5):\n", pair.String()))
		} else {
			b.WriteString(fmt.Sprintf("%s pair open orders:\n", pair.String()))
		}

		for i, order := range openOrders {
			if i == 5 {
				break
			}

			b.WriteString("---\n")
			b.WriteString(fmt.Sprintf(" - ID: %s\n", order.ID))
			b.WriteString(fmt.Sprintf(" - Is filled: %v\n", order.IsFilled))
			b.WriteString(fmt.Sprintf(" - Amount: %s\n", order.Amount.String()))
			b.WriteString(fmt.Sprintf(" - Rate: %s\n", order.Rate.String()))
			b.WriteString(fmt.Sprintf(" - Order side: %s\n", order.Side))
			b.WriteString("---\n")
		}
	}

	if *exchOrderHistory || *exchRetrieveAll {
		start := time.Now().Add(-time.Duration(*exchOrderHistoryDays) * 24 * time.Hour)
		orderHist, err := exch.GetOrderHistory(pair, start, time.Time{})
		if err != nil {
			ebCMD.Fatalf("%s", err)
		}

		if b.Len() > 0 {
			b.WriteString("----------\n")
		}

		if len(orderHist) == 0 {
			b.WriteString(fmt.Sprintf("%s pair has no orders in specified time interval\n", pair.String()))
		} else if len(orderHist) > 5 {
			b.WriteString(fmt.Sprintf("%s pair orders (latest 5):\n", pair.String()))
		} else {
			b.WriteString(fmt.Sprintf("%s pair orders:\n", pair.String()))
		}

		for i, order := range orderHist {
			if i == 5 {
				break
			}

			b.WriteString("---\n")
			b.WriteString(fmt.Sprintf(" - Timestamp: %s\n", order.Timestamp.Format(time.RFC1123)))
			b.WriteString(fmt.Sprintf(" - ID: %s\n", order.ID))
			b.WriteString(fmt.Sprintf(" - Is filled: %v\n", order.IsFilled))
			b.WriteString(fmt.Sprintf(" - Amount: %s\n", order.Amount.String()))
			b.WriteString(fmt.Sprintf(" - Rate: %s\n", order.Rate.String()))
			b.WriteString(fmt.Sprintf(" - Order side: %s\n", order.Side))
			b.WriteString("---\n")
		}
	}

	fmt.Println(b.String())
}
