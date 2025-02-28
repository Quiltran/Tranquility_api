package models

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
