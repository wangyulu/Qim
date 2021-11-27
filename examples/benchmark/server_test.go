package benchmark

import (
	"fmt"
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
