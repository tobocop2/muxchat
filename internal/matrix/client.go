package matrix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a simple Matrix client for admin operations
type Client struct {
	homeserverURL string
	accessToken   string
	httpClient    *http.Client
}

// NewClient creates a new Matrix client
func NewClient(homeserverURL string) *Client {
	return &Client{
		homeserverURL: homeserverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Login authenticates with the homeserver
func (c *Client) Login(username, password string) error {
	payload := map[string]interface{}{
		"type": "m.login.password",
		"identifier": map[string]string{
			"type": "m.id.user",
			"user": username,
		},
		"password": password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(
		c.homeserverURL+"/_matrix/client/v3/login",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: %s", string(respBody))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	c.accessToken = result.AccessToken
	return nil
}

// CreateDirectMessage creates a DM room with a user
func (c *Client) CreateDirectMessage(userID string) (string, error) {
	payload := map[string]interface{}{
		"preset":    "trusted_private_chat",
		"is_direct": true,
		"invite":    []string{userID},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.homeserverURL+"/_matrix/client/v3/createRoom", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create room failed: %s", string(respBody))
	}

	var result struct {
		RoomID string `json:"room_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.RoomID, nil
}

// SendMessage sends a text message to a room
func (c *Client) SendMessage(roomID, message string) error {
	txnID := fmt.Sprintf("%d", time.Now().UnixNano())

	payload := map[string]interface{}{
		"msgtype": "m.text",
		"body":    message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s",
		c.homeserverURL, roomID, txnID)

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send message failed: %s", string(respBody))
	}

	return nil
}

// GetJoinedRooms returns list of rooms the user has joined
func (c *Client) GetJoinedRooms() ([]string, error) {
	req, err := http.NewRequest("GET", c.homeserverURL+"/_matrix/client/v3/joined_rooms", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get joined rooms failed: %s", string(respBody))
	}

	var result struct {
		JoinedRooms []string `json:"joined_rooms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.JoinedRooms, nil
}

// GetRoomMembers returns the members of a room
func (c *Client) GetRoomMembers(roomID string) ([]string, error) {
	url := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/members", c.homeserverURL, roomID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil // Room might not be accessible
	}

	var result struct {
		Chunk []struct {
			StateKey string `json:"state_key"`
		} `json:"chunk"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	members := make([]string, 0, len(result.Chunk))
	for _, m := range result.Chunk {
		members = append(members, m.StateKey)
	}
	return members, nil
}

// FindDirectMessageRoom finds an existing DM room with a user
func (c *Client) FindDirectMessageRoom(userID string) (string, error) {
	rooms, err := c.GetJoinedRooms()
	if err != nil {
		return "", err
	}

	for _, roomID := range rooms {
		members, err := c.GetRoomMembers(roomID)
		if err != nil {
			continue
		}
		// Check if this is a 2-person room with the target user
		if len(members) == 2 {
			for _, m := range members {
				if m == userID {
					return roomID, nil
				}
			}
		}
	}
	return "", nil
}

// GetOrCreateDirectMessage finds existing DM or creates a new one
func (c *Client) GetOrCreateDirectMessage(userID string) (string, bool, error) {
	// Try to find existing room first
	roomID, err := c.FindDirectMessageRoom(userID)
	if err == nil && roomID != "" {
		return roomID, false, nil // existing room
	}

	// Create new room
	roomID, err = c.CreateDirectMessage(userID)
	if err != nil {
		return "", false, err
	}
	return roomID, true, nil // new room
}

// LeaveRoom leaves and forgets a room
func (c *Client) LeaveRoom(roomID string) error {
	// Leave the room
	leaveURL := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/leave", c.homeserverURL, roomID)
	req, err := http.NewRequest("POST", leaveURL, strings.NewReader("{}"))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("leave room failed: %s", string(body))
	}

	// Forget the room so it doesn't show in room list
	forgetURL := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/forget", c.homeserverURL, roomID)
	req, err = http.NewRequest("POST", forgetURL, strings.NewReader("{}"))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
