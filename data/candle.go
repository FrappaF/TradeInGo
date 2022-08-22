package data

import (
	"fmt"
	"strconv"
)

//Represents the OHLCV candle data including the timestamp
type Candle struct {
	Open, Close, High, Low, Volume float32
	Timestamp                      int64
}

//Print the data of the candle
func (candle *Candle) Print() {
	fmt.Printf(candle.ToString())
}

func (candle *Candle) ToString() string {
	return "Open: " + fmt.Sprintf("%f", candle.Open) + "\tClose: " + fmt.Sprintf("%f", candle.Close) + "\nHigh: " + fmt.Sprintf("%f", candle.High) + "\tLow: " + fmt.Sprintf("%f", candle.Low) + "\nVolume: " + fmt.Sprintf("%f", candle.Volume) + "\tTimestamp: " + strconv.FormatUint(uint64(candle.Timestamp), 10) + "\n"
}

//Check if a value is between the low and the high of a candle
func (candle *Candle) Contains(value float32) bool {
	if candle.Open > candle.Close {
		if candle.Close < value && candle.Open > value {
			return true
		}
		return false
	} else {
		if candle.Open < value && candle.Close > value {
			return true
		}
		return false
	}
}
