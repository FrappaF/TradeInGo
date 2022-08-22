package main

//Made by Francesco Pippo (FrappaF)

import (
	"fmt"
	"os"
	"time"

	"github.com/frappaf/tradingBot/backtest"
	"github.com/frappaf/tradingBot/bot"
)

func main() {

	from := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local).Unix()

	args := os.Args[1:]
	if !(len(args) > 0) {
		fmt.Println("COMMAND NOT FOUND TRY live OR test")
		os.Exit(-1)
	}

	if args[0] == "live" {
		to := time.Now().Unix()
		btcBot := bot.Bot{}
		btcBot.Initialize(10000, from, to)
		btcBot.Run()
	} else if args[0] == "test" {
		to := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.Local).Unix()
		backtest.RunBacktest(from, to, "30")
	} else {
		fmt.Println("COMMAND NOT VALID TRY live OR test")
		os.Exit(-1)
	}

}
