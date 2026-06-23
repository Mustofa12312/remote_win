package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL  string
	DeviceID string
	Secret   string
	HTTP     *http.Client
}

func New(baseURL, deviceID, secret string) *Client {
	return &Client{
		BaseURL:  baseURL,
		DeviceID: deviceID,
		Secret:   secret,
		HTTP:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) authHeader() string {
	return "Bearer " + c.DeviceID + ":" + c.Secret
}

// Register sends device registration request and returns device_id, secret
func Register(baseURL, name, osName, hostname, version string) (deviceID, secret string, err error) {
	payload := map[string]string{
		"name":          name,
		"os":            osName,
		"hostname":      hostname,
		"agent_version": version,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(baseURL+"/api/devices/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("registration request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("registration failed (%d): %s", resp.StatusCode, respBody)
	}

	var result struct {
		DeviceID string `json:"device_id"`
		Secret   string `json:"secret"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", fmt.Errorf("parse registration response: %w", err)
	}
	return result.DeviceID, result.Secret, nil
}

// Heartbeat sends metrics to server
func (c *Client) Heartbeat(localIP string, metrics interface{}) error {
	payload := map[string]interface{}{
		"local_ip": localIP,
		"metrics":  metrics,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", c.BaseURL+"/api/agent/heartbeat", bytes.NewReader(body))
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("heartbeat failed (%d): %s", resp.StatusCode, b)
	}
	return nil
}

// Command represents a command received from server
type Command struct {
	ID      uint   `json:"id"`
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Status  string `json:"status"`
}

// PollCommands fetches pending commands from server
func (c *Client) PollCommands() ([]Command, error) {
	req, _ := http.NewRequest("GET", c.BaseURL+"/api/agent/commands/poll", nil)
	req.Header.Set("Authorization", c.authHeader())

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Commands []Command `json:"commands"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Commands, nil
}

// ReportResult sends command execution result to server
func (c *Client) ReportResult(cmdID uint, status string, result interface{}) error {
	payload := map[string]interface{}{
		"status": status,
		"result": result,
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/api/agent/commands/%d/result", c.BaseURL, cmdID),
		bytes.NewReader(body))
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
