// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package lark

import "testing"

func TestNotify_SendAppMessage(t *testing.T) {
	mockNotify := newTestNotify()

	tests := []struct {
		name      string
		appName   string
		msgType   string
		sendTo    string
		content   interface{}
		wantMsgID string
	}{
		{
			name:      "Send text message",
			appName:   "test_app",
			msgType:   "text",
			sendTo:    "user123",
			content:   "Hello, world!",
			wantMsgID: "mock_message_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mockNotify.SendAppMessage(tt.appName, tt.msgType, tt.sendTo, tt.content)
			if err != nil {
				t.Errorf("SendAppMessage() error = %v", err)
			}

			// Check if the message was submitted correctly
			select {
			case msg := <-mockNotify.messages:
				if msg.SendChannelName != tt.appName {
					t.Errorf("Submitted message SendChannelName = %v, want %v", msg.SendChannelName, tt.appName)
				}
				if msg.MsgType != tt.msgType {
					t.Errorf("Submitted message MsgType = %v, want %v", msg.MsgType, tt.msgType)
				}
				if msg.SendTo != tt.sendTo {
					t.Errorf("Submitted message SendTo = %v, want %v", msg.SendTo, tt.sendTo)
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

func TestNotify_SendAppMessage_TextMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "text",
		SendTo:  "seakee",
		Content: "hello world",
	}

	a, ok := n.apps["lark_1"]
	if !ok {
		t.Error("app not found")
	}

	tk, err := a.token()
	if err != nil {
		t.Error(err)
	}

	err = n.sendLarkAppMessage(tk, a.msgAPI, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendAppMessage_ImageMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "image",
		SendTo:  "seakee",
		Content: "img_ecffc3b9-8f14-400f-a014-05eca1a4310g",
	}

	a, ok := n.apps["feishu_1"]
	if !ok {
		t.Error("app not found")
	}

	tk, err := a.token()
	if err != nil {
		t.Error(err)
	}

	err = n.sendLarkAppMessage(tk, a.msgAPI, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendAppMessage_PostMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "post",
		SendTo:  "seakee",
		Content: map[string]interface{}{
			"zh_cn": map[string]interface{}{
				"title": "I am a title",
				"content": [][]map[string]interface{}{
					{
						{
							"tag":   "text",
							"text":  "First line:",
							"style": []string{"bold", "underline"},
						},
						{
							"tag":   "a",
							"href":  "http://www.feishu.cn",
							"text":  "Hyperlink",
							"style": []string{"bold", "italic"},
						},
						{
							"tag":     "at",
							"user_id": "seakee",
							"style":   []string{"lineThrough"},
						},
					},
					{
						{
							"tag":       "img",
							"image_key": "img_7ea74629-9191-4176-998c-2e603c9c5e8g",
						},
					},
					{
						{
							"tag":   "text",
							"text":  "Second line:",
							"style": []string{"bold", "underline"},
						},
						{
							"tag":  "text",
							"text": "Text test",
						},
					},
					{
						{
							"tag":        "emotion",
							"emoji_type": "SMILE",
						},
					},
					{
						{
							"tag": "hr",
						},
					},
					{
						{
							"tag":      "code_block",
							"language": "GO",
							"text":     "func main() int64 {\n    return 0\n}",
						},
					},
					{
						{
							"tag":  "md",
							"text": "**mention user:**<at user_id=\"ou_xxxxxx\">Tom</at>\n**href:**[Open Platform](https://open.feishu.cn)\n**code block:**\n```GO\nfunc main() int64 {\n    return 0\n}\n```\n**text styles:** **bold**, *italic*, ***bold and italic***, ~underline~,~~lineThrough~~\n> quote content\n\n1. item1\n    1. item1.1\n    2. item2.2\n2. item2\n --- \n- item1\n    - item1.1\n    - item2.2\n- item2",
						},
					},
				},
			},
		},
	}

	a, ok := n.apps["lark_2"]
	if !ok {
		t.Error("app not found")
	}

	tk, err := a.token()
	if err != nil {
		t.Error(err)
	}

	err = n.sendLarkAppMessage(tk, a.msgAPI, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendAppMessage_InteractiveMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "interactive",
		SendTo:  "seakee",
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
		},
	}

	a, ok := n.apps["feishu_2"]
	if !ok {
		t.Error("app not found")
	}

	tk, err := a.token()
	if err != nil {
		t.Error(err)
	}

	err = n.sendLarkAppMessage(tk, a.msgAPI, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendAppMessage_ShareChatMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "share_chat",
		SendTo:  "seakee",
		Content: "oc_xxxxx",
	}

	a, ok := n.apps["lark_1"]
	if !ok {
		t.Error("app not found")
	}

	tk, err := a.token()
	if err != nil {
		t.Error(err)
	}

	err = n.sendLarkAppMessage(tk, a.msgAPI, m)
	if err != nil {
		t.Error(err)
	}
}

func TestNotify_SendAppMessage_ShareUserMessage(t *testing.T) {
	n := newTestNotify()

	m := Message{
		MsgType: "share_user",
		SendTo:  "seakee",
		Content: "ou_xxxxxx",
	}

	a, ok := n.apps["feishu_1"]
	if !ok {
		t.Error("app not found")
	}

	tk, err := a.token()
	if err != nil {
		t.Error(err)
	}

	err = n.sendLarkAppMessage(tk, a.msgAPI, m)
	if err != nil {
		t.Error(err)
	}
}
