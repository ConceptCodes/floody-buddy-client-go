package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
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

// MakeRequest sends a request to the specified address and handles retries in case of failure.
// It takes the address as a parameter and returns an error if the request fails after the maximum number of retries.
func MakeRequest(address string) error {
	for i := 0; i < config.AppConfig.MaxRetries; i++ {
		err := sendRequest(address)
		if err != nil {
			err = handleRequestFailure(err, i)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}

	return errors.New("failed to make request")
}

// sendRequest sends an HTTP GET request to the specified address.
// It logs the address and makes the request. If the request is successful,
// it handles the response using the HandleResponse function.
// It returns an error if there was a problem making the request.
func sendRequest(address string) error {
	log.Debug().Str("address", address).Msg("Making request")
	resp, err := http.Get(address)
	if err == nil {
		HandleResponse(resp)
	}
	return err
}

// handleRequestFailure handles the failure of a request and implements retry logic.
// It takes in the error that occurred, the attempt number, the retry interval, and the timeout duration.
// It returns the updated retry interval and an error if the request fails after reaching the max timeout.
func handleRequestFailure(err error, attempt int) error {
	log.Error().Err(err).Msgf("Failed to make request, attempt %d", attempt+1)
	timeout := time.Duration(config.AppConfig.Timeout) * time.Second

	delay, err := getDelay(attempt, timeout)

	if err != nil {
		return err
	}

	log.Debug().Dur("wait_time", delay).Msg("Waiting before retrying")
	time.Sleep(delay)

	return nil
}

// getDelay calculates the delay time for a retry attempt based on the attempt number and timeout duration.
// It applies exponential backoff with jitter to introduce randomness in the delay time.
// If the calculated delay time exceeds the timeout duration, it returns the timeout duration with an error.
func getDelay(attempt int, timeout time.Duration) (time.Duration, error) {
	backoff := time.Duration(1<<uint(attempt)) * time.Second
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	waitTime := backoff + jitter

	if waitTime > timeout {
		return timeout, errors.New("request timed out")
	}

	return waitTime, nil
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
