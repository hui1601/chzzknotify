package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"time"

	jsontime "github.com/liamylian/jsontime/v2/v2"
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

type NoticeResponse struct {
	Item []Notice `json:"item"`
}

func fetchNotices() ([]Notice, error) {
	resp, err := http.Get("https://notice.naver.com/api/v1/services/CHZZK/notices")
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	var response NoticeResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Item, nil
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

func checkNoticesUpdates(bot *tgbotapi.BotAPI) {
	var prevNotices []Notice
	notices, err := fetchNotices()
	if err != nil {
		log.Println("Error fetching initial notices:", err)
		return
	}
	prevNotices = notices

	for {
		notices, err := fetchNotices()
		if err != nil {
			log.Println("Error fetching notices:", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		for _, notice := range notices {
			found := false
			for _, prevNotice := range prevNotices {
				if notice.ID == prevNotice.ID {
					found = true
					if notice.ModDate != prevNotice.ModDate {
						noticeTitle := tgbotapi.EscapeText(tgbotapi.ModeMarkdown, getNoticeTitleWithTags(notice))
						noticeViewURL := tgbotapi.EscapeText(tgbotapi.ModeMarkdown, notice.ViewUrl)
						noticeRegistDate := tgbotapi.EscapeText(tgbotapi.ModeMarkdown, notice.RegDate.Format("2006-01-02 15:04:05"))
						noticeModDate := tgbotapi.EscapeText(tgbotapi.ModeMarkdown, notice.ModDate.Format("2006-01-02 15:04:05"))
						msg := fmt.Sprintf("공지가 변경되었어요!\n*%s*\n작성시간: %s\n변경시간: %s\n링크: %s",
							noticeTitle, noticeRegistDate, noticeModDate, noticeViewURL)
						sendMessageToRegisteredUsers(bot, msg)
					}
					break
				}
			}

			if !found {
				noticeTitle := tgbotapi.EscapeText(tgbotapi.ModeMarkdown, getNoticeTitleWithTags(notice))
				noticeViewURL := tgbotapi.EscapeText(tgbotapi.ModeMarkdown, notice.ViewUrl)
				noticeRegistDate := tgbotapi.EscapeText(tgbotapi.ModeMarkdown, notice.RegDate.Format("2006-01-02 15:04:05"))
				msg := fmt.Sprintf("새로운 공지가 있어요!\n*%s*\n작성시간: %s\n링크: %s",
					noticeTitle, noticeRegistDate, noticeViewURL)
				sendMessageToRegisteredUsers(bot, msg)
			}
		}

		prevNotices = notices
		time.Sleep(1 * time.Minute)
	}
}
