package data

import (
	"math"
	"sort"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
	"github.com/frappaf/tradingBot/api"
	"github.com/frappaf/tradingBot/utils"
)

const (
	//indexes for the fibonacci retracement
	TwentyThree int = iota
	ThirtyEight
	SixtyOne
	SeventyEight
	Fifty
	//
	BufferLength int     = 3      //Used in getResistancesAndSupport
	minRange     float32 = 50.0   //Minimum range for an area
	maxRange             = 2500.0 //Maxim range for an area
)

//Contains a slice of candles
//And useful data:
//Top, Bottom [Candle] represents the maximum top candle and the minimum bottom candle on the whole Data
//InterestAreas [[]Candle] represents the interesting areas in the history of the price.
//The Candles represents the areas and are build like this:
//						   	The High and Open are the highest side of the area
//						   	The Low and Close are the lowest side of the area
type Collection struct {
	History, InterestAreas []Candle
	Top, Bottom            Candle
	KeyLevels              []float32
}

//Get the response for daily resolution
//Then call FetchDailyData passing the response
func (collection *Collection) FetchData(from, to int64) error {

	resDaily, err := api.GetResponse("BINANCE:BTCUSDT", "D", from, to)
	if err != nil {
		return err
	}

	collection.FetchDailyData(resDaily)
	return nil
}

//Fetch the History of the collection
//It takes the responses [finnhub.CryptoCandles]
//Set the History properly and find the Top and Bottom
func (collection *Collection) FetchDailyData(res finnhub.CryptoCandles) {
	top := Candle{}
	bottom := Candle{}
	bottom.Low = 0xFFFFF

	for i := 0; i < len(res.GetC()); i++ {

		var o, c, h, l, v float32
		var t int64

		o = res.GetO()[i]
		c = res.GetC()[i]
		h = res.GetH()[i]
		l = res.GetL()[i]
		v = res.GetV()[i]
		t = res.GetT()[i]

		candle := Candle{Open: o, Close: c, High: h, Low: l, Volume: v, Timestamp: t}
		collection.History = append(collection.History, candle)

		if top.High < h {
			top = candle
		}
		if bottom.Low > l {
			bottom = candle
		}

	}

	collection.Bottom = bottom
	collection.Top = top
}

//Fibonacci retracement using levels 23.6%, 38.2%, 61.8%, and 78.6% and adding 50% level
func (collection *Collection) getFibonacciRetracement() []float32 {
	fibRetracement := make([]float32, 5)

	distance := float32(collection.Bottom.High - collection.Top.Low)

	fibRetracement[TwentyThree] = collection.Top.Low + (distance * 0.236)
	fibRetracement[ThirtyEight] = collection.Top.Low + (distance * 0.382)
	fibRetracement[SixtyOne] = collection.Top.Low + (distance * 0.618)
	fibRetracement[SeventyEight] = collection.Top.Low + (distance * 0.786)
	fibRetracement[Fifty] = collection.Top.Low + (distance * 0.500)

	return fibRetracement
}

