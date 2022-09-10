package pricealerts

import (
	"fmt"
	"log"
	"testing"
)

func TestFindTicker(t *testing.T) {
	tickerName := "BTC-USDT"
	am := NewAlertManager()
	ticker, err := FindTicker(am.api, tickerName)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Print(ticker)
}

func TestSetAlert(t *testing.T) {
	am := NewAlertManager()
	tickerName := "BTC-USDT"
	price := 25000.231
	alert, err := am.SetAlert(tickerName, price)
	if err != nil {
		t.Fatal(err)
	}
	if len(am.Alerts) != 1 {
		t.Fatalf("AlertList length is incorrect. want: %d got: %d", 1, len(am.Alerts))
	}
	fmt.Print(alert)
}

func TestCheckAlertFired(t *testing.T) {
	am := NewAlertManager()
	tickerName := "BTC-USDT"
	price := 25000.231
	alert, err := am.SetAlert(tickerName, price)
	if err != nil {
		t.Fatal(err)
	}

	fired, err := CheckAlertFired(am.api, alert)
	if err != nil {
		t.Fatal(err)
	}

	if fired {
		log.Print("Alert fired")
	} else {
		log.Print("Alert not fired yet!")
	}
}

func TestRemoveAlert(t *testing.T) {
	am := NewAlertManager()
	tickerName := "BTC-USDT"
	price := 25000.231
	_, err := am.SetAlert(tickerName, price)
	if err != nil {
		t.Fatal(err)
	}

	price = 123000
	_, err = am.SetAlert(tickerName, price)
	if err != nil {
		t.Fatal(err)
	}

	if len(am.Alerts) != 2 {
		t.Fatalf("AlertList length is incorrect. want: %d got: %d", 2, len(am.Alerts))
	}

	am.removeAlert(0)
	if len(am.Alerts) != 1 {
		t.Fatalf("AlertList length is incorrect. want: %d got: %d", 1, len(am.Alerts))
	}
	if am.Alerts[0].TargetPrice != price {
		t.Fatalf("Wrong item removed. want: %f got: %f", price, am.Alerts[0].TargetPrice)
	}
}

func TestNotifyAlertFired(t *testing.T) {
	am := NewAlertManager()
	tickerName := "BTC-USDT"
	price := 25000.231
	alert, err := am.SetAlert(tickerName, price)
	if err != nil {
		t.Fatal(err)
	}
	NotifyAlertFired(alert)
}

func TestAlertEngine(t *testing.T) {
	am := NewAlertManager()
	tickerName := "BTC-USDT"
	price := 25000.231
	_, err := am.SetAlert(tickerName, price)
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan string)
	go am.AlertCheckEngineStart(ch)

	for i := 0; i < 3; i++ {
		val := <-ch
		log.Print(val)
	}

}
