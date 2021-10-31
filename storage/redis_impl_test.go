package storage

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"jinv/kim"
	"jinv/kim/wire/pkt"
)

const addr = "localhost:6379"

func InitRedis(addr string, pass string) (*redis.Client, error) {
	redisDb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     pass,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
	})

	if _, err := redisDb.Ping().Result(); err != nil {
		log.Println(err)
		return nil, err
	}

	return redisDb, nil
}

func Test_CRUD(t *testing.T) {
	cli, err := InitRedis(addr, "")
	assert.Nil(t, err)

	store := NewRedisStorage(cli)
	err = store.Add(&pkt.Session{
		ChannelId: "ch1",
		GateId:    "gateway1",
		Account:   "test1",
		Device:    "Phone",
	})
	assert.Nil(t, err)

	err = store.Add(&pkt.Session{
		ChannelId: "ch2",
		GateId:    "gateway1",
		Account:   "test2",
		Device:    "Pc",
	})
	assert.Nil(t, err)

	session, err := store.Get("ch1")
	assert.Nil(t, err)
	assert.Equal(t, "gateway1", session.GateId)
	assert.Equal(t, "test1", session.Account)

	locs1, err := store.GetLocations("test1", "test2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(locs1))
	assert.Equal(t, "ch2", locs1[1].ChannelId)

	locs2, err := store.GetLocations("test6")
	assert.Equal(t, kim.ErrSessionNil, err)
	assert.Equal(t, 0, len(locs2))
}

func Benchmark_MGET(b *testing.B) {
	cli, err := InitRedis(addr, "")
	assert.Nil(b, err)

	store := NewRedisStorage(cli)
	count := 100
	accounts := make([]string, count)

	for i := 0; i < 100; i++ {
		accounts[i] = fmt.Sprintf("account_%d", i)
		err := store.Add(&pkt.Session{
			ChannelId: ksuid.New().String(),
			GateId:    "gateway1",
			Account:   accounts[i],
		})
		assert.Nil(b, err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	// todo
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := store.GetLocations(accounts...)
			assert.Nil(b, err)
		}
	})
}

func Benchmark_GetLocation(b *testing.B) {
	cli, err := InitRedis(addr, "")
	assert.Nil(b, err)

	store := NewRedisStorage(cli)
	count := 100
	accounts := make([]string, count)

	for i := 0; i < count; i++ {
		accounts[i] = ksuid.New().String()
		err := store.Add(&pkt.Session{
			ChannelId: ksuid.New().String(),
			GateId:    fmt.Sprintf("%s_gateway1", ksuid.New().String()),
			Account:   accounts[i],
			Zone:      ksuid.New().String(),
			Isp:       ksuid.New().String(),
			RemoteIP:  "121.232.122.121",
			App:       "kim",
			Tags:      []string{"tag1", "tag2", "tag3"},
		})
		assert.Nil(b, err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			account := accounts[rand.Intn(100)]
			_, err := store.GetLocation(account, "")
			assert.Nil(b, err)
		}
	})
}

func Benchmark_GetSession(b *testing.B) {
	cli, err := InitRedis(addr, "")
	assert.Nil(b, err)

	store := NewRedisStorage(cli)
	count := 100
	channelIds := make([]string, count)

	for i := 0; i < count; i++ {
		channelIds[i] = ksuid.New().String()
		err := store.Add(&pkt.Session{
			ChannelId: channelIds[i],
			GateId:    fmt.Sprintf("%s_gateway1", ksuid.New().String()),
			Account:   ksuid.New().String(),
			Zone:      ksuid.New().String(),
			Isp:       ksuid.New().String(),
			RemoteIP:  "121.232.122.121",
			App:       "kim",
			Tags:      []string{"tag1", "tag2", "tag3"},
		})
		assert.Nil(b, err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			channelId := channelIds[rand.Intn(100)]
			_, err := store.Get(channelId)
			assert.Nil(b, err)
		}
	})
}
