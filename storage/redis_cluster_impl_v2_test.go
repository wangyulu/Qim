package storage

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/stretchr/testify/assert"
	"jinv/kim"
	"jinv/kim/wire/pkt"
)

func Test_RedisClusterStoreV2_Add(t *testing.T) {
	cluster := redis.NewClusterClient(
		&redis.ClusterOptions{
			Addrs:        []string{"127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7003"},
			DialTimeout:  time.Second * 5,
			ReadTimeout:  time.Second * 5,
			WriteTimeout: time.Second * 5,
		},
	)

	_, err := cluster.Ping().Result()
	assert.Nil(t, err)

	storage := NewRedisClusterStorageV2(cluster)

	err = storage.Add(&pkt.Session{
		ChannelId: "ch1",
		GateId:    "gateway01",
		Account:   "test1",
		Device:    "phone",
	})
	assert.Nil(t, err)

	err = storage.Add(&pkt.Session{
		ChannelId: "ch2",
		GateId:    "gateway01",
		Account:   "test2",
		Device:    "pc",
	})
	assert.Nil(t, err)

	session, err := storage.Get("ch1")
	assert.Nil(t, err)
	t.Log(session)
	assert.Equal(t, "ch1", session.ChannelId)
	assert.Equal(t, "test1", session.Account)

	locs, err := storage.GetLocations("test1", "test2")
	assert.Nil(t, err)
	t.Logf("%v", locs)

	loc := locs[1]
	assert.Equal(t, "ch2", loc.ChannelId)
	assert.Equal(t, "gateway01", loc.GateId)

	locs, err = storage.GetLocations("test5")
	assert.Equal(t, kim.ErrSessionNil, err)
	assert.Equal(t, 0, len(locs))
}
