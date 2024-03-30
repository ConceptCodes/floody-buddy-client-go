package models

// Note: whats needed in the clients shutdown message
type Message struct {
	IpAddress []string `json:"ip_address"` // This is the IP address of the client
	Urls      []string `json:"url"`        // This is the URL thats were unable to reach
}
