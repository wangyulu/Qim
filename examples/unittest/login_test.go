package unittest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"jinv/kim"
	"jinv/kim/examples/dialer"
	"jinv/kim/websocket"
)

func login(account string) (kim.Client, error) {
	cli := websocket.NewClient(account, "unittest", websocket.ClientOptions{})

	cli.SetDialer(&dialer.ClientDialer{})

	if err := cli.Connect("ws://localhost:8000"); err != nil {
		return nil, err
	}

	return cli, nil
}

func Test_Login(t *testing.T) {
	cli, err := login("test1")

	assert.Nil(t, err)

	time.Sleep(time.Second * 3)

	cli.Close()
}
