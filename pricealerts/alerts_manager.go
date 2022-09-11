package pricealerts

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Kucoin/kucoin-go-sdk"
)

type PriceRelation int32

const (
	Above PriceRelation = iota
	Below
)

type Alert struct {
	Ticker            string
	TargetPrice       float64
	AlertOn           PriceRelation
	LastPriceRelation PriceRelation
}

func (a Alert) String() string {
	if a.AlertOn == Above {
		return fmt.Sprintf("%s above %.3f", a.Ticker, a.TargetPrice)
	} else {
		return fmt.Sprintf("%s below %.3f", a.Ticker, a.TargetPrice)
	}
}

type AlertManager struct {
	Alerts []Alert
	Api    *kucoin.ApiService
}

func (am *AlertManager) SetAlert(ticker string, price float64) (*Alert, error) {
	t, err := FindTicker(am.Api, ticker)
	if err != nil {
		return nil, err
	}
	currentPrice, err := strconv.ParseFloat(t.Price, 32)
	if err != nil {
		return nil, err
	}
	var alert Alert
	if currentPrice <= price {
		alert = Alert{
			Ticker:            ticker,
			TargetPrice:       price,
			LastPriceRelation: Below,
			AlertOn:           Above,
		}
		log.Printf("Alert set for %s when price goes above %.2f", ticker, price)
	} else if currentPrice > price {
		alert = Alert{
			Ticker:            ticker,
			TargetPrice:       price,
			LastPriceRelation: Above,
			AlertOn:           Below,
		}
		log.Printf("Alert set for %s when price goes below %.2f", ticker, price)
	}
	am.Alerts = append(am.Alerts, alert)
	return &alert, nil
}

func FindTicker(s *kucoin.ApiService, ticker string) (*kucoin.TickerLevel1Model, error) {
	rsp, err := s.TickerLevel1(ticker)
	if err != nil {
		return nil, err
	}

	t := kucoin.TickerLevel1Model{}
	if err = rsp.ReadData(&t); err != nil {
		return nil, err
	}
	if t.Price == "" {
		return nil, errors.New("Ticker not found!")
	}

	return &t, nil
}

func CheckAlertFired(s *kucoin.ApiService, alert *Alert) (bool, error) {
	t, err := FindTicker(s, alert.Ticker)
	if err != nil {
		return false, err
	}
	currentPrice, err := strconv.ParseFloat(t.Price, 32)
	if err != nil {
		return false, err
	}
	if alert.AlertOn == Above && currentPrice > alert.TargetPrice {
		return true, nil
	} else if alert.AlertOn == Below && currentPrice < alert.TargetPrice {
		return true, nil
	}

	return false, nil
}

func NewAlertManager() *AlertManager {
	s := kucoin.NewApiService(
		// kucoin.ApiBaseURIOption("https://api.kucoin.com"),
		kucoin.ApiKeyOption("key"),
		kucoin.ApiSecretOption("secret"),
		kucoin.ApiPassPhraseOption("passphrase"),
		kucoin.ApiKeyVersionOption(kucoin.ApiKeyVersionV2))

	am := AlertManager{Api: s}
	return &am
}
func NotifyAlertFired(alert *Alert) string {
	var text string
	if alert.AlertOn == Above {
		text = fmt.Sprintf("Alert fired: %s went above %f", alert.Ticker, alert.TargetPrice)
	} else {
		text = fmt.Sprintf("Alert fired: %s went below %f", alert.Ticker, alert.TargetPrice)
	}
	log.Print(text)
	return text
}

func (am *AlertManager) RemoveAlert(i int) ([]Alert, error) {
	if len(am.Alerts) == 0 {
		return nil, errors.New("no alerts to remove")
	}
	if i < 0 || i >= len(am.Alerts) {
		return nil, errors.New("index out of rage")
	}
	am.Alerts = append(am.Alerts[:i], am.Alerts[i+1:]...)
	return am.Alerts, nil
}

func (am *AlertManager) AlertCheckEngineStart(ch chan string) error {
	log.Print("Starting Alerts Check Engine!")
	for {
		time.Sleep(1 * time.Second)
		for i, alert := range am.Alerts {
			fired, err := CheckAlertFired(am.Api, &alert)
			if err != nil {
				return err
			}
			if fired {
				text := NotifyAlertFired(&alert)
				ch <- text
				am.RemoveAlert(i)
			}

			// ch <- fmt.Sprintf("Alert %s on %f checked, fired: %v", alert.Ticker, alert.TargetPrice, fired)

		}
	}
}
