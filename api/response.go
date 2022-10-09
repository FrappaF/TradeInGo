package api

import (
	"context"
	"os"

	finnhub "github.com/Finnhub-Stock-API/finnhub-go/v2"
)

// Initialize a new finnhub configuration and retrieve a finnhub.Cryptocandles for the given symbol, resolution, from and to
func GetResponse(symbol, resolution string, from, to int64) (finnhub.CryptoCandles, error) {
	cfg := finnhub.NewConfiguration()
	cfg.AddDefaultHeader("X-Finnhub-Token", os.Getenv("TOKEN"))
	finnhubClient := finnhub.NewAPIClient(cfg).DefaultApi

	res, _, err := finnhubClient.CryptoCandles(context.Background()).Symbol(symbol).Resolution(resolution).From(from).To(to).Execute()
	if err != nil {
		return finnhub.CryptoCandles{}, err
	}
	return res, nil
}
