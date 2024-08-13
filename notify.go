// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package notify provides a unified interface for sending notifications
// through various channels such as Lark, DingTalk, WeChat, Email, Telegram, and Bark.
// It allows for easy configuration and management of multiple notification channels.
package notify

import (
	"errors"
	"github.com/sk-pkg/notify/bark"
	"github.com/sk-pkg/notify/ding"
	"github.com/sk-pkg/notify/email"
	"github.com/sk-pkg/notify/lark"
	"github.com/sk-pkg/notify/msgid"
	"github.com/sk-pkg/notify/telegram"
	"github.com/sk-pkg/notify/wechat"
	"log"
)

// Constants for supported notification channels and message levels
const (
	LarkChan     Channel = "lark"
	DingTalkChan Channel = "ding_talk"
	WechatChan   Channel = "wechat"
	EmailChan    Channel = "email"
	TelegramChan Channel = "telegram"
	BarkChan     Channel = "bark"

	InfoLevel    Level = "info"
	SuccessLevel Level = "success"
	ErrorLevel   Level = "error"
	WarnLevel    Level = "warn"
)

var (
	// InvalidParams is returned when invalid parameters are provided
	InvalidParams = errors.New("invalid params")
)

// Option is a function type for configuring the Manager
type Option func(*option)

// option holds the configuration options for the Manager
type option struct {
	defaultChannel Channel
	defaultLevel   Level

	larkConfig     lark.Config
	dingTalkConfig ding.Config
	wechatConfig   wechat.Config
	telegramConfig telegram.Config
	barkConfig     bark.Config
	emailConfig    email.Config
}

// Channel represents a notification channel
type Channel string

// Level represents the severity level of a message
type Level string

// Manager is the main struct for managing notifications across different channels
type Manager struct {
	defaultChannel Channel
	defaultLevel   Level

	messageID *msgid.ID

	Lark     lark.Notify
	DingTalk ding.Notify
	Wechat   wechat.Notify
	Telegram telegram.Notify
	Bark     bark.Notify
	Email    email.Notify

	channelStatus map[Channel]bool
}

// OptLarkConfig sets the Lark configuration for the Manager
//
// Parameters:
//   - config: The Lark configuration to be set
//
// Returns:
//   - Option: A function that sets the Lark configuration
func OptLarkConfig(config lark.Config) Option {
	return func(o *option) {
		o.larkConfig = config
	}
}

// OptDingTalkConfig sets the DingTalk configuration for the Manager
//
// Parameters:
//   - config: The DingTalk configuration to be set
//
// Returns:
//   - Option: A function that sets the DingTalk configuration
func OptDingTalkConfig(config ding.Config) Option {
	return func(o *option) {
		o.dingTalkConfig = config
	}
}

// OptWechatConfig sets the WeChat configuration for the Manager
//
// Parameters:
//   - config: The WeChat configuration to be set
//
// Returns:
//   - Option: A function that sets the WeChat configuration
func OptWechatConfig(config wechat.Config) Option {
	return func(o *option) {
		o.wechatConfig = config
	}
}

// OptTelegramConfig sets the Telegram configuration for the Manager
//
// Parameters:
//   - config: The Telegram configuration to be set
//
// Returns:
//   - Option: A function that sets the Telegram configuration
func OptTelegramConfig(config telegram.Config) Option {
	return func(o *option) {
		o.telegramConfig = config
	}
}

// OptBarkConfig sets the Bark configuration for the Manager
//
// Parameters:
//   - config: The Bark configuration to be set
//
// Returns:
//   - Option: A function that sets the Bark configuration
func OptBarkConfig(config bark.Config) Option {
	return func(o *option) {
		o.barkConfig = config
	}
}

// OptEmailConfig sets the Email configuration for the Manager
//
// Parameters:
//   - config: The Email configuration to be set
//
// Returns:
//   - Option: A function that sets the Email configuration
func OptEmailConfig(config email.Config) Option {
	return func(o *option) {
		o.emailConfig = config
	}
}

// OptDefaultChannel sets the default notification channel for the Manager
//
// Parameters:
//   - channel: The default channel to be set
//
// Returns:
//   - Option: A function that sets the default channel
func OptDefaultChannel(channel Channel) Option {
	return func(o *option) {
		o.defaultChannel = channel
	}
}

