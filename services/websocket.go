package services

import (
	"context"
	"fmt"
	"sync"
	"tranquility/models"
)

type WebsocketServer[T models.WebsocketResponse] struct {
	mutex           sync.Mutex
	users           map[int32]chan<- T
	commandChannel  chan models.WebsocketCommand[T]
	logger          Logger
	shutdownContext context.Context
}

func NewWebsocketServer[T models.WebsocketResponse](ctx context.Context, logger Logger) *WebsocketServer[T] {
	return &WebsocketServer[T]{
		users:           make(map[int32]chan<- T),
		commandChannel:  make(chan models.WebsocketCommand[T]),
		logger:          logger,
		shutdownContext: ctx,
	}
}

func (ws *WebsocketServer[T]) sendSystemMessage(data T) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	// This channel send is blocking.
	for userId, c := range ws.users {
		ws.logger.INFO(fmt.Sprintf("Sending notification to %d", userId))
		c <- data
	}
}

func (ws *WebsocketServer[T]) connect(userId int32, tx chan<- T) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.logger.INFO(fmt.Sprintf("Adding %d to connections", userId))
	ws.users[userId] = tx
}

func (ws *WebsocketServer[T]) disconnect(userId int32) error {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if _, ok := ws.users[userId]; !ok {
		ws.logger.ERROR(fmt.Sprintf("Tried removing %d from connections while it didn't exist.", userId))
		return fmt.Errorf("disconnect occurred while they were not in the map")
	}
	ws.logger.INFO(fmt.Sprintf("Removing %d from connections", userId))
	delete(ws.users, userId)
	return nil
}

func (ws *WebsocketServer[T]) handleCommand(command models.WebsocketCommand[T]) error {
	switch command.Type {
	case "connect":
		ws.connect(command.UserId, command.Channel)
		command.AcknowledgeChannel <- nil
	case "disconnect":
		err := ws.disconnect(command.UserId)
		command.AcknowledgeChannel <- err
	case "message":
		ws.sendSystemMessage(command.Message)
		command.AcknowledgeChannel <- nil
	default:
		fmt.Println("Unknown command has been provided")
	}
	return nil
}

func (ws *WebsocketServer[T]) Run(ctx context.Context) {
	select {
	case <-ws.shutdownContext.Done():
		return
	case command := <-ws.commandChannel:
		ws.logger.INFO(fmt.Sprintf("Websocket received command from %d", command.UserId))
		if err := ws.handleCommand(command); err != nil {
			fmt.Println("error handling command in websocket")
		}
	}
}

func (ws *WebsocketServer[T]) NewHandler() *WebsocketHandler[T] {
	return &WebsocketHandler[T]{
		commandChannel: ws.commandChannel,
	}
}

type WebsocketHandler[T models.WebsocketResponse] struct {
	commandChannel chan<- models.WebsocketCommand[T]
}

func (wh *WebsocketHandler[T]) Connect(userId int32) (<-chan T, error) {
	command, communicationListener, errorChannel := models.NewWebsocketConnectCommand[T](userId)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return nil, err
	}
	return communicationListener, nil
}

func (wh *WebsocketHandler[T]) Disconnect(userId int32) error {
	command, errorChannel := models.NewWebsocketDisconnectCommand[T](userId)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return err
	}
	return nil
}

func (wh *WebsocketHandler[T]) SendMessage(userId int32, data T) error {
	command, errorChannel := models.NewWebsocketMessageCommand(userId, data)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return err
	}
	return nil
}
