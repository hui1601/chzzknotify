package main

import (
	"io"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var registeredUsers = make(map[int64]bool)

func handleCommands(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "register":
				registerUser(bot, update.Message.Chat.ID)
			case "unregister":
				unregisterUser(bot, update.Message.Chat.ID)
			case "start":
				sendStartMessage(bot, update.Message.Chat.ID)
			}
		}
	}
}

func registerUser(bot *tgbotapi.BotAPI, chatID int64) {
	registeredUsers[chatID] = true
	saveRegisteredUsers()
	msg := tgbotapi.NewMessage(chatID, "등록했어요! 새로운 공지가 올라올 때마다 알려드릴게요.")
	sendMessage(bot, chatID, msg)
}

func unregisterUser(bot *tgbotapi.BotAPI, chatID int64) {
	delete(registeredUsers, chatID)
	saveRegisteredUsers()
	msg := tgbotapi.NewMessage(chatID, "해제했어요! 더 이상 알림을 받지 않아요.")
	sendMessage(bot, chatID, msg)
}

func sendStartMessage(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "안녕하세요! 치지직 공지 알림 봇(비공식)입니다.\n공식 홈페이지 공지사항을 알려드려요.\n\n/register: 알림 등록\n/unregister: 알림 해제")
	sendMessage(bot, chatID, msg)
}

func loadRegisteredUsers() {
	file, err := os.Open("registeredUsers.json")
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Panic(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&registeredUsers)
	if err != nil {
		if err == io.EOF {
			return
		}
		log.Panic(err)
	}
}

func saveRegisteredUsers() {
	file, err := os.Create("registeredUsers.json")
	if err != nil {
		log.Panic(err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(registeredUsers)
	if err != nil {
		log.Panic(err)
	}
}
