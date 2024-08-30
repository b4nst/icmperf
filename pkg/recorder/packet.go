package recorder

import "time"

type Packet struct {
	Size     int
	Sent     time.Time
	Received time.Time
}
