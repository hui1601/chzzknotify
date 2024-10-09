package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"time"
)

type Partner struct {
	ChannelId        string `json:"channelId"`
	ChannelImageUrl  string `json:"channelImageUrl"`
	OriginalNickname string `json:"originalNickname"`
	ChannelName      string `json:"channelName"`
	VerifiedMark     bool   `json:"verifiedMark"`
}
type PartnerResponse struct {
	Code    int         `json:"code"`
	Message interface{} `json:"message"`
	Content struct {
		Size int `json:"size"`
		Page struct {
			Next struct {
				OriginalNickname string `json:"originalNickname"`
			}
		}
		Data []Partner `json:"data"`
	}
}

func fetchPartners() ([]Partner, error) {
	req, err := http.NewRequest("GET", "https://api.chzzk.naver.com/service/v1/streamer-partners", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer resp.Body.Close()

	var response PartnerResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	partners := response.Content.Data
	for response.Content.Size > 0 {
		req, err = http.NewRequest("GET", "https://api.chzzk.naver.com/service/v1/streamer-partners?originalNickname="+escapeURLParam(response.Content.Page.Next.OriginalNickname), nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		//goland:noinspection GoUnhandledErrorResult
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			return nil, err
		}
		partners = append(partners, response.Content.Data...)
	}

	return partners, nil
}

func checkPartnersUpdates(bot *tgbotapi.BotAPI) {
	var prevPartners []Partner
	partners, err := fetchPartners()
	if err != nil {
		log.Println("Error fetching partners:", err)
		return
	}
	prevPartners = partners

	for {
		partners, err = fetchPartners()
		if err != nil {
			log.Println("Error fetching partners:", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		for _, partner := range partners {
			found := false
			for _, prevPartner := range prevPartners {
				if partner.ChannelId == prevPartner.ChannelId {
					found = true
					break
				}
			}
			if !found {
				msg := fmt.Sprintf("%s님이 파트너가 되었어요! 축하해요!👏👏\nhttps://chzzk.naver.com/%s",
					partner.ChannelName, partner.ChannelId)
				sendMessageToRegisteredUsers(bot, msg)
			}
		}
		prevPartners = partners
		time.Sleep(1 * time.Minute)
	}
}
