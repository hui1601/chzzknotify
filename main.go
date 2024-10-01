package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	jsontime "github.com/liamylian/jsontime/v2/v2"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
	_ "time/tzdata"
)

var json = jsontime.ConfigWithCustomTimeFormat

type Notice struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	IsPinned    bool      `json:"isPinned"`
	IsEmergency bool      `json:"isEmergency"`
	RegDate     time.Time `json:"regDate" time_format:"2006-01-02T15:04:05" time_location:"Asia/Seoul"`
	ModDate     time.Time `json:"modDate" time_format:"2006-01-02T15:04:05" time_location:"Asia/Seoul"`
	ViewUrl     string    `json:"viewUrl"`
}

type Response struct {
	Item []Notice `json:"item"`
}

var registeredUsers = make(map[int64]bool)

func main() {

	// memory watchdog
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Printf("Memory usage: %d KiB", m.Alloc/1024)
		}
	}()
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	loadRegisteredUsers()

	go func() {
		for update := range updates {
			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				if update.Message.Command() == "register" {
					chatID := update.Message.Chat.ID
					registeredUsers[chatID] = true
					saveRegisteredUsers() // registeredUsers 파일 저장
					msg := tgbotapi.NewMessage(chatID, "등록했어요! 새로운 공지가 올라올 때마다 알려드릴게요.")
					_, err := bot.Send(msg)
					if err != nil {
						log.Println("Error sending message:", err)
					}
				}
				if update.Message.Command() == "unregister" {
					chatID := update.Message.Chat.ID
					delete(registeredUsers, chatID)
					saveRegisteredUsers()
					msg := tgbotapi.NewMessage(chatID, "해제했어요! 더 이상 알림을 받지 않아요.")
					_, err := bot.Send(msg)
					if err != nil {
						log.Println("Error sending message:", err)
					}
				}
				if update.Message.Command() == "start" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "안녕하세요! 치지직 공지 알림 봇(비공식)입니다.\n공식 홈페이지 공지사항을 알려드려요.\n\n/register: 알림 등록\n/unregister: 알림 해제")
					_, err := bot.Send(msg)
					if err != nil {
						log.Println("Error sending message:", err)
					}
				}
			}
		}
	}()

	var prevNotices []Notice
	resp, err := http.Get("https://notice.naver.com/api/v1/services/CHZZK/notices")
	if err != nil {
		log.Panic(err)
	}

	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		_ = resp.Body.Close()
		log.Panic(err)
	}
	prevNotices = response.Item
	_ = resp.Body.Close()
	for {
		func() {
			resp, err := http.Get("https://notice.naver.com/api/v1/services/CHZZK/notices")
			if err != nil {
				log.Println("Error fetching notices:", err)
				return
			}
			//goland:noinspection GoUnhandledErrorResult
			defer resp.Body.Close()

			var response Response
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				log.Println("Error decoding JSON:", err)
				return
			}

			for _, notice := range response.Item {
				found := false
				for _, prevNotice := range prevNotices {
					if notice.ID == prevNotice.ID {
						found = true
						if notice.ModDate != prevNotice.ModDate {
							msg := fmt.Sprintf("공지가 변경되었어요!\n제목: %s\n작성시간: %s\n변경시간: %s\n링크: %s",
								getNoticeTitleWithTags(notice), notice.RegDate.Format("2006-01-02 15:04:05"),
								notice.ModDate.Format("2006-01-02 15:04:05"), notice.ViewUrl)
							for chatID := range registeredUsers {
								sendMessage(bot, chatID, msg)
							}
						}
						break
					}
				}

				if !found {
					msg := fmt.Sprintf("새로운 공지가 있어요!\n제목: %s\n작성시간: %s\n링크: %s",
						getNoticeTitleWithTags(notice), notice.RegDate.Format("2006-01-02 15:04:05"), notice.ViewUrl)
					for chatID := range registeredUsers {
						sendMessage(bot, chatID, msg)
					}
				}
			}

			prevNotices = response.Item
		}()
		time.Sleep(1 * time.Minute)
	}
}

func getNoticeTitleWithTags(notice Notice) string {
	title := notice.Title
	if notice.IsEmergency {
		title = "[긴급] " + title
	}
	if notice.IsPinned {
		title = "[고정] " + title
	}
	return title
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Println("Error sending message:", err)
	}
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
