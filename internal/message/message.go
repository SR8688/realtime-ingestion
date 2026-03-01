package message

import "time"

type Data struct {
	ID         string
	DeviceID   int
	DataSource string
	Value      int
	TimeStamp  time.Time
}
