package bot

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ably/ably-go/ably"
	"github.com/frappaf/tradingBot/data"
	"github.com/frappaf/tradingBot/utils"
)

const (
	minDifference float32 = 300.0
	long          int8    = 1
	short                 = -1
	neutral               = 0
)

// The bot is the core of the engine
// It contains a collection, the current balance,
// The current area that contains the price
// The current position that could be [long, short, neutral]
// The stopLoss, takeProfit are sensitive values, when the price reaches one of them the current position is closed
// The buyPrice stands for the price when it opend a position
// The units is the number of units long or short
type Bot struct {
	Collection                    data.Collection
	CurrentMoney                  float32
	currentPosition               Position
	currentArea, currentDayCandle data.Candle
}

// Initialize all the values
// It calls the FetchData and FindInterestingAreasAndKeyLevels methods of the collection
func (bot *Bot) Initialize(initialAmount float32, from, to int64) error {
	if initialAmount <= 0 {
		return fmt.Errorf("INITIAL AMOUNT MUST BE POSITIVE")
	}

	bot.CurrentMoney = initialAmount
	bot.currentArea = data.Candle{}
	bot.currentPosition = Position{}

	bot.currentDayCandle = data.Candle{
		Timestamp: 0,
	}

	err := bot.Collection.FetchData(from, to)
	if err != nil {
		return err
	}

	bot.Collection.FindInterestingAreasAndKeyLevels()

	return nil
}

// Find and returns, if exists, the area that contains the given candle
// If it not exists returns an empty candle and an error
func (bot *Bot) findArea(can data.Candle) (data.Candle, error) {

	index := binarySearchForCandles(bot.Collection.InterestAreas, can, 0, len(bot.Collection.InterestAreas))
	if index == -1 {
		return data.Candle{}, fmt.Errorf("AREA NOT FOUND")
	}

	return bot.Collection.InterestAreas[index], nil
}

// Recursive binary search for candles
func binarySearchForCandles(arr []data.Candle, can data.Candle, from, to int) int {
	index := (to-from)/2 + from

	if from > to {
		return -1
	}

	if arr[index].Contains(can.High) {
		return index
	}

	if arr[index].High < can.Close {

		if index == len(arr)-1 {
			return -1
		}
		return binarySearchForCandles(arr, can, index+1, to)
	} else {
		if index == 0 {
			return -1
		}
		return binarySearchForCandles(arr, can, from, index-1)
	}
}

// Close the current position
// It set all the data to 0
// Calculate the Profit/Loss and add to the current balance
func (bot *Bot) closePosition(value float32) {

	p_l := float32(bot.currentPosition.Position) * (value - bot.currentPosition.BuyPrice) * bot.currentPosition.Units
	bot.CurrentMoney += p_l

	utils.PrintStatus("POSITION CLOSED", "Closing position with P/L: "+fmt.Sprintf("%f", p_l))
	bot.currentPosition.BuyPrice = 0
	bot.currentPosition.TakeProfit = 0
	bot.currentPosition.StopLoss = 0
	bot.currentPosition.Units = 0
	bot.currentPosition.Position = neutral

}

// Print the global status of the bot
func (bot *Bot) Print() {

	body := "Current balance: " + fmt.Sprintf("%f", bot.CurrentMoney)
	switch bot.currentPosition.Position {
	case neutral:
		body += "\nCurrent position: NEUTRAL"
	case long:
		body += "\nCurrent position: LONG\tStopLoss: " + fmt.Sprintf("%f", bot.currentPosition.StopLoss) + "\tTakeProfit: " + fmt.Sprintf("%f", bot.currentPosition.TakeProfit)
	case short:
		body += "\nCurrent position: SHORT\tStopLoss: " + fmt.Sprintf("%f", bot.currentPosition.StopLoss) + "\tTakeProfit: " + fmt.Sprintf("%f", bot.currentPosition.TakeProfit)
	}

	body += "\n\nCurrentArea:\n"
	body += bot.currentArea.ToString()

	body += "\n\nCurrent daily candle:\n"
	body += bot.currentDayCandle.ToString()

	utils.PrintStatus("BOT STATUS", body)
}

