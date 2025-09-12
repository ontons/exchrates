package store

import "exchrates/internal/provider"

type Store interface {
	SaveRates([]provider.Rate) error
	GetLatest() ([]provider.Rate, error)
	GetHistory(string) ([]provider.Rate, error)
}
