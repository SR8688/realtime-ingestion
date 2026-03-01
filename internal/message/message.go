package message

import "time"

type Data struct {
	id         string
	dataSource string
	value      int
	timeStamp  time.Time
}
