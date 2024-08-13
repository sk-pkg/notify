// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package lark

import (
	"encoding/json"
	"errors"
	"fmt"
)

type templateCardData struct {
	Type string `json:"type"`
	Data struct {
		TemplateID       string `json:"template_id"`
		TemplateVariable any    `json:"template_variable"`
	} `json:"data"`
}

// SendAppMessage sends a message via a Lark App.
//
// Parameters:
//   - appName: The name of the Lark App.
//   - msgType: The type of message to send.
//     Options: "text", "image", "audio", "file", "media", "sticker", "share_chat", "share_user", "post", "interactive", "system"
//   - sendTo: Lark(Feishu) user ID.
//   - content: The content of the message.
//
// Returns:
//   - string: The message ID if the message is sent successfully, an empty string otherwise.
//   - error: An error if the message cannot be sent, nil otherwise.
func (n *notify) SendAppMessage(appName, msgType, sendTo string, content any) (msgID string, err error) {
	return n.SubmitMessage(Message{
		Content:         content,
		MsgType:         msgType,
		SendChannelName: appName,
		SendTo:          sendTo,
	})
}

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
func (n *notify) SendTemplateCardMessage(appName, templateID, sendTo string, content any) (msgID string, err error) {
	msg := &templateCardData{Type: "template"}
	msg.Data.TemplateID = templateID
	msg.Data.TemplateVariable = content

	return n.SubmitMessage(Message{
		Content:         msg,
		MsgType:         "interactive",
		SendChannelName: appName,
		SendTo:          sendTo,
	})
}

// sendLarkAppMessage sends a message via a Lark App.
//
// Parameters:
//   - token: The access token for the Lark App.
//   - msgAPI: The API endpoint for sending Message.
//   - m: The Message struct containing the message details. Message.MsgType must be set.
//
// Returns:
//   - error: An error if the message cannot be sent, nil otherwise.
func (n *notify) sendLarkAppMessage(token, msgAPI string, m Message) error {
	var marshal []byte
	var err error

	params := map[string]string{"receive_id": m.SendTo, "msg_type": m.MsgType}

	switch m.MsgType {
	case "text":
		marshal, _ = json.Marshal(map[string]any{"text": m.Content.(string)})
		if err != nil {
			return fmt.Errorf("failed to marshal text content: %v", err)
		}
	case "image":
		marshal, err = json.Marshal(map[string]any{"image_key": m.Content.(string)})
		if err != nil {
			return fmt.Errorf("failed to marshal image content: %v", err)
		}
	case "audio", "file", "media", "sticker":
		marshal, err = json.Marshal(map[string]any{"file_key": m.Content.(string)})
		if err != nil {
			return fmt.Errorf("failed to marshal file content: %v", err)
		}
	case "share_chat":
		marshal, err = json.Marshal(map[string]any{"chat_id": m.Content.(string)})
		if err != nil {
			return fmt.Errorf("failed to marshal share chat content: %v", err)
		}
	case "share_user":
		marshal, err = json.Marshal(map[string]any{"user_id": m.Content.(string)})
		if err != nil {
			return fmt.Errorf("failed to marshal share user content: %v", err)
		}
	case "post", "interactive", "system":
		marshal, err = json.Marshal(m.Content)
		if err != nil {
			return fmt.Errorf("failed to marshal %s content: %v", m.MsgType, err)
		}
	default:
		return fmt.Errorf("invalid message type: %s", m.MsgType)
	}

	params["content"] = string(marshal)

	request := &Request{
		Method: "POST",
		URL:    msgAPI,
		Headers: map[string]string{
			"Content-Type":  "application/json; charset=utf-8",
			"Authorization": "Bearer " + token,
		},
		QueryParams: map[string]string{"receive_id_type": "user_id"},
		Body:        params,
	}

	response, err := n.sendLarkAPIRequest(request, 3)
	if err != nil {
		return fmt.Errorf("failed to send app message: %w", err)
	}

	// Check response status
	var rs messageResp
	if err = json.Unmarshal(response.Body, &rs); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if rs.Code != 0 {
		return errors.New(rs.Msg)
	}

	return nil
}
