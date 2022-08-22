package backtest

import (
	"time"

	"github.com/frappaf/tradingBot/api"
	"github.com/frappaf/tradingBot/bot"
	"github.com/frappaf/tradingBot/data"
)

func RunBacktest(from, to int64, resolution string) {
	btcBot := bot.Bot{}
	err := btcBot.Initialize(10000.0, from, to)
	if err != nil {
		panic(err)
	}

	res, err := api.GetResponse("BINANCE:BTCUSDT", resolution, to, time.Now().Unix())

	for i := 0; i < len(res.GetC()); i++ {
		o := res.GetO()[i]
		c := res.GetC()[i]
		h := res.GetH()[i]
		l := res.GetL()[i]
		v := res.GetV()[i]
		t := res.GetT()[i]

		candle := data.Candle{Open: o, Close: c, High: h, Low: l, Volume: v, Timestamp: t}
		btcBot.Predict(candle)

	}
}
