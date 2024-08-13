// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ding

import "github.com/sk-pkg/notify/msgid"

type (
	Config struct {
		Enabled bool

		ChannelSize int
		PoolSize    int
		IdleSize    int
	}

	Notify interface {
		StartProcessor()
		SubmitMessage(message Message) (msgID string)
		Close()
	}

	notify struct {
		msgID    *msgid.ID
		messages chan Message
	}

	Message struct {
		ID      string
		Title   string
		Content string
	}
)

func (n *notify) Close() {
	return
}

func (n *notify) StartProcessor() {

}

func (n *notify) SubmitMessage(message Message) (msgID string) {
	if message.ID == "" {
		message.ID = n.msgID.New()
	}

	n.messages <- message

	return message.ID
}

func New(config Config) (Notify, error) {
	return &notify{
		messages: make(chan Message, config.ChannelSize),
		msgID:    msgid.NewMessageID(),
	}, nil
}
