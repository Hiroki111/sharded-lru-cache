package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new instance of the cache client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type setRequest struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	TTL   int         `json:"ttl"`
}

type getResponse struct {
	Value string `json:"value"`
}

func (c *Client) Set(key string, value interface{}, ttl time.Duration) error {
	url := fmt.Sprintf("%s/set", c.BaseURL)

	payload := setRequest{
		Key:   key,
		Value: value,
		TTL:   int(ttl.Seconds()),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to set key, status: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Get(key string) (string, error) {
	url := fmt.Sprintf("%s/get?key=%s", c.BaseURL, key)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("key not found")
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get key, status: %d", resp.StatusCode)
	}

	var res getResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Value, nil
}