// Given a new candle it decides whether to open a position or not
func (bot *Bot) Predict(candle data.Candle, present time.Time) {

	bot.updateCurrentDailyCandle(candle.Close, present)

	bot.Print()
	utils.PrintStatus("CURRENT PRICE", candle.ToString())

	//Check if the price has reached the stopLoss or the takeProfit
	if bot.currentPosition.Position == long && (bot.currentPosition.StopLoss >= candle.Close || bot.currentPosition.TakeProfit <= candle.Close) {
		bot.closePosition(candle.Close)
	}
	if bot.currentPosition.Position == short && (bot.currentPosition.StopLoss <= candle.Close || bot.currentPosition.TakeProfit >= candle.Close) {
		bot.closePosition(candle.Close)
	}

	//If the price in not in an interesting area yet search again
	if bot.currentArea.Close == 0.0 {
		closest, err := bot.findArea(candle)
		if err == nil {
			bot.currentArea = closest
			fmt.Println("Price inside interesting area")
			closest.Print()
		} else {
			fmt.Println("Area not yet discovered...")
		}

	} else if bot.currentPosition.Position == neutral { //If there is no opend position it can open one

		low, high := utils.GetHighLow(bot.currentArea.High, bot.currentArea.Low)

		//If the price is under the current area there is a possible short
		if candle.Close < low && (low-candle.Close) >= minDifference {
			bot.currentArea.Close = 0 //Need to find another area to condsider
			tp := bot.findNextInterestingLevel(candle.Close, short) + minDifference*0.2
			sl := candle.Close + minDifference*1.5
			if tp > 0 && math.Abs(float64(candle.Close)-float64(tp)) > math.Abs(float64(candle.Close)-float64(sl))-(float64(minDifference)*1.1) {
				fmt.Printf("Price under the area --> Opening short position\nStoploss: %v    takeProfit: %v    units: %v\n", sl, tp, bot.CurrentMoney/candle.Close)
				bot.currentPosition.Position = short
				bot.currentPosition.StopLoss = sl
				bot.currentPosition.TakeProfit = tp
				bot.currentPosition.BuyPrice = candle.Close
				bot.currentPosition.Units = bot.CurrentMoney / candle.Close
			}
		} else if candle.Close > high && (candle.Close-high) >= minDifference { //Else if the price is on top of the current area there is a possible long
			bot.currentArea.Close = 0 // Need to find another area to condsider
			tp := bot.findNextInterestingLevel(candle.Close, long) - minDifference*0.2
			sl := candle.Close - minDifference*1.5
			if tp > 0 && math.Abs(float64(candle.Close)-float64(tp)) > math.Abs(float64(candle.Close)-float64(sl))-(float64(minDifference)*1.1) {
				fmt.Printf("Price over the area --> Opening long position\nStoploss: %v    takeProfit: %v    units: %v\n", sl, tp, bot.CurrentMoney/candle.Close)

				bot.currentPosition.Position = long
				bot.currentPosition.StopLoss = sl
				bot.currentPosition.TakeProfit = tp
				bot.currentPosition.BuyPrice = candle.Close
				bot.currentPosition.Units = bot.CurrentMoney / candle.Close
			}
		}

	}

}

// Find the next interesting levels for the take profit
// If the position is long it search for the first level + delta  > value from the first to the last
// If the position is short it search for the first level < value + delta from the last to the first
func (bot *Bot) findNextInterestingLevel(value float32, position int8) float32 {

	delta := float32(50.0)

	switch position {
	case short:

		for i := len(bot.Collection.KeyLevels) - 1; i >= 0; i-- {
			if bot.Collection.KeyLevels[i]+delta < value {
				return float32(bot.Collection.KeyLevels[i])
			}
		}
		return 0
	case long:
		for i := 0; i < len(bot.Collection.KeyLevels); i++ {
			if bot.Collection.KeyLevels[i] > value+delta {
				return float32(bot.Collection.KeyLevels[i])
			}
		}
		return 0
	default:
		return 0

	}

}

// Update the current daily candle checking if the day has gone
// Or if there is a new High or a new Low
func (bot *Bot) updateCurrentDailyCandle(value float32, present time.Time) {
	currDay := time.Unix(bot.currentDayCandle.Timestamp, 0)

	if sub := present.Sub(currDay); sub.Abs().Hours() >= 24 {
		bot.currentDayCandle.Close = value

		//If it's not the first day candle
		if bot.currentDayCandle.Timestamp != 0 {
			utils.PrintStatus("NEW CANDLE APPENDED", bot.currentDayCandle.ToString())
			bot.Collection.History = append(bot.Collection.History, bot.currentDayCandle)
		}

		bot.Collection.FindInterestingAreasAndKeyLevels()

		bot.currentDayCandle = data.Candle{
			Open:      value,
			Timestamp: present.Unix(),
			High:      0,
			Low:       0xFFFFFF,
		}

	} else {

		if value > bot.currentDayCandle.High {
			bot.currentDayCandle.High = value
		} else if value < bot.currentDayCandle.Low {
			bot.currentDayCandle.Low = value
		}
	}
}

// Stream BTC price data using ably package and call the predict on that
func (bot *Bot) Run() {
	client, err := ably.NewRealtime(
		ably.WithKey("TsoT_A.ll-gaA:PPOPgVew_cMvzi_SrVd_QbuQvm_u_puG1IYMQVjR0S0"),
		ably.WithAutoConnect(false),
	)
	if err != nil {
		panic(err)
	}

	client.Connect()
	channel := client.Channels.Get("[product:ably-coindesk/bitcoin]bitcoin:usd")

	_, err = channel.SubscribeAll(context.Background(), func(msg *ably.Message) {

		price, ok := msg.Data.(string)
		if !ok {
			fmt.Println("Cannot convert the value in string")
			return
		}
		value, err := strconv.ParseFloat(price, 32)
		if err != nil {
			fmt.Println("Cannot convert the value in float32")
			return
		}

		value32 := float32(value)

		present := time.Now()

		candle := data.Candle{
			Open:      value32,
			Close:     value32,
			High:      value32,
			Low:       value32,
			Volume:    0,
			Timestamp: present.Unix(),
		}

		bot.Predict(candle, present)

	})

	if err != nil {
		err := fmt.Errorf("subscribing to channel: %w", err)
		fmt.Println(err)
	}

	//Used to never stop listening to price data
	time.Sleep(time.Hour * 0xFFFFF)
}
