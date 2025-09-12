package provider

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type RssXml struct {
	Channel struct {
		Items []struct {
			Description string `xml:"description"`
		} `xml:"item"`
	} `xml:"channel"`
}

func (r *RssXml) ParseCurrencyRates() (map[string]float64, error) {
	if r.Channel.Items == nil || len(r.Channel.Items) == 0 {
		return nil, fmt.Errorf("No items found in RSS feed")
	}

	parts := strings.Fields(r.Channel.Items[len(r.Channel.Items)-1].Description)
	rates := make(map[string]float64)

	if len(parts)%2 != 0 {
		return nil, fmt.Errorf("Invalid input string")
	}

	for i := 0; i < len(parts); i += 2 {
		currency := parts[i]
		valueStr := parts[i+1]

		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("Invalid value for %s: %v", currency, err)
		}

		rates[currency] = value
	}

	return rates, nil
}

type RSSProvider struct {
	url string
}

func NewRSSProvider(url string) *RSSProvider {
	return &RSSProvider{url: url}
}

func (r *RSSProvider) GetCurrencyRates() (map[string]float64, error) {
	resp, err := http.Get(r.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)
	var rssXml RssXml
	if err := decoder.Decode(&rssXml); err != nil && err != io.EOF {
		return nil, err
	}

	return rssXml.ParseCurrencyRates()
}

func (r *RSSProvider) FetchRate(t time.Time, currency string) (Rate, error) {
	rates, err := r.GetCurrencyRates()
	if err != nil {
		return Rate{}, err
	}

	if val, ok := rates[currency]; !ok {
		return Rate{}, fmt.Errorf("Currency %s not found in RSS feed", currency)
	} else {
		return Rate{
			Currency: currency,
			Value:    val,
			Date:     t,
		}, nil
	}
}

func (r *RSSProvider) FetchRates() ([]Rate, error) {
	rates, err := r.GetCurrencyRates()
	if err != nil {
		return nil, err
	}

	ch := make(chan Rate, len(rates))
	t := time.Now()

	g := new(errgroup.Group)
	for currency, _ := range rates {
		c := currency
		g.Go(func() error {
			rate, err := r.FetchRate(t, c)
			if err != nil {
				return err
			}
			ch <- rate
			return nil
		})
	}

	go func() {
		g.Wait()
		close(ch)
	}()

	result := make([]Rate, 0, len(rates))
	for rate := range ch {
		result = append(result, rate)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}
