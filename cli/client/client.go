package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/0DayMonxrch/vaultify/cli/config"
)

type Client struct {
	httpClient *http.Client
	host       string
	token      string
}

func NewClient() (*Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		host:       cfg.Host,
		token:      cfg.Token,
	}, nil
}

type Project struct {
	ID        string `json:"ID"`
	Name      string `json:"Name"`
	Slug      string `json:"Slug"`
	CreatedAt string `json:"CreatedAt"`
}

type Secret struct {
	ID          string `json:"ID"`
	KeyName     string `json:"KeyName"`
	Environment string `json:"Environment"`
	CreatedAt   string `json:"CreatedAt"`
	UpdatedAt   string `json:"UpdatedAt"`
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	host := strings.TrimRight(c.host, "/")
	if !strings.HasSuffix(host, "/api/v1") {
		host = fmt.Sprintf("%s/api/v1", host)
	}
	url := fmt.Sprintf("%s%s", host, path)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vaultify-Token", c.token)
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) ListProjects() ([]Project, error) {
	req, err := c.newRequest(http.MethodGet, "/projects")
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var projects []Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *Client) ListSecrets(projectID string, env string) ([]Secret, error) {
	path := fmt.Sprintf("/projects/%s/secrets", projectID)
	if env != "" {
		path = fmt.Sprintf("%s?env=%s", path, env)
	}

	req, err := c.newRequest(http.MethodGet, path)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var secrets []Secret
	if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
		return nil, err
	}
	return secrets, nil
}

type DecryptedSecret struct {
	Value string `json:"value"`
}

func (c *Client) GetDecryptedSecret(projectID string, secretID string) (string, error) {
	path := fmt.Sprintf("/projects/%s/secrets/%s", projectID, secretID)
	req, err := c.newRequest(http.MethodGet, path)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var dec DecryptedSecret
	if err := json.NewDecoder(resp.Body).Decode(&dec); err != nil {
		return "", err
	}
	return dec.Value, nil
}