// OptDefaultLevel sets the default message level for the Manager
//
// Parameters:
//   - level: The default level to be set
//
// Returns:
//   - Option: A function that sets the default level
func OptDefaultLevel(level Level) Option {
	return func(o *option) {
		o.defaultLevel = level
	}
}

// New creates a new Manager instance with the provided options
//
// Parameters:
//   - options: A variadic list of Option functions to configure the Manager
//
// Returns:
//   - *Manager: A pointer to the newly created Manager
//   - error: An error if any occurred during creation
//
// Example:
//
//	manager, err := New(
//	    OptLarkConfig(larkConfig),
//	    OptDefaultChannel(LarkChan),
//	    OptDefaultLevel(InfoLevel),
//	)
//	if err != nil {
//	    log.Fatalf("Failed to create Manager: %v", err)
//	}
func New(options ...Option) (*Manager, error) {
	var err error
	opt := &option{} // Initialize opt

	// Apply all provided options
	for _, f := range options {
		f(opt)
	}

	// Create a new Manager instance
	m := &Manager{
		channelStatus: map[Channel]bool{
			LarkChan:     opt.larkConfig.Enabled,
			DingTalkChan: opt.dingTalkConfig.Enabled,
			WechatChan:   opt.wechatConfig.Enabled,
			TelegramChan: opt.telegramConfig.Enabled,
			BarkChan:     opt.barkConfig.Enabled,
			EmailChan:    opt.emailConfig.Enabled,
		},
	}

	// Initialize enabled channels
	if opt.larkConfig.Enabled {
		m.Lark, err = lark.New(opt.larkConfig)
		if err != nil {
			return m, err
		}
		m.Lark.StartProcessor()
	}

	if opt.dingTalkConfig.Enabled {
		m.DingTalk, err = ding.New(opt.dingTalkConfig)
		if err != nil {
			return m, err
		}
		m.DingTalk.StartProcessor()
	}

	if opt.wechatConfig.Enabled {
		m.Wechat, err = wechat.New(opt.wechatConfig)
		if err != nil {
			return m, err
		}
		m.Wechat.StartProcessor()
	}

	if opt.telegramConfig.Enabled {
		m.Telegram, err = telegram.New(opt.telegramConfig)
		if err != nil {
			return m, err
		}
		m.Telegram.StartProcessor()
	}

	if opt.barkConfig.Enabled {
		m.Bark, err = bark.New(opt.barkConfig)
		if err != nil {
			return m, err
		}
		m.Bark.StartProcessor()
	}

	if opt.emailConfig.Enabled {
		m.Email, err = email.New(opt.emailConfig)
		if err != nil {
			return m, err
		}
		m.Email.StartProcessor()
	}

	// Set default channel and level
	m.defaultChannel = opt.defaultChannel
	m.defaultLevel = opt.defaultLevel
	m.messageID = msgid.NewMessageID()

	return m, nil
}

// submit is an internal method to submit a message to specified channels
//
// Parameters:
//   - level: The severity level of the message
//   - sendTo: The recipient of the message
//   - title: The title of the message
//   - content: The content of the message
//   - channels: A variadic list of channels to send the message through
//
// Returns:
//   - string: The message ID
//   - error: An error if any occurred during submission
func (m *Manager) submit(level Level, sendTo, title, content string, channels ...Channel) (string, error) {
	// Validate input parameters
	if title == "" && content == "" {
		return "", InvalidParams
	}

	// Use default channel if none specified
	channelCount := len(channels)
	if channelCount == 0 {
		channels = append(channels, m.defaultChannel)
	}

	// Use default level if none specified
	if level == "" {
		level = m.defaultLevel
	}

	// Generate a new message ID
	msgID := m.messageID.New()

	var err error
	// Submit message to each specified channel
	for _, channel := range channels {
		switch channel {
		case LarkChan:
			_, err = m.Lark.SubmitMessage(lark.Message{
				ID:       msgID,
				SendTo:   sendTo,
				MsgLevel: string(level),
				Title:    title,
				Content:  content,
			})
			if err != nil {
				log.Println(err)
			}
		case DingTalkChan:
			m.DingTalk.SubmitMessage(ding.Message{
				ID:      msgID,
				Title:   title,
				Content: content,
			})
		case WechatChan:
			m.Wechat.SubmitMessage(wechat.Message{
				ID:      msgID,
				Title:   title,
				Content: content,
			})
		case EmailChan:
			m.Email.SubmitMessage(email.Message{
				ID:      msgID,
				Title:   title,
				Content: content,
			})
		case TelegramChan:
			m.Telegram.SubmitMessage(telegram.Message{
				ID:      msgID,
				Title:   title,
				Content: content,
			})
		case BarkChan:
			m.Bark.SubmitMessage(bark.Message{
				ID:      msgID,
				Title:   title,
				Content: content,
			})
		}
	}

	return msgID, nil
}

