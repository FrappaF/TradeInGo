package utils

//This package contains all the useful boring func

import (
	"fmt"
	"math"
)

//Retrieve the absolute difference between two float32
func AbsDifference(x1, x2 float32) float32 { return float32(math.Abs(float64(x1 - x2))) }

//Retrieve the highest and the lowest between the two number x1, x2
//Useful to build candles from other candles
func GetHighLow(x1, x2 float32) (high, low float32) {
	if x1 > x2 {
		low = x2
		high = x1
	} else {
		low = x2
		high = x1
	}
	return
}

//Log with a title and a body
func PrintStatus(title, body string) {
	fmt.Printf("============================== *%v* ====================================\n", title)
	fmt.Printf("%v\n", body)
	fmt.Println("=====================================°============================================")
	fmt.Println("")
}
