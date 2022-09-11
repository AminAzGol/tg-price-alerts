package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/AminAzGol/tg-price-alerts/pricealerts"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const authorizedChatId = 1116041467

func main() {
	bot, err := tgbotapi.NewBotAPI("5655274864:AAEs4hyQ90FCpsvfRsCxEePZsiJqr4NdHog")
	if err != nil {
		log.Panic(err)
	}

	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	am := pricealerts.NewAlertManager()

	ch := make(chan string)
	go am.AlertCheckEngineStart(ch)

	for {
		select {
		case update := <-updates:
			handleTGUpdate(bot, am, &update)
		case engineMsg := <-ch:
			handleEngineMessage(bot, engineMsg)
		}
	}
}
func handleEngineMessage(bot *tgbotapi.BotAPI, text string) {
	msg := tgbotapi.NewMessage(authorizedChatId, text)
	bot.Send(msg)
}
func handleTGUpdate(bot *tgbotapi.BotAPI, am *pricealerts.AlertManager, update *tgbotapi.Update) {
	if update.Message != nil { // If we got a message
		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		var text string
		if update.Message.Chat.ID != authorizedChatId {
			text = fmt.Sprintf("Hi, you chat id is: %d", update.Message.Chat.ID)
		} else {
			re := regexp.MustCompile(`\s+`)
			commandArgs := re.Split(update.Message.Text, -1)

			if commandArgs[0] == "/price" {
				text = handlePriceCommand(commandArgs)
			} else if commandArgs[0] == "/set" {
				text = handleSetPriceAlert(am, commandArgs)
			} else if commandArgs[0] == "/alerts" {
				text = handleGetAlerts(am)
			} else if commandArgs[0] == "/remove" {
				text = handleRemoveCommand(am, commandArgs)
			} else {
				text = "Unknown Command"
			}
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	}
}
func handlePriceCommand(args []string) string {
	var text string
	if len(args) < 2 {
		text = "Not enough command arguments!"
	} else {

		am := pricealerts.NewAlertManager()
		ticker, err := pricealerts.FindTicker(am.Api, args[1])
		if err != nil {
			text = fmt.Sprintf("Error: %s", err)
		} else {
			text = fmt.Sprintf("%s price: %s", args[1], ticker.Price)
		}
	}
	return text
}

func handleSetPriceAlert(am *pricealerts.AlertManager, args []string) string {
	var text string
	if len(args) < 3 {
		text = "Not enough command arguments!"
	} else {
		price, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			text = fmt.Sprintf("Error: %s", err)
			return text
		}
		alert, err := am.SetAlert(args[1], price)
		if err != nil {
			text = fmt.Sprintf("Error: %s", err)
			return text
		}
		text = fmt.Sprintf("Alert set for %s at %.2f", alert.Ticker, alert.TargetPrice)
	}
	return text
}

func handleGetAlerts(am *pricealerts.AlertManager) string {
	if len(am.Alerts) == 0 {
		return "No alerts currently!"
	}
	text := "Alerts:\n"
	for i, a := range am.Alerts {
		text += fmt.Sprintf("%d. %s\n", i, a)
	}
	return text
}

func handleRemoveCommand(am *pricealerts.AlertManager, args []string) string {
	if len(args) < 2 {
		return "not enough command arguments!"
	}
	i, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return err.Error()
	}
	_, err = am.RemoveAlert(int(i))
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("alert %d removed", i)
}
