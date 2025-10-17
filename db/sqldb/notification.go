package sqldb

type Notification struct {
	PID     uint32 // process ID of the backend that sent the notification
	Channel string // channel name
	Payload string // message payload
}
