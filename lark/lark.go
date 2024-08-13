// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package lark provides functionality for sending messages via Lark (Feishu) platform.
// It supports sending messages through both bot webhooks and Lark Apps.
package lark

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/panjf2000/ants/v2"
	"github.com/sk-pkg/notify/cache"
	"github.com/sk-pkg/notify/msgid"
	"github.com/sk-pkg/notify/util"
	"log"
	"runtime"
	"sync"
)

// Constants used throughout the package
const (
	// feishuHost is the base URL for Feishu API calls.
	feishuHost = "https://open.feishu.cn"

	// larkHost is the base URL for Lark API calls.
	larkHost = "https://open.larksuite.com"

	// appAccessTokenAPI is the URL for retrieving the app access token.
	appAccessTokenAPI = "/open-apis/auth/v3/app_access_token/internal/"

	// messageAPI is the URL for sending messages through Lark Apps.
	messageAPI = "/open-apis/im/v1/messages"

	// tokenCacheKey is the key used to store the app access token in the cache.
	// %s will be replaced with the Lark app ID.
	tokenCacheKey = "lark:token:%s"
)

// Config represents the configuration for the Lark notifier.
type Config struct {
	// Enabled indicates whether the notifier is active. Set to true to enable the notifier.
	Enabled bool

	// DefaultSendChannelName is the default channel name for sending messages when not specified in the message.
	// This must be set to a valid channel name from either BotWebhooks or Larks.
	DefaultSendChannelName string

	// ChannelSize defines the buffer size for the message channel.
	// If set to 0, it defaults to 10 * GOMAXPROCS.
	ChannelSize int

	// PoolSize defines the number of goroutines in the worker pool.
	// If set to 0, it defaults to 10 * GOMAXPROCS.
	PoolSize int

	// BotWebhooks is a map of bot names to their corresponding webhook URLs.
	// Use this to configure message sending via bot webhooks.
	// The key will be used as the send channel name.
	BotWebhooks map[string]string

	// Larks is a map of Lark App configurations, keyed by a unique identifier for each app.
	// Use this to configure message sending via Lark Apps.
	// The key will be used as the send channel name.
	Larks map[string]Lark
}

// Lark represents the configuration for a Lark App.
type Lark struct {
	// AppType is the type of Lark App.
	//
	// Options:
	// 	- feishu: If the app is a feishu app, the feishu API will be used.
	// 	- lark: If the app is a lark app, the lark API will be used.
	AppType string

	// Token is a function that returns the access token for the Lark App.
	// If not provided, a default function using AppID and AppSecret will be used.
	Token func() (string, error)

	// AppID is the unique identifier for the Lark App.
	// This is required if Token is not provided.
	AppID string

	// AppSecret is the secret key for the Lark App.
	// This is required if Token is not provided.
	AppSecret string
}

// Notify is the interface that wraps the basic methods for the notifier.
// It provides functionality for starting the message processor, submitting messages,
// and gracefully closing the notifier.
type Notify interface {
	// StartProcessor initiates the message processing routine.
	// It should be called once before submitting any messages.
	// This method typically starts a goroutine that continuously processes
	// submitted messages from an internal channel.
	StartProcessor()

	// SendAppMessage sends a message through a specific Lark App.
	//
	// Parameters:
	// 	- appName: The name of the Lark App to send the message through.
	// 	- msgType: The type of message (e.g., interactive, image, share_chat, post, text).
	// 	- sendTo: Lark(Feishu) user ID.
	// 	- content: The content of the message, which can be of any type depending on the message format.
	//
	// Returns:
	// 	- msgID: A unique identifier for the submitted message.
	// 	- err: An error that occurred while sending the message.
	SendAppMessage(appName, msgType, sendTo string, content any) (msgID string, err error)

	// SendTemplateCardMessage sends a template card message via a Lark App.
	//
	// Parameters:
	//   - appName: The name of the Lark App.
	//   - templateID: The ID of the template to use.
	//   - sendTo: Lark(Feishu) user ID.
	//   - content: The data of the template card.
	//
	// Returns:
	//   - string: The message ID if the message is sent successfully, an empty string otherwise.
	//   - error: An error if the message cannot be sent, nil otherwise.
	SendTemplateCardMessage(appName, templateID, sendTo string, content any) (msgID string, err error)

	// SendBotMessage sends a message through a specific bot.
	//
	// Parameters:
	// 	- botName: The name of the bot to send the message through.
	// 	- msgType: The type of message (e.g., interactive, image, share_chat, post, text).
	// 	- content: The content of the message, which can be of any type depending on the message format.
	//
	// Returns:
	// 	- msgID: A unique identifier for the submitted message.
	// 	- err: An error that occurred while sending the message.
	SendBotMessage(botName, msgType string, content any) (msgID string, err error)

	// SubmitMessage adds a new message to the processing queue.
	// This method is used to send a message through the notifier.
	// The message will be processed asynchronously by the processor started with StartProcessor.
	//
	// Parameters:
	// 	- message: The Message struct containing all necessary information for sending the notification.
	//
	// Returns:
	// 	- msgID: A unique identifier for the submitted message.
	// 	- err: An error that occurred while submitting the message.
	SubmitMessage(message Message) (msgID string, err error)

	// Token retrieves the access token for a specific Lark App.
	//
	// Parameters:
	// 	- appName: The name of the Lark App.
	//
	// Returns:
	// 	- token: The access token for the Lark App.
	// 	- err: An error that occurred while retrieving the token.
	Token(appName string) (token string, err error)

	// Close stops the notifier, ensuring all pending messages are processed
	// before shutting down. It waits for all goroutines to finish and releases any resources.
	// After calling Close, the notifier should not be used anymore.
	Close()
}

