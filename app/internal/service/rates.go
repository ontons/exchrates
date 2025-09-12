package service

import (
	"exchrates/internal/provider"
	"exchrates/internal/store"
)

type RateService struct {
	provider provider.Provider
	store    store.Store
}

func NewRateService(provider provider.Provider, store store.Store) *RateService {
	return &RateService{
		provider: provider,
		store:    store,
	}
}

func (s *RateService) FetchAndSave() error {
	rates, err := s.provider.FetchRates()
	if err != nil {
		return err
	}
	return s.store.SaveRates(rates)
}

func (s *RateService) GetLatest() ([]provider.Rate, error) {
	return s.store.GetLatest()
}

func (s *RateService) GetHistory(currecy string) ([]provider.Rate, error) {
	return s.store.GetHistory(currecy)
}
