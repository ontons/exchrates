package service

import (
	"exchrates/internal/provider"
	"exchrates/internal/store"
	"exchrates/pkg/logger"
)

type RateService struct {
	provider provider.Provider
	store    store.Store
	Logger   *logger.Logger
}

func NewRateService(provider provider.Provider, store store.Store, logger *logger.Logger) *RateService {
	return &RateService{
		provider: provider,
		store:    store,
		Logger:   logger,
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
