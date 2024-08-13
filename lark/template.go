// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package lark provides functionality for interacting with the Lark (Feishu) API,
// specifically for sending notifications using interactive cards.
package lark

import (
	"encoding/json"
	"strings"
	"sync"
	"text/template"
	"time"
)

// defaultCardMsgTmplData represents the data structure used to populate the card message template.
type defaultCardMsgTmplData struct {
	Content    string // The main content of the card
	Time       string // The time the card was created
	Title      string // The title of the card
	LevelColor string // The color of the card, indicating its level (e.g., success, error, warn)
}

// defaultCardMsgTmpl is the JSON template for the default card message.
// It defines the structure and styling of the interactive card,
// including the content, time, title, and color scheme.
const defaultCardMsgTmpl = `
{
    "config": {},
    "i18n_elements": {
        "zh_cn": [
            {
                "tag": "markdown",
                "content": "{{.Content}}",
                "text_align": "left",
                "text_size": "normal"
            },
            {
                "tag": "column_set",
                "flex_mode": "none",
                "background_style": "default",
                "horizontal_spacing": "8px",
                "horizontal_align": "left",
                "columns": [
                    {
                        "tag": "column",
                        "width": "weighted",
                        "vertical_align": "top",
                        "vertical_spacing": "8px",
                        "background_style": "default",
                        "elements": [
                            {
                                "tag": "hr"
                            },
                            {
                                "tag": "column_set",
                                "flex_mode": "none",
                                "horizontal_spacing": "default",
                                "background_style": "default",
                                "columns": [
                                    {
                                        "tag": "column",
                                        "elements": [
                                            {
                                                "tag": "div",
                                                "text": {
                                                    "tag": "plain_text",
                                                    "content": "{{.Time}}",
                                                    "text_size": "notation",
                                                    "text_align": "left",
                                                    "text_color": "grey"
                                                },
                                                "icon": {
                                                    "tag": "standard_icon",
                                                    "token": "time_outlined",
                                                    "color": "grey"
                                                }
                                            }
                                        ],
                                        "width": "weighted",
                                        "weight": 1
                                    }
                                ]
                            }
                        ],
                        "weight": 1,
                        "padding": "0px 0px 0px 0px"
                    }
                ],
                "margin": "16px 0px 0px 0px"
            }
        ]
    },
    "i18n_header": {
        "zh_cn": {
            "title": {
                "tag": "plain_text",
                "content": "{{.Title}}"
            },
            "subtitle": {
                "tag": "plain_text",
                "content": ""
            },
            "template": "{{.LevelColor}}"
        }
    }
}
`

// levelColorMap maps notification levels to their corresponding colors.
var levelColorMap = map[string]string{
	"success": "green",
	"error":   "red",
	"warn":    "yellow",
}

var (
	once     sync.Once
	cardTmpl *template.Template
)

// getDefaultCardMsgTmpl returns the parsed template for the default card message.
// It ensures that the template is only parsed once using sync.Once.
//
// Returns:
//   - *template.Template: The parsed template
//   - error: Any error encountered during template parsing
func getDefaultCardMsgTmpl() (*template.Template, error) {
	var err error
	once.Do(func() {
		// Parse the template only once
		cardTmpl, err = template.New("defaultCardMsg").Parse(defaultCardMsgTmpl)
	})
	return cardTmpl, err
}

// generateTextCardMsgWithLevel creates a card message based on the given level, title, and content.
//
// Parameters:
//   - level: The level of the notification (e.g., "success", "error", "warn")
//   - title: The title of the card
//   - content: The main content of the card
//
// Returns:
//   - map[string]any: A map representing the JSON structure of the card message
//   - error: Any error encountered during the process
//
// Example usage:
//
//	cardMsg, err := n.setCardMsg("success", "Task Completed", "Your task has been successfully completed.")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Use cardMsg to send the notification
func (n *notify) generateTextCardMsgWithLevel(level, title, content string) (map[string]any, error) {
	// Get the parsed template
	t, err := getDefaultCardMsgTmpl()
	if err != nil {
		return nil, err
	}

	// Determine the color based on the level
	levelColor, ok := levelColorMap[level]
	if !ok {
		levelColor = "blue" // Default color if level is not recognized
	}

	// Prepare the data for the template
	data := defaultCardMsgTmplData{
		Content:    content,
		Time:       time.Now().Format("2006-01-02 15:04:05"),
		Title:      title,
		LevelColor: levelColor,
	}

	// Execute the template with the prepared data
	var result strings.Builder
	err = t.Execute(&result, data)
	if err != nil {
		return nil, err
	}

	// Unmarshal the resulting JSON string into a map
	var prettyJSON map[string]any
	err = json.Unmarshal([]byte(result.String()), &prettyJSON)
	if err != nil {
		return nil, err
	}

	return prettyJSON, nil
}
