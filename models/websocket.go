package models

import (
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

// This interface determines what data can be sent as data over the websocket.
type WebsocketMessageData interface {
	WebsocketData()
}

// This is what's received over the websocket.
type WebsocketMessageWrapper struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

func (wm *WebsocketMessageWrapper) ToMessage() (*WebsocketMessage, error) {
	var data WebsocketMessageData

	switch wm.Type {
	case "message":
		data = &Message{}
	case "channel":
		data = &Channel{}
	case "":
		return nil, fmt.Errorf("no type was provided to the message")
	default:
		data = nil
	}

	if data != nil {
		if err := json.Unmarshal(wm.Data, data); err != nil {
			return nil, err
		}
	}

	return &WebsocketMessage{wm.Type, data}, nil
}

// This is what is used in the handler to process and send data to the websocket server.
type WebsocketMessage struct {
	Type string               `json:"type"`
	Data WebsocketMessageData `json:"data,omitempty"`
}

type WebsocketCommand struct {
	Type                string
	UserId              int32
	Message             *WebsocketMessage
	Connection          *websocket.Conn
	NotificationTargets map[int32]bool
	AcknowledgeChannel  chan<- error
}

func NewWebsocketConnectCommand(userId int32, conn *websocket.Conn) (*WebsocketCommand, <-chan error) {
	errorChannel := make(chan error)
	return &WebsocketCommand{
			Type:               "connect",
			UserId:             userId,
			Message:            nil,
			Connection:         conn,
			AcknowledgeChannel: errorChannel,
		},
		errorChannel
}

func NewWebsocketDisconnectCommand(userId int32) (*WebsocketCommand, <-chan error) {
	errorChannel := make(chan error)
	return &WebsocketCommand{
			Type:               "disconnect",
			UserId:             userId,
			Message:            nil,
			AcknowledgeChannel: errorChannel,
		},
		errorChannel
}

func NewWebsocketMessageCommand(userId int32, data *WebsocketMessage, targets map[int32]bool) (*WebsocketCommand, <-chan error) {
	errorChannel := make(chan error)
	return &WebsocketCommand{
			Type:                "message",
			UserId:              userId,
			Message:             data,
			NotificationTargets: targets,
			AcknowledgeChannel:  errorChannel,
		},
		errorChannel
}