func (collection *Collection) FindInterestingAreasAndKeyLevels() {

	resSup := collection.getResistancesAndSupport()

	for i := 1; i < len(resSup)-1; {

		candle := resSup[i]

		body := utils.AbsDifference(candle.High, candle.Low)
		meanBody := float32(math.Abs(float64((candle.High + candle.Low) / 2)))

		//The Candle itself is an interesting area
		if body >= minRange {
			collection.KeyLevels = append(collection.KeyLevels, meanBody)
			collection.InterestAreas = append(collection.InterestAreas,
				Candle{
					Open:      candle.Open,
					Close:     candle.Close,
					High:      candle.High,
					Low:       candle.Low,
					Timestamp: candle.Timestamp,
				})

			i += 2
		} else {

			//Calculate the difference between the candle and its neighbor
			diffPrevCandle := math.Abs(float64(meanBody - resSup[i-1].High))
			diffNextCandle := math.Abs(float64(meanBody - resSup[i+1].Low))

			//If the candle is closer to the prev and the range is considerable big enough
			absDiff := utils.AbsDifference(meanBody, resSup[i-1].High)
			if diffPrevCandle < diffNextCandle && absDiff > minRange && absDiff < maxRange {
				high, low := utils.GetHighLow(candle.Low, resSup[i-1].High)

				collection.KeyLevels = append(collection.KeyLevels, meanBody)

				collection.InterestAreas = append(collection.InterestAreas,
					Candle{
						High:      high,
						Low:       low,
						Open:      high,
						Close:     low,
						Timestamp: candle.Timestamp,
					})

				i += 2 //Skip candles two by two
			} else {
				//Check if the range between the candle and the next is big enough
				absDiff := utils.AbsDifference(meanBody, resSup[i+1].Low)
				if absDiff > minRange && absDiff < maxRange {

					high, low := utils.GetHighLow(candle.High, resSup[i+1].Low)

					collection.KeyLevels = append(collection.KeyLevels, meanBody)

					collection.InterestAreas = append(collection.InterestAreas,
						Candle{
							High:      high,
							Low:       low,
							Open:      high,
							Close:     low,
							Timestamp: candle.Timestamp,
						})

					i += 3 //+2 and choosen the next candle -> +1
				} else {
					//The candle is not used for any areas so  skip to the next
					i += 1
				}
			}
		}

	}

	collection.KeyLevels = append(collection.KeyLevels, collection.getFibonacciRetracement()...)

	//Sorting the arrays to better perform searching engine
	sort.Slice(collection.KeyLevels, func(i, j int) bool { return collection.KeyLevels[i] < collection.KeyLevels[j] })
	sort.Slice(collection.InterestAreas, func(i, j int) bool {
		return collection.InterestAreas[i].High < collection.InterestAreas[j].High
	})

}

//Find resistance and supports
//If it finds [BurreLength] candles that shares a price area in their TOP shadows -> resistance
//If it finds [BurreLength] candles that shares a price area in their BOTTOM shadows -> support
func (collection *Collection) getResistancesAndSupport() []Candle {

	indexs := make([]int, 2)
	indexs[0] = BufferLength
	indexs[1] = 2

	var resSup []Candle
	var buffer [BufferLength]Candle

	for index, candle := range collection.History {
		if index < BufferLength {
			buffer[index] = candle
		} else {

			minTopShadow := minTopShadow(buffer)
			maxBody := maxBody(buffer)

			minBottomShadow := minBottomShadow(buffer)
			minBody := minBody(buffer)

			if minTopShadow > maxBody {
				resSup = append(resSup, Candle{Open: minTopShadow, Close: maxBody, High: minTopShadow, Low: maxBody, Volume: 0, Timestamp: candle.Timestamp})
			}

			if minBottomShadow < minBody {
				resSup = append(resSup, Candle{Open: minBody, Close: minBottomShadow, High: minBody, Low: minBottomShadow, Volume: 0, Timestamp: candle.Timestamp})
			}

			//Adding the new candle in the buffer and eliminate the first (Simulating a FILO)
			for i := 1; i <= BufferLength; i++ {
				if i == BufferLength {
					buffer[i-1] = candle
				} else {
					buffer[i-1] = buffer[i]
				}
			}

		}
	}

	return resSup
}

//Find the minimum top shadow of a given buffer of candles
func minTopShadow(buffer [BufferLength]Candle) float32 {

	var min float32
	min = 0xFFFFFF

	for _, candle := range buffer {
		if candle.High < min {
			min = candle.High
		}
	}
	return min
}

//Find the minimum bottom shadow of a given buffer of candles
func minBottomShadow(buffer [BufferLength]Candle) float32 {

	var min float32
	min = 0xFFFFFF

	for _, candle := range buffer {
		if candle.Low < min {
			min = candle.Low
		}
	}
	return min
}

//Find the maximum high body of a given buffer of candles
func maxBody(buffer [BufferLength]Candle) (res float32) {
	var max float32

	for _, candle := range buffer {
		if candle.Close > candle.Open {
			if candle.Close > max {
				max = candle.Close
			}
		} else {
			if candle.Open > max {
				max = candle.Open
			}
		}
	}

	return max
}

//Find the minimum high body of a given buffer of candles
func minBody(buffer [BufferLength]Candle) (res float32) {
	var min float32
	min = 0xFFFFFF
	for _, candle := range buffer {
		if candle.Close > candle.Open {
			if candle.Close < min {
				min = candle.Close
			}
		} else {
			if candle.Open < min {
				min = candle.Open
			}
		}
	}

	return min
}
