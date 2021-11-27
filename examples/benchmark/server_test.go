package benchmark

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
	"jinv/kim"
	"jinv/kim/examples/mock"
	"jinv/kim/logger"
	"jinv/kim/websocket"
)

const wsurl = "ws://localhost:9001"

func Test_Parallel(t *testing.T) {
	const count = 10000

	pool, _ := ants.NewPool(50, ants.WithPreAlloc(true))
	defer pool.Release()

	var wg sync.WaitGroup
	wg.Add(count)

	clis := make([]kim.Client, count)

	t0 := time.Now()

	for i := 0; i < count; i++ {
		idx := i

		_ = pool.Submit(func() {
			cli := websocket.NewClient(fmt.Sprintf("test_%v", idx), "client", websocket.ClientOptions{
				Heartbeat: kim.DefaultHeartbeat,
			})

			cli.SetDialer(&mock.WebsocketDialer{})

			err := cli.Connect(wsurl)
			if err != nil {
				logger.Error(err)
			}

			clis[idx] = cli
			wg.Done()
		})
	}

	wg.Wait()

	t.Logf("logined %d cost %v", count, time.Since(t0))
	t.Logf("done connecting..")

	time.Sleep(time.Second * 5)

	for i := 0; i < count; i++ {
		clis[i].Close()
	}

	t.Logf("closed")
}

func Test_Message(t *testing.T) {
	const count = 1000 * 100

	cli := websocket.NewClient(fmt.Sprintf("test_%v", 1), "client", websocket.ClientOptions{
		Heartbeat: kim.DefaultHeartbeat,
	})

	cli.SetDialer(&mock.WebsocketDialer{})

	err := cli.Connect(wsurl)
	if err != nil {
		logger.Error(err)
	}

	// 消息内容50个字符，不可超过缓冲区大小（不然会绕过缓冲区）
	msg := []byte(strings.Repeat("hello", 10))

	t0 := time.Now()

	go func() {
		for i := 0; i < count; i++ {
			_ = cli.Send(msg)
		}
	}()

	recv := 0

	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info("time", time.Now().UnixNano(), err)
			break
		}

		if frame.GetOpCode() != kim.OpBinary {
			continue
		}

		recv++

		if recv == count {
			break
		}
	}

	t.Logf("message %d cost %v", count, time.Since(t0))
}
