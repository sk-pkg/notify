# Notify 

## Overview

The Notify is a comprehensive notification management system written in Go. It provides a unified interface for sending notifications through various channels, including Lark, DingTalk, WeChat, Email, Telegram, and Bark. This package is designed to be flexible and extensible, allowing users to easily configure and use multiple notification channels within their applications.

## Features

- Support for multiple notification channels:
    - Lark
    - DingTalk
    - WeChat
    - Email
    - Telegram
    - Bark
- Configurable default channel and notification level
- Message ID generation for tracking
- Asynchronous message processing
- Graceful shutdown of notification processors

## Installation

To use the Notify package in your Go project, you can install it using:

```
go get github.com/sk-pkg/notify
```

## Usage

### Initialization

To create a new Notify Manager, use the `New` function with appropriate options:

```go
import (
    "github.com/sk-pkg/notify"
    "github.com/sk-pkg/notify/lark"
    // Import other channel packages as needed
)

manager, err := notify.New(
    notify.OptLarkConfig(lark.Config{
        Enabled: true,
        // Other Lark-specific configurations
    }),
    notify.OptDefaultChannel(notify.LarkChan),
    notify.OptDefaultLevel(notify.InfoLevel),
    // Add other channel configurations and options as needed
)
if err != nil {
    // Handle error
}
```

### Sending Notifications

You can send notifications using the following methods:

```go
// Send a general notification
msgID, err := manager.Send(notify.InfoLevel, "recipient", "Title", "Content", notify.LarkChan)

// Send an info notification
msgID, err := manager.Info("recipient", "Info Title", "Info Content")

// Send a success notification
msgID, err := manager.Success("recipient", "Success Title", "Success Content")

// Send an error notification
msgID, err := manager.Error("recipient", "Error Title", "Error Content")

// Send a warning notification
msgID, err := manager.Warn("recipient", "Warning Title", "Warning Content")
```

### Closing the Manager

When you're done using the Notify Manager, make sure to close it to shut down all notification processors:

```go
manager.Close()
```

## Configuration Options

The Notify package provides several configuration options:

- `OptLarkConfig`: Configure Lark notifications
- `OptDingTalkConfig`: Configure DingTalk notifications
- `OptWechatConfig`: Configure WeChat notifications
- `OptEmailConfig`: Configure Email notifications
- `OptTelegramConfig`: Configure Telegram notifications
- `OptBarkConfig`: Configure Bark notifications
- `OptDefaultChannel`: Set the default notification channel
- `OptDefaultLevel`: Set the default notification level

## Supported Channels

1. Lark (`LarkChan`)
2. DingTalk (`DingTalkChan`)
3. WeChat (`WechatChan`)
4. Email (`EmailChan`)
5. Telegram (`TelegramChan`)
6. Bark (`BarkChan`)

### Lark-specific Documentation

For detailed information about the Lark notification channel, please refer to the [Lark README](https://github.com/sk-pkg/notify/blob/main/lark/README.MD).

## Notification Levels

The package supports four notification levels:

1. Info (`InfoLevel`)
2. Success (`SuccessLevel`)
3. Error (`ErrorLevel`)
4. Warning (`WarnLevel`)

## Error Handling

The package defines an `InvalidParams` error, which is returned when invalid parameters are provided to the notification methods.

## Concurrency and Goroutines

Each notification channel runs its own processor in a separate goroutine, allowing for asynchronous message handling.

## License

This package is governed by a MIT style license. Please refer to the LICENSE file for more details.

## Contributing

Contributions to the Notify package are welcome. Please ensure that your code adheres to the existing style and includes appropriate tests and documentation.

## Disclaimer

This README is based on the provided code snippet. Some features or functionalities might not be fully represented if they are implemented in other files or packages not shown in the given code.

Would you like me to explain or break down any part of this README or the original code?