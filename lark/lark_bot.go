// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package lark

import (
	"encoding/json"
	"fmt"
)

// SendBotMessage sends a message via a Bot.
//
// Parameters:
//   - botName: The name of the bot.
//   - msgType: The type of message to send.
//     Options: "text", "post", "share_chat", "image", "interactive"
//   - content: The content of the message.
//
// Returns:
//   - string: The message ID.
//   - error: An error if the message cannot be sent, nil otherwise.
func (n *notify) SendBotMessage(botName, msgType string, content any) (msgID string, err error) {
	return n.SubmitMessage(Message{
		SendChannelName: botName,
		MsgType:         msgType,
		Content:         content,
	})
}

// sendBotWebhookMessage sends a message via a BotWebhook.
//
// Parameters:
//   - webhook: The webhook URL.
//   - m: The Message struct containing the message details.
//
// Returns:
//   - error: An error if the message cannot be sent, nil otherwise.
func (n *notify) sendBotWebhookMessage(webhook string, m Message) error {
	params := make(map[string]any)
	params["msg_type"] = m.MsgType

	switch m.MsgType {
	case "text":
		params["content"] = map[string]string{"text": m.Content.(string)}
	case "post":
		params["content"] = map[string]any{"post": m.Content}
	case "share_chat":
		params["content"] = map[string]string{"share_chat_id": m.Content.(string)}
	case "image":
		params["content"] = map[string]string{"image_key": m.Content.(string)}
	case "interactive":
		cardJson, err := json.Marshal(m.Content)
		if err != nil {
			return fmt.Errorf("marshal card content failed: %w", err)
		}
		params["card"] = string(cardJson)
	}

	request := &Request{
		Method:  "POST",
		URL:     webhook,
		Headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
		Body:    params,
	}

	response, err := n.sendLarkAPIRequest(request, 3)
	if err != nil {
		return fmt.Errorf("failed to send bot message: %w", err)
	}

	// Check response status
	var rs messageResp
	if err = json.Unmarshal(response.Body, &rs); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if rs.Code != 0 {
		return fmt.Errorf("failed to send bot message: %s", rs.Msg)
	}

	return nil
}
