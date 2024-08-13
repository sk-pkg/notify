// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package lark

import "testing"

func TestNotify_SendBotMessage(t *testing.T) {
	// Create a mock Notify instance
	mockNotify := newTestNotify()

	tests := []struct {
		name    string
		botName string
		msgType string
		content interface{}
	}{
		{
			name:    "Send interactive message",
			botName: "test_bot",
			msgType: "test",
			content: "test_bot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockNotify.SendBotMessage(tt.botName, tt.msgType, tt.content)
			if err != nil {
				t.Error(err)
			}

			// Check if the message was submitted correctly
			select {
			case msg := <-mockNotify.messages:
				if msg.SendChannelName != tt.botName {
					t.Errorf("Submitted message SendChannelName = %v, want %v", msg.SendChannelName, tt.botName)
				}
				if msg.MsgType != tt.msgType {
					t.Errorf("Submitted message MsgType = %v, want %v", msg.MsgType, tt.msgType)
				}
				if msg.Content != tt.content {
					t.Errorf("Submitted message Content = %v, want %v", msg.Content, tt.content)
				}
			default:
				t.Error("No message was submitted to the channel")
			}
		})
	}
}

func TestNotify_SendBotMessage_TextMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "text",
		Content: "hello world",
	}

	webhook, ok := n.botWebhooks["test_bot_1"]
	if !ok {
		t.Error("webhook not found")
	}

	err := n.sendBotWebhookMessage(webhook, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendBotMessage_ImageMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "image",
		Content: "img_ecffc3b9-8f14-400f-a014-05eca1a4310g",
	}

	webhook, ok := n.botWebhooks["test_bot_2"]
	if !ok {
		t.Error("webhook not found")
	}

	err := n.sendBotWebhookMessage(webhook, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendBotMessage_PostMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "post",
		Content: map[string]any{
			"zh_cn": map[string]any{
				"title": "test",
				"content": []any{
					[]any{
						map[string]any{
							"tag":  "text",
							"text": "test",
						},
					},
				},
			},
		},
	}

	webhook, ok := n.botWebhooks["test_bot_1"]
	if !ok {
		t.Error("webhook not found")
	}

	err := n.sendBotWebhookMessage(webhook, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendBotMessage_ShareChatMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "share_chat",
		Content: "oc_8c058d7ce5d70c4193c156625729866a",
	}

	webhook, ok := n.botWebhooks["test_bot_2"]
	if !ok {
		t.Error("webhook not found")
	}

	err := n.sendBotWebhookMessage(webhook, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendBotMessage_InteractiveMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "interactive",
		Content: map[string]any{
			"elements": []any{
				map[string]any{
					"tag": "div",
					"text": map[string]any{
						"content": "**West Lake**, located at No. 1 Longjing Road, West Lake District, Hangzhou City, Zhejiang Province, in the western part of Hangzhou urban area. The total area of the scenic area is 49 square kilometers, with a water catchment area of 21.22 square kilometers and a lake surface area of 6.38 square kilometers.",
						"tag":     "lark_md",
					},
				},
				map[string]any{
					"actions": []any{
						map[string]any{
							"tag": "button",
							"text": map[string]any{
								"content": "More Scenic Spots Introduction",
								"tag":     "lark_md",
							},
							"url":   "https://www.example.com",
							"type":  "default",
							"value": map[string]any{},
						},
					},
					"tag": "action",
				},
			},
			"header": map[string]any{
				"title": map[string]any{
					"content": "Today's Travel Recommendation",
					"tag":     "plain_text",
				},
			},
		}}

	webhook, ok := n.botWebhooks["test_bot_1"]
	if !ok {
		t.Error("webhook not found")
	}

	err := n.sendBotWebhookMessage(webhook, m)
	if err != nil {
		t.Error(err)
	}
}
