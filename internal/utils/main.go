package utils

import (
	"bufio"
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

func loadUserAgentStrings() ([]string, error) {
	file, err := os.Open("user_agents.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var userAgents []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		userAgent := strings.TrimRight(scanner.Text(), ",\n")
		userAgents = append(userAgents, userAgent)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return userAgents, nil
}

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
		req, err := buildGetRequest(address)
		if err != nil {
			return err
		}
		err = sendRequest(*req)
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

// buildRequest builds an HTTP GET request with the specified address and sets the necessary headers.
func buildGetRequest(address string) (*http.Request, error) {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", genUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Referer", genReferer())

	return req, nil
}

// sendRequest sends an HTTP GET request to the specified address.
// It logs the address and makes the request. If the request is successful,
// it handles the response using the HandleResponse function.
// It returns an error if there was a problem making the request.
func sendRequest(req http.Request) error {
	log.Debug().Str("address", req.URL.Host).Str("user_agent", req.UserAgent()).Str("remote_address", req.RemoteAddr).Msg("Making request")
	resp, err := http.DefaultClient.Do(&req)

	if err == nil {
		err = handleResponse(resp)
		if err != nil {
			return err
		}
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
func handleResponse(resp *http.Response) error {
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	switch {
	case statusCode >= 200 && statusCode < 300:
		log.Debug().Int("status_code", statusCode).Msg("Request was successful")
		return nil
	case statusCode >= 300 && statusCode < 400:
		log.Warn().Int("status_code", statusCode).Msg("Request was redirected")
		return errors.New("request was redirected")
	case statusCode >= 400 && statusCode < 500:
		log.Error().Int("status_code", statusCode).Msg("Request failed due to client error")
		return errors.New("client error")
	case statusCode >= 500:
		log.Error().Int("status_code", statusCode).Msg("Request failed due to server error")
		return errors.New("server error")
	default:
		log.Error().Int("status_code", statusCode).Msg("Request failed")
		return errors.New("request failed")
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

// genUserAgent generates a random user agent string.
// It loads user agent strings from a source and returns a randomly selected one.
func genUserAgent() string {
	userAgentStrings, err := loadUserAgentStrings()

	if err != nil {
		log.Error().Err(err).Msg("Failed to load user agent strings")
		return ""
	}

	return userAgentStrings[rand.Intn(len(userAgentStrings))]
}

// genReferer generates a random referer URL from a list of predefined referers.
func genReferer() string {
	referers := []string{
		"https://www.google.com/search?q=",
		"https://check-host.net/",
		"https://www.facebook.com/",
		"https://www.youtube.com/",
		"https://www.fbi.com/",
		"https://www.bing.com/search?q=",
		"https://r.search.yahoo.com/",
		"https://www.cia.gov/index.html",
		"https://vk.com/profile.php?auto=",
		"https://www.usatoday.com/search/results?q=",
		"https://help.baidu.com/searchResult?keywords=",
		"https://steamcommunity.com/market/search?q=",
		"https://www.ted.com/search?q=",
		"https://play.google.com/store/search?q=",
	}

	return referers[rand.Intn(len(referers))]
}
