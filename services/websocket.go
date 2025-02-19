package services

import (
	"context"
	"fmt"
	"sync"
	"tranquility/models"
)

type WebsocketServer struct {
	mutex           sync.Mutex
	users           map[int32]chan<- models.WebsocketResponse
	commandChannel  chan models.WebsocketCommand
	logger          Logger
	shutdownContext context.Context
}

func NewWebsocketServer(ctx context.Context, logger Logger) *WebsocketServer {
	return &WebsocketServer{
		users:           make(map[int32]chan<- models.WebsocketResponse),
		commandChannel:  make(chan models.WebsocketCommand),
		logger:          logger,
		shutdownContext: ctx,
	}
}

func (ws *WebsocketServer) sendSystemMessage(data models.WebsocketResponse) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	// This channel send is blocking.
	for userId, c := range ws.users {
		ws.logger.INFO(fmt.Sprintf("Sending notification to %d", userId))
		c <- data
	}
}

func (ws *WebsocketServer) connect(userId int32, tx chan<- models.WebsocketResponse) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.logger.INFO(fmt.Sprintf("Adding %d to connections", userId))
	ws.users[userId] = tx
}

func (ws *WebsocketServer) disconnect(userId int32) error {
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

func (ws *WebsocketServer) handleCommand(command models.WebsocketCommand) error {
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

func (ws *WebsocketServer) Run() {
	ws.logger.INFO("Websocket server has started...")
	for {
		select {
		case <-ws.shutdownContext.Done():
			ws.logger.INFO("Websocket server is shutting down...")
			return
		case command := <-ws.commandChannel:
			ws.logger.INFO(fmt.Sprintf("Websocket received command from %d", command.UserId))
			if err := ws.handleCommand(command); err != nil {
				fmt.Println("error handling command in websocket")
			}
		}
	}
}

func (ws *WebsocketServer) NewHandler() *WebsocketHandler {
	return &WebsocketHandler{
		commandChannel: ws.commandChannel,
	}
}

type WebsocketHandler struct {
	commandChannel chan<- models.WebsocketCommand
}

func (wh *WebsocketHandler) Connect(userId int32) (<-chan models.WebsocketResponse, error) {
	command, communicationListener, errorChannel := models.NewWebsocketConnectCommand(userId)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return nil, err
	}
	return communicationListener, nil
}

func (wh *WebsocketHandler) Disconnect(userId int32) error {
	command, errorChannel := models.NewWebsocketDisconnectCommand(userId)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return err
	}
	return nil
}

func (wh *WebsocketHandler) SendMessage(userId int32, data models.WebsocketResponse) error {
	command, errorChannel := models.NewWebsocketMessageCommand(userId, data)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return err
	}
	return nil
}
