package models

import "encoding/json"

type PushNotificationRegistration struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

type PushNotificationInfo struct {
	UserID   int32  `json:"user_id,omitempty" db:"user_id"`
	Endpoint string `json:"endpoint" db:"endpoint"`
	P256dh   string `json:"p256dh" db:"p256dh"`
	Auth     string `json:"auth" db:"auth"`
}

type PushNotificationMessage struct {
	Title   string      `json:"title,omitempty"`
	Body    string      `json:"body,omitempty"`
	Icon    string      `json:"icon,omitempty"`
	Badge   string      `json:"badge,omitempty"`
	Tag     string      `json:"tag,omitempty"`
	Url     string      `json:"url,omitempty"`
	Actions []action    `json:"action,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (p *PushNotificationMessage) WithAction(title, actionName string) *PushNotificationMessage {
	if p.Actions == nil {
		p.Actions = []action{}
	}
	p.Actions = append(p.Actions, action{Action: actionName, Title: title})
	return p
}

func (p *PushNotificationMessage) Marhsal() ([]byte, error) {
	return json.Marshal(p)
}

type action struct {
	Action string `json:"action"`
	Title  string `json:"title"`
	Icon   string `json:"icon,omitempty"`
}

func NewPushNotificationMessage(title, body, url string, data interface{}) *PushNotificationMessage {
	return &PushNotificationMessage{
		Title: title,
		Body:  body,
		Icon:  "/favicon.png",
		Url:   url,
		Data:  data,
	}
}
