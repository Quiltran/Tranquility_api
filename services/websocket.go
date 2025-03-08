package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"tranquility/models"

	"github.com/coder/websocket"
)

// WebsocketServer should created in the main process, and passed to the WebsocketController as a pointer.
//
// This struct is in charge of sending communications and notificates between connections.
type WebsocketServer struct {
	// The mutex is required to not allow new users to be added until all messages are sent.
	mutex sync.Mutex
	// When the user connects to WebsocketServer, they pass their connection with it so that
	// we don't have to manage communication back to the requester.
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

func (ws *WebsocketServer) sendSystemMessage(data *models.WebsocketMessage, notificationTargets map[int32]bool) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	bytes, err := json.Marshal(data)
	if err != nil {
		ws.logger.ERROR(fmt.Sprintf("Error marshaling data: %+v", data))
		return
	}

	for userId, conn := range ws.users {
		if _, ok := notificationTargets[userId]; !ok {
			continue
		}
		ws.logger.INFO(fmt.Sprintf("Sending notification to %d", userId))
		w, err := conn.Writer(ws.shutdownContext, 1)
		if err != nil {
			ws.logger.ERROR(fmt.Sprintf("Error getting writer for connection %d: %v", userId, err))
			return
		}
		defer w.Close()

		_, err = w.Write(bytes)
		if err != nil {
			ws.logger.ERROR(fmt.Sprintf("Error writing to connection %d: %v", userId, err))
			return
		}
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
		ws.sendSystemMessage(command.Message, command.NotificationTargets)
		command.AcknowledgeChannel <- nil
	default:
		return fmt.Errorf("unknown command has been provided: %s", command.Type)
	}
	return nil
}

// # This function should be ran in a goroutine.
// This function allows for
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

// This function will be called any time a new websocket connection is created.
// Each connection has it's own WebsocketHandler so they can communicate to the WebsocketServer
// but not directly to each other.
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

func (wh *WebsocketHandler) SendMessage(userId int32, data *models.WebsocketMessage, receivers map[int32]bool) error {
	command, errorChannel := models.NewWebsocketMessageCommand(userId, data, receivers)

	wh.commandChannel <- *command

	if err := <-errorChannel; err != nil {
		return err
	}
	return nil
}
