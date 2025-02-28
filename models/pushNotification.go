package models

type PushNotificationRegistration struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"pd256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

type PushNotificationInfo struct {
	UserID   int32  `json:"user_id,omitempty" db:"user_id"`
	Endpoint string `json:"endpoint" db:"endpoint"`
	P256dh   string `json:"pd256dh" db:"pd256dh"`
	Auth     string `json:"auth" db:"auth"`
}
