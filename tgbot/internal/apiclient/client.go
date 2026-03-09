package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) doJSON(ctx context.Context, method, path, token string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+"/api"+path, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("API error %d: %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

// Auth

func (c *Client) Register(ctx context.Context, email, displayName, password string) (*AuthResponse, error) {
	var resp AuthResponse
	err := c.doJSON(ctx, http.MethodPost, "/auth/register", "", map[string]string{
		"email":       email,
		"displayName": displayName,
		"password":    password,
	}, &resp)
	return &resp, err
}

func (c *Client) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	var resp AuthResponse
	err := c.doJSON(ctx, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	}, &resp)
	return &resp, err
}

// Meetings

func (c *Client) ListMyMeetings(ctx context.Context, token string) ([]Meeting, error) {
	var meetings []Meeting
	err := c.doJSON(ctx, http.MethodGet, "/meetings/", token, nil, &meetings)
	return meetings, err
}

func (c *Client) ListAllMeetings(ctx context.Context, token string) ([]Meeting, error) {
	var meetings []Meeting
	err := c.doJSON(ctx, http.MethodGet, "/meetings/all", token, nil, &meetings)
	return meetings, err
}

func (c *Client) GetMeeting(ctx context.Context, token, meetingID string) (*Meeting, error) {
	var meeting Meeting
	err := c.doJSON(ctx, http.MethodGet, "/meetings/"+meetingID+"/", token, nil, &meeting)
	return &meeting, err
}

func (c *Client) CreateMeeting(ctx context.Context, token string, input *CreateMeetingInput) (*Meeting, error) {
	var meeting Meeting
	err := c.doJSON(ctx, http.MethodPost, "/meetings/", token, input, &meeting)
	return &meeting, err
}

func (c *Client) DeleteMeeting(ctx context.Context, token, meetingID string) error {
	return c.doJSON(ctx, http.MethodDelete, "/meetings/"+meetingID+"/", token, nil, nil)
}

func (c *Client) ConfirmMeeting(ctx context.Context, token, meetingID string, slotID *string) (*Meeting, error) {
	body := map[string]interface{}{}
	if slotID != nil {
		body["timeSlotId"] = *slotID
	}
	var meeting Meeting
	err := c.doJSON(ctx, http.MethodPost, "/meetings/"+meetingID+"/confirm", token, body, &meeting)
	return &meeting, err
}

func (c *Client) Vote(ctx context.Context, token, meetingID string, slotIDs []string) error {
	return c.doJSON(ctx, http.MethodPost, "/meetings/"+meetingID+"/votes", token, map[string]interface{}{
		"timeSlotIds": slotIDs,
	}, nil)
}

func (c *Client) UpdateRSVP(ctx context.Context, token, meetingID, status string) error {
	return c.doJSON(ctx, http.MethodPut, "/meetings/"+meetingID+"/participants/rsvp", token, map[string]string{
		"status": status,
	}, nil)
}

func (c *Client) SetTags(ctx context.Context, token, meetingID string, tags []string) error {
	return c.doJSON(ctx, http.MethodPut, "/meetings/"+meetingID+"/tags", token, map[string]interface{}{
		"tags": tags,
	}, nil)
}

func (c *Client) GetAllTags(ctx context.Context, token string) ([]string, error) {
	var tags []string
	err := c.doJSON(ctx, http.MethodGet, "/meetings/tags/all", token, nil, &tags)
	return tags, err
}

func (c *Client) SearchUsers(ctx context.Context, token, query string) ([]User, error) {
	var users []User
	err := c.doJSON(ctx, http.MethodGet, "/users/search?q="+url.QueryEscape(query)+"&limit=10", token, nil, &users)
	return users, err
}