// SendResult represents the result of a message send operation.
type SendResult struct {
	// MsgID is a unique identifier for the submitted message.
	MsgID string

	// State is the state of the message.
	// Possible values:
	// 	- success: The message was sent successfully.
	// 	- failed: An error occurred while sending the message.
	// 	- pending: The message is still being processed.
	State string

	// Err is an error that occurred while sending the message.
	// If State is "failed", this field contains the error.
	Err error
}

// notify implements the Notify interface.
type notify struct {
	// defaultSendChannelName is the default channel name for sending messages when not specified in the message.
	defaultSendChannelName string

	// msgID is ID instance used for generating unique message IDs.
	msgID *msgid.ID

	// messages is a channel for buffering incoming messages before processing.
	messages chan Message

	// request is a resty client used for making HTTP requests to the Lark API.
	request *resty.Client

	// apps is a map of Lark App names to their corresponding configurations.
	// The key will be used as the send channel name.
	apps map[string]*app

	// botWebhooks is a map of bot names to their corresponding webhook URLs.
	// The key will be used as the send channel name.
	botWebhooks map[string]string

	// pool is a goroutine pool used for concurrent message processing.
	pool *ants.PoolWithFunc

	// wg is used to wait for all goroutines to finish before closing the notifier.
	wg sync.WaitGroup

	// cache is a cache instance used for caching tokens.
	cache cache.Cache

	// sendResult is a channel for sending the result of each message send operation.
	sendResult chan SendResult
}

// Token retrieves the access token for a specific Lark App.
//
// Parameters:
//   - appName: The name of the Lark App.
//
// Returns:
//   - token: The access token for the Lark App.
//   - err: An error that occurred while retrieving the token.
func (n *notify) Token(appName string) (token string, err error) {
	a, ok := n.apps[appName]
	if !ok {
		return "", fmt.Errorf("lark app %s not found", appName)
	}

	return a.token()
}

// app represents the configuration for a Lark App.
type app struct {
	// token is a function that returns the access token for the Lark App.
	token func() (string, error)

	// msgAPI is the URL for sending messages through Lark Apps.
	msgAPI string
}

// Message represents a message to be sent via the notifier.
type Message struct {
	// ID is a unique identifier for the message. It can be used for tracking or deduplication.
	ID string

	// SendChannelName specifies the channel through which the message should be sent.
	// If empty, the DefaultSendChannelName from the Config will be used.
	SendChannelName string

	// SendTo Lark(Feishu) user ID
	SendTo string

	// MsgType specifies the type of message
	// It must be set to a valid value.
	//
	// Bot message support interactive, image, share_chat, post, text.
	//
	// Lark App message support text, image, audio, file, media, sticker, share_chat, share_user,
	// post, interactive, system.
	MsgType string

	// MsgLevel indicates the importance or category of the message (e.g., "info", "warn", "error", "success").
	MsgLevel string

	// Title is the title or subject of the message.It only used for text message.
	Title string

	// Content contains the main body of the message. It can be of any type, depending on the message type.
	// More details about the content format can be found in the Lark API documentation.
	// https://open.larksuite.com/document/server-docs/im-v1/message-content-description/create_json
	Content any
}

// appTokenResp represents the response from the Lark App Token API.
type appTokenResp struct {
	Code           int    `json:"code"`             // Response code, 0 indicates success
	Msg            string `json:"msg"`              // Error message if the request failed
	AppAccessToken string `json:"app_access_token"` // The access token for the Lark App
	Expire         int    `json:"expire"`           // Token expiration time in seconds
}

