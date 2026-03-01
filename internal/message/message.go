package message

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const (
	amps = "amps"
)

type Data struct {
	ID         string
	DeviceID   int
	DataSource string
	Value      int
	TimeStamp  time.Time
}

func NewPrefixedUUID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String())
}
func NewAmpsData(deviceID int) Data {
	data := Data{
		ID:         NewPrefixedUUID("data"),
		DeviceID:   deviceID,
		DataSource: amps,
		Value:      rand.Intn(20) + 10, //10 to 30 amps
		TimeStamp:  time.Now(),
	}
	return data
}
