package data_model

import (
	"time"
)

type User struct {
	Uid               string    `json:"uid" gorm:"primarykey"`
	Name              string    `json:"name"`
	EncryptedPassword string    `json:"encrypted_password"`
	ClientIP          string    `json:"client_ip"`
	ClientPort        string    `json:"client_port"`
	LoginTime         time.Time `json:"login_time"`
	HeartbeatTime     time.Time `json:"heartbeat_time"`
	LogoutTime        time.Time `json:"logout_time"`
	IsLogout          bool      `json:"is_logout"`
	DeviceInfo        string    `json:"device_info"`
}

func (table *User) TableName() string {
	return "user_basic"
}
