package models

type WebsocketResponse interface {
	WebsocketData()
}

type WebsocketCommand struct {
	Type               string
	UserId             int32
	Message            WebsocketResponse
	Channel            chan<- WebsocketResponse
	AcknowledgeChannel chan<- error
}

func NewWebsocketConnectCommand(userId int32) (*WebsocketCommand, <-chan WebsocketResponse, <-chan error) {
	errorChannel := make(chan error)
	communicationChannel := make(chan WebsocketResponse)
	var zero WebsocketResponse
	return &WebsocketCommand{
			Type:               "connect",
			UserId:             userId,
			Message:            zero,
			Channel:            communicationChannel,
			AcknowledgeChannel: errorChannel,
		},
		communicationChannel,
		errorChannel
}

func NewWebsocketDisconnectCommand(userId int32) (*WebsocketCommand, <-chan error) {
	errorChannel := make(chan error)
	var zero WebsocketResponse
	return &WebsocketCommand{
			Type:               "disconnect",
			UserId:             userId,
			Message:            zero,
			Channel:            nil,
			AcknowledgeChannel: errorChannel,
		},
		errorChannel
}

func NewWebsocketMessageCommand(userId int32, data WebsocketResponse) (*WebsocketCommand, <-chan error) {
	errorChannel := make(chan error)
	var zero WebsocketResponse
	return &WebsocketCommand{
			Type:               "message",
			UserId:             userId,
			Message:            zero,
			Channel:            nil,
			AcknowledgeChannel: errorChannel,
		},
		errorChannel
}