// Send submits a message with a specified level to the given channels
//
// Parameters:
//   - level: The severity level of the message
//   - sendTo: The recipient of the message
//   - title: The title of the message
//   - content: The content of the message
//   - channels: A variadic list of channels to send the message through
//
// Returns:
//   - string: The message ID
//   - error: An error if any occurred during sending
//
// Example:
//
//	msgID, err := manager.Send(InfoLevel, "user123", "Test Message", "This is a test", LarkChan, EmailChan)
//	if err != nil {
//	    log.Printf("Failed to send message: %v", err)
//	}
func (m *Manager) Send(level Level, sendTo, title, content string, channels ...Channel) (string, error) {
	return m.submit(level, sendTo, title, content, channels...)
}

// Info sends an info level message to the specified channels
//
// Parameters:
//   - sendTo: The recipient of the message
//   - title: The title of the message
//   - content: The content of the message
//   - channels: A variadic list of channels to send the message through
//
// Returns:
//   - string: The message ID
//   - error: An error if any occurred during sending
//
// Example:
//
//	msgID, err := manager.Info("user123", "Info Message", "This is an informational message", LarkChan)
//	if err != nil {
//	    log.Printf("Failed to send info message: %v", err)
//	}
func (m *Manager) Info(sendTo, title, content string, channels ...Channel) (string, error) {
	return m.submit(InfoLevel, sendTo, title, content, channels...)
}

// Success sends a success level message to the specified channels
//
// Parameters:
//   - sendTo: The recipient of the message
//   - title: The title of the message
//   - content: The content of the message
//   - channels: A variadic list of channels to send the message through
//
// Returns:
//   - string: The message ID
//   - error: An error if any occurred during sending
//
// Example:
//
//	msgID, err := manager.Success("user123", "Success Message", "Operation completed successfully", EmailChan)
//	if err != nil {
//	    log.Printf("Failed to send success message: %v", err)
//	}
func (m *Manager) Success(sendTo, title, content string, channels ...Channel) (string, error) {
	return m.submit(SuccessLevel, sendTo, title, content, channels...)
}

// Error sends an error level message to the specified channels
//
// Parameters:
//   - sendTo: The recipient of the message
//   - title: The title of the message
//   - content: The content of the message
//   - channels: A variadic list of channels to send the message through
//
// Returns:
//   - string: The message ID
//   - error: An error if any occurred during sending
//
// Example:
//
//	msgID, err := manager.Error("user123", "Error Message", "An error occurred: Database connection failed", LarkChan, EmailChan)
//	if err != nil {
//	    log.Printf("Failed to send error message: %v", err)
//	}
func (m *Manager) Error(sendTo, title, content string, channels ...Channel) (string, error) {
	return m.submit(ErrorLevel, sendTo, title, content, channels...)
}

// Warn sends a warning level message to the specified channels
//
// Parameters:
//   - sendTo: The recipient of the message
//   - title: The title of the message
//   - content: The content of the message
//   - channels: A variadic list of channels to send the message through
//
// Returns:
//   - string: The message ID
//   - error: An error if any occurred during sending
//
// Example:
//
//	msgID, err := manager.Warn("user123", "Warning Message", "System resources are running low", DingTalkChan)
//	if err != nil {
//	    log.Printf("Failed to send warning message: %v", err)
//	}
func (m *Manager) Warn(sendTo, title, content string, channels ...Channel) (string, error) {
	return m.submit(WarnLevel, sendTo, title, content, channels...)
}

// Close gracefully shuts down all enabled notification channels
//
// This method should be called when the Manager is no longer needed to ensure
// proper cleanup of resources.
func (m *Manager) Close() {
	// Iterate through all channels and close enabled ones
	for channel, status := range m.channelStatus {
		if status {
			switch channel {
			case LarkChan:
				m.Lark.Close()
			case DingTalkChan:
				m.DingTalk.Close()
			case WechatChan:
				m.Wechat.Close()
			case EmailChan:
				m.Email.Close()
			case TelegramChan:
				m.Telegram.Close()
			case BarkChan:
				m.Bark.Close()
			}
		}
	}
}
