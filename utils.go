package main

import (
	"log"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, msg tgbotapi.MessageConfig) {
	msg.ChatID = chatID
	if _, err := bot.Send(msg); err != nil {
		log.Println("Error sending message:", err)
	}
}

func sendMessageToRegisteredUsers(bot *tgbotapi.BotAPI, text string) {
	for chatID := range registeredUsers {
		msg := tgbotapi.NewMessage(chatID, text)
		sendMessage(bot, chatID, msg)
	}
}

func escapeURLParam(param string) string {
	return url.QueryEscape(param)
}
