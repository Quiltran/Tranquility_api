package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"tranquility/models"

	"github.com/coder/websocket"
)

type WebsocketServer struct {
	mutex sync.Mutex
	// The channel here is for the server to data back to the handler
	users map[int32]*websocket.Conn
	// This is used for handlers to send commands to the server
	commandChannel  chan models.WebsocketCommand
	logger          Logger
	shutdownContext context.Context
}

func NewWebsocketServer(ctx context.Context, logger Logger) *WebsocketServer {
	return &WebsocketServer{
		users:           make(map[int32]*websocket.Conn),
		commandChannel:  make(chan models.WebsocketCommand),
		logger:          logger,
		shutdownContext: ctx,
	}
}

func (ws *WebsocketServer) sendSystemMessage(data models.WebsocketMessage) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	bytes, err := json.Marshal(data)
	if err != nil {
		ws.logger.ERROR(fmt.Sprintf("Error marshaling data: %+v", data))
		return
	}
	fmt.Println(string(bytes))
	for userId, conn := range ws.users {
		ws.logger.INFO(fmt.Sprintf("Sending notification to %d", userId))
		w, err := conn.Writer(ws.shutdownContext, 1)
		if err != nil {
			ws.logger.ERROR(fmt.Sprintf("Error getting writer for connection %d: %v", userId, err))
			return
		}
		defer w.Close()

		x, err := w.Write(bytes)
		if err != nil {
			ws.logger.ERROR(fmt.Sprintf("Error writing to connection %d: %v", userId, err))
			return
		}
		fmt.Println(x)
		ws.logger.INFO(fmt.Sprintf("Notification sent %d", userId))
	}
}

func (ws *WebsocketServer) connect(userId int32, conn *websocket.Conn) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.logger.INFO(fmt.Sprintf("Adding %d to connections", userId))
	ws.users[userId] = conn
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
		ws.connect(command.UserId, command.Connection)
		command.AcknowledgeChannel <- nil
	case "disconnect":
		err := ws.disconnect(command.UserId)
		command.AcknowledgeChannel <- err
	case "message":
		ws.sendSystemMessage(*command.Message)
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
				ws.logger.ERROR(fmt.Sprintf("an error occurred while handling websocket command: %+v", command))
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

func (wh *WebsocketHandler) Connect(userId int32, conn *websocket.Conn) error {
	command, errorChannel := models.NewWebsocketConnectCommand(userId, conn)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return err
	}
	return nil
}

func (wh *WebsocketHandler) Disconnect(userId int32) error {
	command, errorChannel := models.NewWebsocketDisconnectCommand(userId)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return err
	}
	return nil
}

func (wh *WebsocketHandler) SendMessage(userId int32, data *models.WebsocketMessage) error {
	command, errorChannel := models.NewWebsocketMessageCommand(userId, data)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return err
	}
	return nil
}
