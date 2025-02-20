package models

import (
	"encoding/json"
	"fmt"

	"github.com/coder/websocket"
)

type WebsocketMessageData interface {
	WebsocketData()
}

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

type WebsocketMessage struct {
	Type string               `json:"type"`
	Data WebsocketMessageData `json:"data,omitempty"`
}

type WebsocketCommand struct {
	Type               string
	UserId             int32
	Message            *WebsocketMessage
	Connection         *websocket.Conn
	AcknowledgeChannel chan<- error
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

func NewWebsocketMessageCommand(userId int32, data *WebsocketMessage) (*WebsocketCommand, <-chan error) {
	errorChannel := make(chan error)
	return &WebsocketCommand{
			Type:               "message",
			UserId:             userId,
			Message:            data,
			AcknowledgeChannel: errorChannel,
		},
		errorChannel
}
