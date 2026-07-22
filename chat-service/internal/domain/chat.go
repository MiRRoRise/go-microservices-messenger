package domain

import "time"

type Chat struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}
