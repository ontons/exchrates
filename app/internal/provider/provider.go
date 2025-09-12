package provider

import "time"

type Rate struct {
	Currency string
	Value    float64
	Date     time.Time
}

type Provider interface {
	FetchRates() ([]Rate, error)
}
