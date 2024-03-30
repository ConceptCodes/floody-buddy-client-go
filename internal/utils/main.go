package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

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
// It logs the address and handles the response.
func MakeRequest(address string) error {
	log.Debug().Str("address", address).Msg("Making request")
	resp, err := http.Get(address)
	if err != nil {
		log.Error().Err(err).Msg("Failed to make request")
		return err
	} else {
		HandleResponse(resp)
	}
	return nil
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

func FormatMessage(msg models.Message) (string, error) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(jsonData), nil
}
