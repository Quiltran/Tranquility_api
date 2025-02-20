package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
	app.AddRoute("GET", "/ws", wc.echo)
}

func (wc *WebsocketController) echo(w http.ResponseWriter, r *http.Request) {
	limiter := rate.NewLimiter(rate.Every(time.Millisecond*100), 10)
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		wc.logger.ERROR(fmt.Sprintf("Error accepting ws conection: %v", err))
		return
	}
	defer c.CloseNow()

	handler := wc.websocketServer.NewHandler()
	err = handler.Connect(123, c)
	if err != nil {
		wc.logger.ERROR(fmt.Sprintf("Error connecting user to websocket server: %v", err))
		return
	}
	defer func() {
		if err := handler.Disconnect(123); err != nil {
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
			fmt.Println("handler", msg)
			handler.SendMessage(123, msg)
			fmt.Println("Message from client:", msg)
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
