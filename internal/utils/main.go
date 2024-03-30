package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"floody-buddy/config"
	"floody-buddy/internal/models"
	"floody-buddy/pkg/logger"
)

var log = logger.New()

// FormatUrls formats the URLs by adding "http://" prefix to each domain in the AppConfig.Domains string.
// It splits the AppConfig.Domains string by comma and returns a slice of formatted URLs.
func FormatUrls() []string {
	urls := strings.Split(config.AppConfig.Domains, ",")
	var formattedUrls []string
	for _, url := range urls {
		formattedUrls = append(formattedUrls, fmt.Sprintf("http://%s", url))
	}
	return formattedUrls
}


// MakeRequest sends an HTTP GET request to the specified address.
// It retries the request up to 10 times with increasing intervals if it fails.
// If the request is successful, it calls the HandleResponse function.
// If the request fails after 10 attempts, it returns an error.
func MakeRequest(address string) error {
	log := logger.New()
	timeout := time.Duration(config.AppConfig.Timeout) * time.Second
	retryInterval := timeout / 10

	for i := 0; i < 10; i++ {
		log.Debug().Str("address", address).Msg("Making request")
		resp, err := http.Get(address)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to make request, attempt %d", i+1)
			time.Sleep(time.Duration(retryInterval) * time.Second)
			retryInterval += retryInterval / 10
			if retryInterval > timeout {
				return fmt.Errorf("failed to make request after reaching max timeout: %w", err)
			}
		} else {
			HandleResponse(resp)
			return nil
		}
	}

	return fmt.Errorf("failed to make request after 10 attempts")
}

// HandleResponse handles the HTTP response and logs the appropriate message based on the status code.
func HandleResponse(resp *http.Response) {
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	switch {
	case statusCode >= 200 && statusCode < 300:
		log.Debug().Int("status_code", statusCode).Msg("Request was successful")
	case statusCode >= 300 && statusCode < 400:
		log.Warn().Int("status_code", statusCode).Msg("Request was redirected")
	case statusCode >= 400 && statusCode < 500:
		log.Error().Int("status_code", statusCode).Msg("Request failed due to client error")
	case statusCode >= 500:
		log.Error().Int("status_code", statusCode).Msg("Request failed due to server error")
	default:
		log.Error().Int("status_code", statusCode).Msg("Request failed")
	}
}

// Shutdown shuts down the application gracefully.
func Shutdown() {
	log.Info().Msg("Shutting down application")
	os.Exit(0)
}

// GetIPAddresses returns a list of IPv4 addresses associated with the system's network interfaces.
func GetIPAddresses() ([]string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

// FormatMessage formats the given message as a JSON string.
// It takes a message of type models.Message as input and returns the formatted JSON string and an error (if any).
func FormatMessage(msg models.Message) (string, error) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(jsonData), nil
}
