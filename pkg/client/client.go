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

type statsResponse struct {
	Hits      uint64 `json:"hits"`
	Misses    uint64 `json:"misses"`
	Evictions uint64 `json:"evictions"`
	HitRate   string `json:"hit_rate"`
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

func (c *Client) Stats() (statsResponse, error) {
	url := fmt.Sprintf("%s/stats", c.BaseURL)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return statsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return statsResponse{}, fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	var res statsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return statsResponse{}, err
	}

	return res, nil
}

func (c *Client) Compact() error {
	url := fmt.Sprintf("%s/compact", c.BaseURL)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
