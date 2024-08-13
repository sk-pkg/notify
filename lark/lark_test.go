// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package lark

import (
	"github.com/go-resty/resty/v2"
	"github.com/sk-pkg/notify/msgid"
	"github.com/sk-pkg/notify/util"
	"testing"
)

const (
	feishuTestToken = "t-xxxxxxx"
	larkTestToken   = "t-xxxxxxx"

	feishuTestAppID     = "cli_xxxxxx"
	feishuTestAppSecret = "xxxxxxx"

	larkTestAppID     = "cli_xxxxx"
	larkTestAppSecret = "xxxxxx"

	botTestWebhook1 = "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxx"
	botTestWebhook2 = "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxx"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config with bot webhook",
			config: Config{
				Enabled:                true,
				DefaultSendChannelName: "default_bot",
				BotWebhooks: map[string]string{
					"default_bot": "https://example.com/webhook",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid config with Lark app",
			config: Config{
				Enabled:                true,
				DefaultSendChannelName: "default_app",
				Larks: map[string]Lark{
					"default_app": {
						AppType:   "feishu",
						AppID:     "test_app_id",
						AppSecret: "test_app_secret",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid config - no channels",
			config: Config{
				Enabled:                true,
				DefaultSendChannelName: "default_channel",
			},
			wantErr: true,
		},
		{
			name: "Invalid config - missing DefaultSendChannelName",
			config: Config{
				Enabled: true,
				BotWebhooks: map[string]string{
					"test_bot": "https://example.com/webhook",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func getFeishuToken() (string, error) {
	return feishuTestToken, nil
}

func getLarkToken() (string, error) {
	return larkTestToken, nil
}

func newTestNotifyInterface() (Notify, error) {
	cfg := Config{
		Enabled:                true,
		DefaultSendChannelName: "test_bot_1",
		BotWebhooks: map[string]string{
			"test_bot_1": botTestWebhook1,
			"test_bot_2": botTestWebhook2,
		},
		Larks: map[string]Lark{
			"lark_1": {
				AppType:   "lark",
				AppID:     larkTestAppID,
				AppSecret: larkTestAppSecret,
			},
			"lark_2": {
				AppType: "lark",
				Token:   getLarkToken,
			},
			"feishu_1": {
				AppType:   "feishu",
				AppID:     feishuTestAppID,
				AppSecret: feishuTestAppSecret,
			},
			"feishu_2": {
				AppType: "feishu",
				Token:   getFeishuToken,
			},
		},
	}

	return New(cfg)
}

func newTestNotify() *notify {
	botWebhooks := map[string]string{
		"test_bot_1": botTestWebhook1,
		"test_bot_2": botTestWebhook2,
	}

	n := &notify{
		msgID:                  msgid.NewMessageID(),
		messages:               make(chan Message, 10),
		request:                resty.New(),
		botWebhooks:            botWebhooks,
		defaultSendChannelName: "test_bot_1",
	}

	apps := map[string]*app{
		"lark_1": {
			msgAPI: util.SpliceStr(larkHost, messageAPI),
			token: func() (string, error) {
				return n.getToken(larkTestAppID, larkTestAppSecret, util.SpliceStr(larkHost, appAccessTokenAPI))
			},
		},
		"lark_2": {
			msgAPI: util.SpliceStr(larkHost, messageAPI),
			token:  getLarkToken,
		},
		"feishu_1": {
			msgAPI: util.SpliceStr(feishuHost, messageAPI),
			token: func() (string, error) {
				return n.getToken(feishuTestAppID, feishuTestAppSecret, util.SpliceStr(feishuHost, appAccessTokenAPI))
			},
		},
		"feishu_2": {
			msgAPI: util.SpliceStr(feishuHost, messageAPI),
			token:  getFeishuToken,
		},
	}

	n.apps = apps

	return n
}

func TestNotify_SubmitMessage(t *testing.T) {
	// Create a mock Notify instance
	mockNotify := newTestNotify()

	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "Submit message with ID",
			message: Message{
				ID:              "existing_id",
				SendChannelName: "test_channel",
				MsgType:         "text",
				Content:         "Test message",
			},
		},
		{
			name: "Submit message without ID",
			message: Message{
				SendChannelName: "test_channel",
				MsgType:         "text",
				Content:         "Test message",
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := mockNotify.SubmitMessage(tt.message)
			if err != nil {
				t.Errorf("SubmitMessage() error = %v", err)
			}

			// Check if the message was submitted to the channel
			select {
			case msg := <-mockNotify.messages:
				if msg.ID != gotID {
					t.Errorf("Submitted message ID = %v, want %v", msg.ID, gotID)
				}
			default:
				t.Error("No message was submitted to the channel")
			}
		})
	}
}

func TestNotify_SubmitMessage_by_interface(t *testing.T) {
	n, err := newTestNotifyInterface()
	if err != nil {
		t.Errorf("newTestNotifyInterface() error = %v", err)
	}

	n.StartProcessor()

	defer n.Close()

	tests := []struct {
		name    string
		message Message
	}{
		{
			name: "use test_bot_1 channel",
			message: Message{
				SendChannelName: "test_bot_1",
				MsgLevel:        "info",
				Title:           "Bot Test",
				Content:         "Bot Test message 1",
			},
		},
		{
			name: "use test_bot_2 channel",
			message: Message{
				SendChannelName: "test_bot_2",
				MsgLevel:        "success",
				Title:           "Bot Test",
				Content:         "Bot Test message 2",
			},
		},
		{
			name: "use lark_1 channel",
			message: Message{
				SendChannelName: "lark_1",
				MsgType:         "text",
				MsgLevel:        "warn",
				SendTo:          "seakee",
				Title:           "Lark Test",
				Content:         "lark_1 message",
			},
		},
		{
			name: "use lark_2 channel",
			message: Message{
				SendChannelName: "lark_2",
				MsgType:         "text",
				MsgLevel:        "error",
				SendTo:          "seakee",
				Title:           "Lark Test",
				Content:         "lark_2 message",
			},
		},
		{
			name: "use feishu_1 channel",
			message: Message{
				SendChannelName: "feishu_1",
				MsgType:         "text",
				Content:         "feishu_1 message",
			},
		},
		{
			name: "use feishu_2 channel",
			message: Message{
				SendChannelName: "feishu_2",
				MsgType:         "text",
				Content:         "feishu_2 message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = n.SubmitMessage(tt.message)
			if err != nil {
				t.Errorf("SubmitMessage() error = %v", err)
			}
		})
	}
}
