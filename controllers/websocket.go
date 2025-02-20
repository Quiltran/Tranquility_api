package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"tranquility/app"
	"tranquility/data"
	"tranquility/models"
	"tranquility/services"

	"github.com/coder/websocket"
	"golang.org/x/time/rate"
)

var (
	clientTimeout = 10 * time.Second
)

type WebsocketController struct {
	db              data.IDatabase
	logger          services.Logger
	websocketServer *services.WebsocketServer
}

func NewWebsocketController(db data.IDatabase, logger services.Logger, websocketServer *services.WebsocketServer) *WebsocketController {
	return &WebsocketController{
		db,
		logger,
		websocketServer,
	}
}

func (wc *WebsocketController) RegisterRoutes(app *app.App) {
	app.AddRoute("GET", "/ws/{id}/{token}", wc.Websocket)
}

func (wc *WebsocketController) Websocket(w http.ResponseWriter, r *http.Request) {
	limiter := rate.NewLimiter(rate.Every(time.Millisecond*100), 10)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	userId, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil {
		handleError(w, wc.logger, err, nil, http.StatusBadRequest, "warning")
		return
	}
	websocketToken := r.PathValue("token")

	user, err := wc.db.WebsocketLogin(ctx, int32(userId), websocketToken)
	if err != nil {
		if err == sql.ErrNoRows {
			handleError(w, wc.logger, err, nil, http.StatusUnauthorized, "warning")
			return
		}
		handleError(w, wc.logger, err, nil, http.StatusInternalServerError, "error")
		return
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		wc.logger.ERROR(fmt.Sprintf("Error accepting ws conection: %v", err))
		return
	}
	defer c.CloseNow()

	handler := wc.websocketServer.NewHandler()
	err = handler.Connect(user.ID, c)
	if err != nil {
		wc.logger.ERROR(fmt.Sprintf("Error connecting user to websocket server: %v", err))
		return
	}
	defer func() {
		if err := handler.Disconnect(user.ID); err != nil {
			wc.logger.ERROR(fmt.Sprintf("Error disconnecting user to websocket server: %v", err))
		}
	}()

	incoming := make(chan *models.WebsocketMessage)
	errChan := make(chan error)
	ping := make(chan struct{})

	go func() {
		defer close(incoming)
		for {
			isPing, err := handleConnection(ctx, c, limiter, incoming)
			if err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure || err == context.Canceled {
					return
				}
				errChan <- err
			}
			if isPing {
				ping <- struct{}{}
			}
		}
	}()

	ticker := time.NewTicker(clientTimeout)
	lastHeartbeat := time.Now()

	for {
		select {
		case <-ticker.C:
			if time.Since(lastHeartbeat) > clientTimeout {
				wc.logger.ERROR("connection timeout, disconnecting")
				return
			}
		case <-ping:
			lastHeartbeat = time.Now()
		case msg := <-incoming:
			msg, err := wc.handleIncomingMessage(ctx, user, msg)
			if err != nil {
				wc.logger.ERROR(fmt.Sprintf("an error occurred while handling request: %v", err))
				return
			}
			handler.SendMessage(user.ID, msg)
		case err := <-errChan:
			wc.logger.ERROR(fmt.Sprintf("error reading from websocket: %v", err))
			return
		case <-ctx.Done():
			return
		}
	}
}

func handleConnection(ctx context.Context, conn *websocket.Conn, limiter *rate.Limiter, incoming chan<- *models.WebsocketMessage) (bool, error) {
	err := limiter.Wait(ctx)
	if err != nil {
		return false, err
	}

	typ, r, err := conn.Read(ctx)
	if err != nil {
		return false, err
	}

	if typ != websocket.MessageText {
		return false, fmt.Errorf("unexpected message type: %d", typ)
	}

	var message models.WebsocketMessageWrapper
	err = json.Unmarshal(r, &message)
	if err != nil {
		return false, err
	}
	if message.Type == "Ping" {
		return true, nil
	}
	data, err := message.ToMessage()
	if err != nil {
		return false, err
	}

	select {
	case incoming <- data:
	case <-ctx.Done():
		return false, ctx.Err()
	}

	return false, nil
}

func (wc *WebsocketController) handleIncomingMessage(ctx context.Context, user *models.AuthUser, message *models.WebsocketMessage) (*models.WebsocketMessage, error) {

	switch message.Type {
	case "message":
		output, err := wc.db.CreateMessage(ctx, message.Data.(*models.Message), user.ID)
		if err != nil {
			return nil, err
		}
		message.Data = output
	default:
		wc.logger.ERROR(fmt.Sprintf("an unknown message type was handled by handleIncomingMessage: %s", message.Type))
		return nil, fmt.Errorf("an unknown message type was passed")
	}

	return message, nil
}
