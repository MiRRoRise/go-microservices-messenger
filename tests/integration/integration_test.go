package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultGatewayURL = "http://localhost:9000"
	defaultAuthURL    = "http://localhost:9080"
	defaultChatURL    = "http://localhost:9081"
)

func gatewayURL() string {
	if v := os.Getenv("GATEWAY_URL"); v != "" {
		return v
	}
	return defaultGatewayURL
}

func authURL() string {
	if v := os.Getenv("AUTH_URL"); v != "" {
		return v
	}
	return gatewayURL()
}

func chatURL() string {
	if v := os.Getenv("CHAT_URL"); v != "" {
		return v
	}
	return gatewayURL()
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CreateChatRequest struct {
	Name string `json:"name"`
}

type CreateChatResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type CreateMessageRequest struct {
	Text string `json:"text"`
}

type CreateMessageResponse struct {
	ID        int64  `json:"id"`
	ChatID    int64  `json:"chat_id"`
	SenderID  int64  `json:"sender_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type ChatResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type ListChatsResponse struct {
	Chats []ChatResponse `json:"chats"`
}

type MessageResponse struct {
	ID        int64  `json:"id"`
	ChatID    int64  `json:"chat_id"`
	SenderID  int64  `json:"sender_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

type ListMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
}

func TestIntegration_FullFlow(t *testing.T) {
	t.Run("Full auth and chat flow", func(t *testing.T) {
		email := fmt.Sprintf("test%d@gmail.com", time.Now().UnixNano())
		password := "password"

		registerResp := registerUser(t, email, password)
		assert.NotZero(t, registerResp.ID)
		assert.Equal(t, email, registerResp.Email)

		loginResp := loginUser(t, email, password)
		assert.NotEmpty(t, loginResp.AccessToken)
		assert.NotEmpty(t, loginResp.RefreshToken)

		token := loginResp.AccessToken

		chatResp := createChat(t, token, "integration test")
		assert.NotZero(t, chatResp.ID)
		assert.Equal(t, "integration test", chatResp.Name)

		sendResp := sendMessage(t, token, "integration message", chatResp.ID)
		assert.NotZero(t, sendResp.ID)
		assert.Equal(t, chatResp.ID, sendResp.ChatID)
		assert.Equal(t, "integration message", sendResp.Text)

		messages := getMessages(t, token, chatResp.ID)
		assert.NotEmpty(t, messages)
		assert.GreaterOrEqual(t, len(messages.Messages), 1)

		chats := getChats(t, token)
		assert.NotEmpty(t, chats)
		assert.GreaterOrEqual(t, len(chats.Chats), 1)
	})

	t.Run("Invalid login should fail", func(t *testing.T) {
		resp, err := http.Post(
			authURL()+"/auth/login",
			"application/json",
			bytes.NewBuffer([]byte(`{"email":"nonexistent@test.com","password":"wrong"}`)),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Create chat without token should fail", func(t *testing.T) {
		resp, err := http.Post(
			chatURL()+"/chats",
			"application/json",
			bytes.NewBuffer([]byte(`{"name":"Unauthorized Chat"}`)),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func registerUser(t *testing.T, email, password string) RegisterResponse {
	reqBody := RegisterRequest{
		Email:    email,
		Password: password,
	}
	data, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp, err := http.Post(
		authURL()+"/auth/register",
		"application/json",
		bytes.NewBuffer(data),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result RegisterResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func loginUser(t *testing.T, email, password string) LoginResponse {
	reqBody := LoginRequest{
		Email:    email,
		Password: password,
	}
	data, err := json.Marshal(reqBody)
	require.NoError(t, err)

	resp, err := http.Post(
		authURL()+"/auth/login",
		"application/json",
		bytes.NewBuffer(data),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func createChat(t *testing.T, token, name string) CreateChatResponse {
	reqBody := CreateChatRequest{Name: name}
	data, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", chatURL()+"/chats", bytes.NewBuffer(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result CreateChatResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func sendMessage(t *testing.T, token, text string, chatID int64) CreateMessageResponse {
	reqBody := CreateMessageRequest{Text: text}
	data, err := json.Marshal(reqBody)
	require.NoError(t, err)

	url := fmt.Sprintf("%s/chats/%d/messages", chatURL(), chatID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result CreateMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func getChats(t *testing.T, token string) ListChatsResponse {
	req, err := http.NewRequest("GET", chatURL()+"/chats", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result ListChatsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	return result
}

func getMessages(t *testing.T, token string, chatID int64) ListMessagesResponse {
	url := fmt.Sprintf("%s/chats/%d/messages", chatURL(), chatID)
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result ListMessagesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result
}
