package spammer

import (
	"context"
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

// Flood is a function that continuously sends requests to URLs until the context is cancelled or all URLs are burned.
func Flood(ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("Context done")
			return
		default:
			if AllBurned() {
				TriggerShutdown()
				cancel()
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
}

// TriggerShutdown triggers the shutdown process.
// It retrieves the IP addresses, formats the message, and publishes it to Kafka.
// If any error occurs during the process, it will log the error and initiate a shutdown.
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

// AllBurned checks if all the URLs in the list have been burned.
// It iterates over the URLs and returns false if any URL is not burned.
// If all URLs are burned, it logs a message and returns true.
func AllBurned() bool {
	for _, url := range urls {
		if !IsBurned(url) {
			return false
		}
	}
	log.Info().Msg("All URLs have been burned")
	return true
}

// BurnUrl marks the given URL as burned.
// It adds the URL to the `burnedUrls` map with a value of `true`.
func BurnUrl(url string) {
	log.Debug().Str("url", url).Msg("Burning URL")
	burnedUrls[url] = true
}

// IsBurned checks if a URL is burned.
// It returns true if the URL is burned, false otherwise.
func IsBurned(url string) bool {
	return burnedUrls[url]
}
