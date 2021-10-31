package database

import (
	"fmt"
)

func KeyMessageAckIndex(accont string) string {
	return fmt.Sprintf("chat:ack:%s", accont)
}