// messageResp represents the response from the Lark Message API.
type messageResp struct {
	Code int    `json:"code"` // Response code, 0 indicates success
	Msg  string `json:"msg"`  // Error message if the request failed
}

// validateConfig checks the provided configuration for validity.
// It returns an error if the configuration is invalid.
//
// Parameters:
//   - config: A pointer to the Config struct to be validated.
//
// Returns:
//   - error: An error if the configuration is invalid, nil otherwise.
func validateConfig(config *Config) error {
	webhookCount := len(config.BotWebhooks)
	larkCount := len(config.Larks)

	// Check if there are any available sending channels
	if webhookCount == 0 && larkCount == 0 {
		return errors.New("There are no available sending channels for lark. Please configure BotWebhooks or Larks ")
	}

	// Check if DefaultSendChannelName is set
	if config.DefaultSendChannelName == "" {
		return errors.New("DefaultLarkSendChannelName is required")
	}

	// Log a warning if only BotWebhook is configured
	if webhookCount != 0 && larkCount == 0 {
		log.Println("Only the BotWebhook channel is detected. Messages will only be sent via BotWebhook.")
	}

	// Check Lark configurations if only Lark Apps are configured
	if webhookCount == 0 && larkCount != 0 {
		for name, l := range config.Larks {
			switch {
			case l.Token == nil && l.AppID == "" && l.AppSecret == "":
				fallthrough
			case l.AppID != "" && l.AppSecret == "":
				fallthrough
			case l.AppID == "" && l.AppSecret != "":
				return fmt.Errorf("lark config error: %s", name)
			}
		}

		log.Println("only the Lark App channel is detected. Messages will only be sent via Lark App")
	}

	// Set default values for ChannelSize and PoolSize if not provided
	// Default to GOMAXPROCS * 10
	if config.ChannelSize == 0 {
		config.ChannelSize = 10 * runtime.GOMAXPROCS(0)
	}

	// Set default values for ChannelSize and PoolSize if not provided
	// Default to GOMAXPROCS * 10
	if config.PoolSize == 0 {
		config.PoolSize = 10 * runtime.GOMAXPROCS(0)
	}

	return nil
}

// StartProcessor starts the message processing goroutine.
// It continuously reads messages from the channel and submits them to the goroutine pool.
func (n *notify) StartProcessor() {
	go func() {
		for m := range n.messages {
			n.wg.Add(1)
			err := n.pool.Invoke(m)
			if err != nil {
				log.Printf("failed to submit lark task to pool: %v\n", err)
			}
		}
	}()
}

// SubmitMessage submits a message to the notifier's message channel.
//
// Parameters:
//   - message: The Message struct to be submitted.
//
// Returns:
//   - msgID: The ID of the submitted message. If the message ID is not provided, a new one will be generated.
//   - error: Any error encountered during the process.
func (n *notify) SubmitMessage(message Message) (msgID string, err error) {
	// Generate a new message ID if not provided
	if message.ID == "" {
		message.ID = n.msgID.New()
	}

	// Check if we need to generate a card message using the message level and title
	if shouldGenerateCardMsg(message) {
		content, ok := message.Content.(string)
		if ok {
			cardContent, err := n.generateTextCardMsgWithLevel(message.MsgLevel, message.Title, content)
			if err != nil {
				return message.ID, fmt.Errorf("failed to generate card message: %w", err)
			}

			message.MsgType = "interactive"
			message.Content = cardContent
		}
	}

	// Submit the message to the channel
	n.messages <- message

	return message.ID, nil
}

// shouldGenerateCardMsg checks if a card message should be generated based on the message properties.
func shouldGenerateCardMsg(message Message) bool {
	return message.MsgLevel != "" &&
		message.Title != "" &&
		(message.MsgType == "" || message.MsgType == "text")
}

