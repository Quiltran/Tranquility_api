package models

type WebsocketResponse interface {
	Message
}

type WebsocketCommand[T WebsocketResponse] struct {
	Type               string
	UserId             int32
	Message            T
	Channel            chan<- T
	AcknowledgeChannel chan<- error
}

func NewWebsocketConnectCommand[T WebsocketResponse](userId int32) (*WebsocketCommand[T], <-chan T, <-chan error) {
	errorChannel := make(chan error)
	communicationChannel := make(chan T)
	var zero T
	return &WebsocketCommand[T]{
			Type:               "connect",
			UserId:             userId,
			Message:            zero,
			Channel:            communicationChannel,
			AcknowledgeChannel: errorChannel,
		},
		communicationChannel,
		errorChannel
}

func NewWebsocketDisconnectCommand[T WebsocketResponse](userId int32) (*WebsocketCommand[T], <-chan error) {
	errorChannel := make(chan error)
	var zero T
	return &WebsocketCommand[T]{
			Type:               "disconnect",
			UserId:             userId,
			Message:            zero,
			Channel:            nil,
			AcknowledgeChannel: errorChannel,
		},
		errorChannel
}

func NewWebsocketMessageCommand[T WebsocketResponse](userId int32, data T) (*WebsocketCommand[T], <-chan error) {
	errorChannel := make(chan error)
	var zero T
	return &WebsocketCommand[T]{
			Type:               "message",
			UserId:             userId,
			Message:            zero,
			Channel:            nil,
			AcknowledgeChannel: errorChannel,
		},
		errorChannel
}
