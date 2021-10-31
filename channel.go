package kim

import (
	"errors"
	"sync"
	"time"

	"jinv/kim/logger"
)

type ChannelImpl struct {
	sync.Mutex
	once sync.Once

	id string

	Conn

	writeChan chan []byte

	writeWait time.Duration
	readWait  time.Duration

	closed *Event
}

func NewChannel(id string, conn Conn) Channel {
	log := logger.WithFields(logger.Fields{
		"module": "tcp_channel",
		"id":     id,
	})

	ch := &ChannelImpl{
		id:        id,
		Conn:      conn,
		writeChan: make(chan []byte, 5),
		writeWait: DefaultWriteWait,
		readWait:  DefaultReadWait,
		closed:    NewEvent(),
	}

	go func() {
		err := ch.writeLoop()
		if err != nil {
			log.Info(err)
		}
	}()

	return ch
}

func (ch *ChannelImpl) writeLoop() error {
	for {
		select {
		case payload := <-ch.writeChan:
			err := ch.WriteFrame(OpBinary, payload)
			if err != nil {
				return err
			}

			// todo 批量写 why
			chanLen := len(ch.writeChan)
			for i := 0; i < chanLen; i++ {
				payload := <-ch.writeChan
				err := ch.WriteFrame(OpBinary, payload)
				if err != nil {
					return err
				}
			}

			err = ch.Conn.Flush()
			if err != nil {
				return err
			}
		case <-ch.closed.Done():
			return nil
		}
	}
}

func (ch *ChannelImpl) WriteFrame(code OpCode, payload []byte) error {
	if err := ch.Conn.SetWriteDeadline(time.Now().Add(ch.writeWait)); err != nil {
		return err
	}

	return ch.Conn.WriteFrame(code, payload)
}

func (ch *ChannelImpl) ReadLoop(lst MessageListener) error {
	// todo 这是加锁的目的是什么
	ch.Lock()
	defer ch.Unlock()

	log := logger.WithFields(logger.Fields{
		"struct": "ChannelImpl",
		"func":   "ReadLoop",
		"id":     ch.id,
	})

	for {
		if err := ch.SetReadDeadline(time.Now().Add(ch.readWait)); err != nil {
			return err
		}

		frame, err := ch.ReadFrame()
		if err != nil {
			return err
		}

		if frame.GetOpCode() == OpClose {
			return errors.New("remote side close the channel")
		}

		if frame.GetOpCode() == OpPing {
			log.Trace("recv a ping; resp with a pong")
			_ = ch.WriteFrame(OpPong, nil)
			continue
		}

		payload := frame.GetPayload()
		if len(payload) == 0 {
			continue
		}

		go lst.Receive(ch, payload)
	}
}

func (ch *ChannelImpl) ID() string {
	return ch.id
}

func (ch *ChannelImpl) Push(payload []byte) error {
	if ch.closed.HasFired() {
		return errors.New("channel has closed")
	}

	// 异步写
	ch.writeChan <- payload

	return nil
}

func (ch *ChannelImpl) SetWriteWait(writeWait time.Duration) {
	if writeWait == 0 {
		return
	}

	ch.writeWait = writeWait
}

func (ch *ChannelImpl) SetReadWait(readWait time.Duration) {
	if readWait == 0 {
		return
	}

	ch.readWait = readWait
}

func (ch *ChannelImpl) Close() error {
	ch.once.Do(func() {
		ch.closed.Fire()

		close(ch.writeChan)
	})

	return nil
}
