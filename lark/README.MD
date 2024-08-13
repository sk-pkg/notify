# Lark Notifier

## Overview

The Lark Notifier is a Go package that provides functionality for sending messages via the Lark (Feishu) platform. It supports sending messages through both bot webhooks and Lark Apps, offering a flexible and robust solution for integrating Lark messaging into your Go applications.

## Features

- Support for both Lark Bot Webhooks and Lark Apps
- Asynchronous message processing with goroutine pool
- Configurable channel and pool sizes
- Automatic token management for Lark Apps
- Support for various message types (text, interactive, image, etc.)
- Template card message support
- Message level-based card generation
- Caching mechanism for access tokens

## Installation

To use this package in your Go project, you can install it using:

```
go get github.com/sk-pkg/notify
```

## Configuration

To use the Lark Notifier, you need to create a `Config` struct with the following fields:

```go
type Config struct {
    Enabled                bool
    DefaultSendChannelName string
    ChannelSize            int
    PoolSize               int
    BotWebhooks            map[string]string
    Larks                  map[string]Lark
}
```

- `Enabled`: Set to true to enable the notifier.
- `DefaultSendChannelName`: The default channel name for sending messages when not specified in the message.
- `ChannelSize`: The buffer size for the message channel (defaults to 10 * GOMAXPROCS if set to 0).
- `PoolSize`: The number of goroutines in the worker pool (defaults to 10 * GOMAXPROCS if set to 0).
- `BotWebhooks`: A map of bot names to their corresponding webhook URLs.
- `Larks`: A map of Lark App configurations, keyed by a unique identifier for each app.

For Lark Apps, you need to provide the following information:

```go
type Lark struct {
    AppType   string
    Token     func() (string, error)
    AppID     string
    AppSecret string
}
```

- `AppType`: Either "feishu" or "lark", depending on which API you're using.
- `Token`: A function that returns the access token (optional if AppID and AppSecret are provided).
- `AppID`: The unique identifier for the Lark App.
- `AppSecret`: The secret key for the Lark App.

## Usage

### Creating a Notifier

```go
config := lark.Config{
    Enabled:                true,
    DefaultSendChannelName: "default_channel",
    BotWebhooks: map[string]string{
        "bot1": "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
    },
    Larks: map[string]lark.Lark{
        "app1": {
            AppType:   "feishu",
            AppID:     "your_app_id",
            AppSecret: "your_app_secret",
        },
    },
}

notifier, err := lark.New(config)
if err != nil {
    log.Fatalf("Failed to create notifier: %v", err)
}
```

### Starting the Processor

Before sending messages, start the message processor:

```go
notifier.StartProcessor()
```

### Sending Messages

#### Sending a Bot Message

```go
msgID, err := notifier.SendBotMessage("bot1", "text", "Hello, World!")
if err != nil {
    log.Printf("Failed to send bot message: %v", err)
}
```

#### Sending an App Message

```go
msgID, err := notifier.SendAppMessage("app1", "text", "user123", "Hello, User!")
if err != nil {
    log.Printf("Failed to send app message: %v", err)
}
```

#### Sending a Template Card Message

```go
templateID := "your_template_id"
content := map[string]interface{}{
    "title": "Card Title",
    "content": "Card Content",
}
msgID, err := notifier.SendTemplateCardMessage("app1", templateID, "user123", content)
if err != nil {
    log.Printf("Failed to send template card message: %v", err)
}
```

#### Submitting a Custom Message

```go
message := lark.Message{
    SendChannelName: "app1",
    SendTo:          "user123",
    MsgType:         "text",
    MsgLevel:        "info",
    Title:           "Important Message",
    Content:         "This is an important message.",
}
msgID, err := notifier.SubmitMessage(message)
if err != nil {
    log.Printf("Failed to submit message: %v", err)
}
```

### Closing the Notifier

When you're done sending messages, close the notifier to ensure all pending messages are processed:

```go
notifier.Close()
```

## Error Handling

The package provides error returns for all operations. Make sure to check these errors and handle them appropriately in your application.

## Concurrency and Performance

The package uses a goroutine pool for concurrent message processing. You can adjust the `ChannelSize` and `PoolSize` in the configuration to optimize performance based on your needs.

## Caching

The package implements caching for access tokens to reduce API calls. The default cache duration is the token expiration time minus 100 seconds.

## Contributing

Contributions to this package are welcome. Please ensure that your code adheres to the existing style and includes appropriate tests and documentation.

## License

This package is released under the MIT License. See the LICENSE file for details.