// New creates a new Notify instance with the provided configuration.
//
// Parameters:
//   - config: The Config struct containing the notifier configuration.
//
// Returns:
//   - Notify: A new Notify instance.
//   - error: An error if the configuration is invalid or if the goroutine pool cannot be created.
func New(config Config) (Notify, error) {
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	n := &notify{
		msgID:                  msgid.NewMessageID(),
		messages:               make(chan Message, config.ChannelSize),
		request:                resty.New(),
		apps:                   make(map[string]*app),
		botWebhooks:            config.BotWebhooks,
		defaultSendChannelName: config.DefaultSendChannelName,
		sendResult:             make(chan SendResult, config.ChannelSize),
		cache:                  cache.New(),
	}

	// Create a new goroutine pool
	pool, err := ants.NewPoolWithFunc(config.PoolSize, func(i interface{}) {
		msg := i.(Message)
		err := n.sendMsg(msg)
		if err != nil {
			log.Printf("failed to send lark message: %v\n", err)
		}

		n.wg.Done()
	}, ants.WithPreAlloc(true))
	if err != nil {
		return nil, fmt.Errorf("failed to create lark goroutine pool: %v", err)
	}

	n.pool = pool

	// Initialize Lark Apps
	for name, lark := range config.Larks {
		var msgAPI, appTokenAPI string
		switch lark.AppType {
		case "feishu":
			msgAPI = util.SpliceStr(feishuHost, messageAPI)
			appTokenAPI = util.SpliceStr(feishuHost, appAccessTokenAPI)
		case "lark":
			msgAPI = util.SpliceStr(larkHost, messageAPI)
			appTokenAPI = util.SpliceStr(larkHost, appAccessTokenAPI)
		default:
			return nil, fmt.Errorf("invalid lark app type: %s", lark.AppType)
		}

		// If Token is not provided,
		// lark.AppID and lark.AppSecret will be used to generate the token.
		if lark.Token == nil {
			if lark.AppID == "" || lark.AppSecret == "" {
				return nil, fmt.Errorf("lark config error: %s", name)
			}

			lark.Token = func() (string, error) {
				return n.getToken(lark.AppID, lark.AppSecret, appTokenAPI)
			}
		}

		a := &app{
			token:  lark.Token,
			msgAPI: msgAPI,
		}

		n.apps[name] = a
	}

	return n, nil
}

// getToken retrieves an access token for a Lark App.
//
// Parameters:
//   - appID: The App ID for the Lark App.
//   - appSecret: The App Secret for the Lark App.
//   - appTokenAPI: The API endpoint for obtaining the Lark App Token.
//
// Returns:
//   - string: The access token if successful, an empty string otherwise.
//   - error: An error if the token cannot be retrieved, nil otherwise.
func (n *notify) getToken(appID, appSecret, appTokenAPI string) (string, error) {
	cacheKey := fmt.Sprintf(tokenCacheKey, appID)

	token, err := n.cache.GetString(cacheKey)
	if err == nil && token != "" {
		return token, nil
	}

	request := &Request{
		Method: "POST",
		URL:    appTokenAPI,
		Headers: map[string]string{
			"Content-Type": "application/json; charset=utf-8",
		},
		Body: map[string]string{
			"app_id":     appID,
			"app_secret": appSecret,
		},
	}

	response, err := n.sendLarkAPIRequest(request, 3)
	if err != nil {
		return "", fmt.Errorf("failed to request Lark App Token: %w", err)
	}

	var rs appTokenResp
	if err = json.Unmarshal(response.Body, &rs); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if rs.Code != 0 {
		return "", fmt.Errorf("failed to obtain Lark App Token: %s", rs.Msg)
	}

	if err = n.cache.SetString(cacheKey, rs.AppAccessToken, rs.Expire-100); err != nil {
		log.Printf("failed to cache token: %s", err)
	}

	return rs.AppAccessToken, nil
}

// sendMsg sends a message using the appropriate channel (BotWebhook or Lark App).
//
// Parameters:
//   - m: The Message struct containing the message details.
//
// Returns:
//   - error: An error if the message cannot be sent, nil otherwise.
func (n *notify) sendMsg(m Message) error {
	channel := m.SendChannelName
	if channel == "" {
		channel = n.defaultSendChannelName
	}

	// Check if the channel is a BotWebhook
	if webhook, ok := n.botWebhooks[channel]; ok {
		return n.sendBotWebhookMessage(webhook, m)
	}

	// Check if the channel is a Lark App
	if a, ok := n.apps[channel]; ok {
		if m.SendTo == "" {
			return fmt.Errorf("sendTo is required for lark app %s", channel)
		}

		t, err := a.token()
		if err != nil {
			return fmt.Errorf("failed to get token for lark app %s: %w", channel, err)
		}

		return n.sendLarkAppMessage(t, a.msgAPI, m)
	}

	return fmt.Errorf("channel %s is not found in lark", channel)
}

// Close stops the notifier, waits for all messages to be processed, and releases resources.
func (n *notify) Close() {
	// Close the message channel to stop accepting new messages
	close(n.messages)

	// Wait for all messages to be processed
	n.wg.Wait()

	// Release the goroutine pool
	n.pool.Release()

	log.Println("Lark notify closed")
}
