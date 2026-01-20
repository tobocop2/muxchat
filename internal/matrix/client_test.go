package matrix

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8008")
	if client.homeserverURL != "http://localhost:8008" {
		t.Errorf("expected homeserverURL to be http://localhost:8008, got %s", client.homeserverURL)
	}
	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestLogin_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_matrix/client/v3/login" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["type"] != "m.login.password" {
			t.Errorf("expected type m.login.password, got %v", payload["type"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": "test_token_123",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Login("testuser", "testpass")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if client.accessToken != "test_token_123" {
		t.Errorf("expected access_token to be test_token_123, got %s", client.accessToken)
	}
}

func TestLogin_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errcode": "M_FORBIDDEN", "error": "Invalid password"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Login("testuser", "wrongpass")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid password") {
		t.Errorf("expected error to contain 'Invalid password', got %v", err)
	}
}

func TestCreateDirectMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_matrix/client/v3/createRoom" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test_token" {
			t.Errorf("expected Bearer test_token, got %s", r.Header.Get("Authorization"))
		}

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["preset"] != "trusted_private_chat" {
			t.Errorf("expected preset trusted_private_chat, got %v", payload["preset"])
		}
		if payload["is_direct"] != true {
			t.Errorf("expected is_direct true, got %v", payload["is_direct"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"room_id": "!newroom:localhost",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	roomID, err := client.CreateDirectMessage("@user:localhost")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if roomID != "!newroom:localhost" {
		t.Errorf("expected room_id !newroom:localhost, got %s", roomID)
	}
}

func TestCreateDirectMessage_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errcode": "M_UNKNOWN", "error": "Cannot invite user"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	_, err := client.CreateDirectMessage("@user:localhost")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Cannot invite user") {
		t.Errorf("expected error to contain 'Cannot invite user', got %v", err)
	}
}

func TestSendMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/_matrix/client/v3/rooms/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["msgtype"] != "m.text" {
			t.Errorf("expected msgtype m.text, got %v", payload["msgtype"])
		}
		if payload["body"] != "Hello, World!" {
			t.Errorf("expected body 'Hello, World!', got %v", payload["body"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"event_id": "$event123",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	err := client.SendMessage("!room:localhost", "Hello, World!")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSendMessage_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"errcode": "M_FORBIDDEN", "error": "Not in room"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	err := client.SendMessage("!room:localhost", "Hello")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Not in room") {
		t.Errorf("expected error to contain 'Not in room', got %v", err)
	}
}

func TestGetJoinedRooms_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_matrix/client/v3/joined_rooms" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]string{
			"joined_rooms": {"!room1:localhost", "!room2:localhost"},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	rooms, err := client.GetJoinedRooms()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(rooms) != 2 {
		t.Errorf("expected 2 rooms, got %d", len(rooms))
	}
	if rooms[0] != "!room1:localhost" {
		t.Errorf("expected first room !room1:localhost, got %s", rooms[0])
	}
}

func TestGetJoinedRooms_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errcode": "M_UNKNOWN_TOKEN", "error": "Invalid token"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "bad_token"

	_, err := client.GetJoinedRooms()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetRoomMembers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"chunk": []map[string]string{
				{"state_key": "@user1:localhost"},
				{"state_key": "@user2:localhost"},
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	members, err := client.GetRoomMembers("!room:localhost")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}
	if members[0] != "@user1:localhost" {
		t.Errorf("expected first member @user1:localhost, got %s", members[0])
	}
}

func TestGetRoomMembers_NotAccessible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	members, err := client.GetRoomMembers("!room:localhost")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if members != nil {
		t.Errorf("expected nil members for inaccessible room, got %v", members)
	}
}

func TestFindDirectMessageRoom_Found(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/_matrix/client/v3/joined_rooms" {
			json.NewEncoder(w).Encode(map[string][]string{
				"joined_rooms": {"!room1:localhost", "!room2:localhost"},
			})
			return
		}

		if strings.Contains(r.URL.Path, "/members") {
			requestCount++
			if requestCount == 1 {
				// First room has 3 members (not a DM)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"chunk": []map[string]string{
						{"state_key": "@me:localhost"},
						{"state_key": "@other1:localhost"},
						{"state_key": "@other2:localhost"},
					},
				})
			} else {
				// Second room has 2 members (a DM with target user)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"chunk": []map[string]string{
						{"state_key": "@me:localhost"},
						{"state_key": "@target:localhost"},
					},
				})
			}
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	roomID, err := client.FindDirectMessageRoom("@target:localhost")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if roomID != "!room2:localhost" {
		t.Errorf("expected room !room2:localhost, got %s", roomID)
	}
}

func TestFindDirectMessageRoom_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/_matrix/client/v3/joined_rooms" {
			json.NewEncoder(w).Encode(map[string][]string{
				"joined_rooms": {"!room1:localhost"},
			})
			return
		}

		if strings.Contains(r.URL.Path, "/members") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"chunk": []map[string]string{
					{"state_key": "@me:localhost"},
					{"state_key": "@someone_else:localhost"},
				},
			})
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	roomID, err := client.FindDirectMessageRoom("@target:localhost")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if roomID != "" {
		t.Errorf("expected empty room ID, got %s", roomID)
	}
}

func TestGetOrCreateDirectMessage_ExistingRoom(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/_matrix/client/v3/joined_rooms" {
			json.NewEncoder(w).Encode(map[string][]string{
				"joined_rooms": {"!existing:localhost"},
			})
			return
		}

		if strings.Contains(r.URL.Path, "/members") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"chunk": []map[string]string{
					{"state_key": "@me:localhost"},
					{"state_key": "@target:localhost"},
				},
			})
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	roomID, isNew, err := client.GetOrCreateDirectMessage("@target:localhost")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if isNew {
		t.Error("expected isNew to be false for existing room")
	}
	if roomID != "!existing:localhost" {
		t.Errorf("expected room !existing:localhost, got %s", roomID)
	}
}

func TestGetOrCreateDirectMessage_NewRoom(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/_matrix/client/v3/joined_rooms" {
			json.NewEncoder(w).Encode(map[string][]string{
				"joined_rooms": {},
			})
			return
		}

		if r.URL.Path == "/_matrix/client/v3/createRoom" {
			json.NewEncoder(w).Encode(map[string]string{
				"room_id": "!newroom:localhost",
			})
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	roomID, isNew, err := client.GetOrCreateDirectMessage("@target:localhost")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !isNew {
		t.Error("expected isNew to be true for new room")
	}
	if roomID != "!newroom:localhost" {
		t.Errorf("expected room !newroom:localhost, got %s", roomID)
	}
}

func TestLeaveRoom_Success(t *testing.T) {
	leaveCalled := false
	forgetCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/_matrix/client/v3/rooms/!room:localhost/leave" && r.Method == "POST" {
			leaveCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}

		if r.URL.Path == "/_matrix/client/v3/rooms/!room:localhost/forget" && r.Method == "POST" {
			forgetCalled = true
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	err := client.LeaveRoom("!room:localhost")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !leaveCalled {
		t.Error("expected leave endpoint to be called")
	}
	if !forgetCalled {
		t.Error("expected forget endpoint to be called")
	}
}

func TestLeaveRoom_LeaveFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/_matrix/client/v3/rooms/!room:localhost/leave" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"errcode": "M_FORBIDDEN",
				"error":   "You are not in this room",
			})
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.accessToken = "test_token"

	err := client.LeaveRoom("!room:localhost")
	if err == nil {
		t.Error("expected error when leave fails")
	}
}
