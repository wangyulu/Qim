package container

import (
	"sync"

	"jinv/kim"
	"jinv/kim/logger"
)

type ClientMap interface {
	Add(client kim.Client)
	Remove(id string)
	Get(id string) (client kim.Client, ok bool)
	Services(kvs ...string) []kim.Service
}

type ClientsImpl struct {
	clients *sync.Map
}

func NewClients(num int) ClientMap {
	return &ClientsImpl{
		clients: new(sync.Map),
	}
}

func (ch *ClientsImpl) Add(client kim.Client) {
	if client.ServiceID() == "" {
		logger.WithFields(logger.Fields{"module": "ClientImpl"}).Error("client id is required")
	}

	ch.clients.Store(client.ServiceID(), client)
}

func (ch *ClientsImpl) Remove(id string) {
	ch.clients.Delete(id)
}

func (ch *ClientsImpl) Get(id string) (client kim.Client, ok bool) {
	if id == "" {
		logger.WithFields(logger.Fields{"module": "ClientImpl"}).Error("client id is required")
	}

	v, ok := ch.clients.Load(id)
	if !ok {
		return nil, false
	}

	return v.(kim.Client), true
}

func (ch *ClientsImpl) Services(kvs ...string) []kim.Service {
	kvLen := len(kvs)
	if kvLen != 0 && kvLen != 2 {
		return nil
	}

	arr := make([]kim.Service, 0)

	ch.clients.Range(func(key, value interface{}) bool {
		srv := value.(kim.Service)
		if kvLen > 0 && srv.GetMeta()[kvs[0]] != kvs[1] {
			return true
		}

		arr = append(arr, srv)

		return true
	})

	return arr
}
