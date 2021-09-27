package kim

import (
	"sync"

	"jinv/kim/logger"
)

type ChannelMap interface {
	Add(channel Channel)
	Remove(id string)
	Get(id string) (Channel, bool)
	All() []Channel
}

type ChannelsImpl struct {
	channels *sync.Map
}

func NewChannels(num int) ChannelMap {
	return &ChannelsImpl{
		channels: new(sync.Map),
	}
}

func (c *ChannelsImpl) Add(channel Channel) {
	if channel.ID() == "" {
		logger.WithFields(logger.Fields{
			"module": "ChannelsImpl",
		}).Error("channel id is required")
	}

	c.channels.Store(channel.ID(), channel)

	return
}

func (c *ChannelsImpl) Remove(id string) {
	c.channels.Delete(id)

	return
}

func (c *ChannelsImpl) Get(id string) (Channel, bool) {
	if id == "" {
		logger.WithFields(logger.Fields{
			"module": "ChannelsImpl",
		}).Error("channel id is required")
	}

	channel, ok := c.channels.Load(id)

	if !ok {
		return nil, false
	}

	return channel.(Channel), true
}

func (c *ChannelsImpl) All() []Channel {
	arr := make([]Channel, 0)

	c.channels.Range(func(key, value interface{}) bool {
		arr = append(arr, value.(Channel))

		return true
	})

	return arr
}
