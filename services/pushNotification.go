package services

import (
	"errors"
	"fmt"
	"net/http"
	"tranquility/config"
	"tranquility/models"

	"github.com/SherClockHolmes/webpush-go"
)

var (
	ErrUnsubscribed = errors.New("user has unsubscribed")
)

type PushNotificationService struct {
	*config.PushNotificationConfig
	logger Logger
}

func NewPushNotificationService(config *config.PushNotificationConfig, logger Logger) *PushNotificationService {
	return &PushNotificationService{
		config,
		logger,
	}
}

func (p *PushNotificationService) SimplePush(subscription *webpush.Subscription, message *models.PushNotificationMessage) error {
	data, err := message.Marhsal()
	if err != nil {
		return fmt.Errorf("an error occurred while marshaling push notification message %v", err)
	}
	resp, err := webpush.SendNotification(data, subscription, &webpush.Options{
		Subscriber:      p.Sub,
		VAPIDPublicKey:  p.VapidPublicKey,
		VAPIDPrivateKey: p.VapidPrivateKey,
		TTL:             2419200,
		Urgency:         "normal",
		Topic:           "updates",
	})
	if err != nil {
		return fmt.Errorf("an error ocurred while trying to send push notification: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("an invalid status code was returned while sending push notification: %v", resp.StatusCode)
	}

	return nil
}
