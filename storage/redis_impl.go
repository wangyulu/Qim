package storage

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/golang/protobuf/proto"
	"jinv/kim"
	"jinv/kim/wire/pkt"
)

const (
	LocationExpired = time.Hour * 48
)

type RedisStorage struct {
	cli *redis.Client
}

func NewRedisStorage(cli *redis.Client) kim.SessionStorage {
	return &RedisStorage{
		cli: cli,
	}
}

func (r *RedisStorage) Add(session *pkt.Session) error {
	// 1. 保存 Location
	loc := kim.Location{
		ChannelId: session.ChannelId,
		GateId:    session.GateId,
	}

	locKey := KeyLocation(session.Account, "")
	if err := r.cli.Set(locKey, loc.Bytes(), LocationExpired).Err(); err != nil {
		return err
	}

	// 2. 保存 Session
	snKey := KeySession(session.ChannelId)
	buf, _ := proto.Marshal(session)
	if err := r.cli.Set(snKey, buf, LocationExpired).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisStorage) Delete(account string, channelId string) error {
	locKey := KeyLocation(account, "")
	if err := r.cli.Del(locKey).Err(); err != nil {
		return err
	}

	snKey := KeySession(channelId)
	err := r.cli.Del(snKey).Err()

	return err
}

func (r *RedisStorage) Get(channelId string) (*pkt.Session, error) {
	snKey := KeySession(channelId)

	bts, err := r.cli.Get(snKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, kim.ErrSessionNil
		}

		return nil, err
	}

	var session pkt.Session

	err = proto.Unmarshal(bts, &session)

	return &session, err
}

func (r *RedisStorage) GetLocations(account ...string) ([]*kim.Location, error) {
	locKey := KeyLocations(account...)

	list, err := r.cli.MGet(locKey...).Result()
	if err != nil {
		return nil, err
	}

	result := make([]*kim.Location, 0)

	for _, l := range list {
		if l == nil {
			continue
		}

		var loc kim.Location
		_ = loc.Unmarshal([]byte(l.(string)))
		result = append(result, &loc)
	}

	if len(result) == 0 {
		return nil, kim.ErrSessionNil
	}

	return result, nil
}

func (r *RedisStorage) GetLocation(account string, device string) (*kim.Location, error) {
	locKey := KeyLocation(account, device)

	bts, err := r.cli.Get(locKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, kim.ErrSessionNil
		}

		return nil, err
	}

	var loc kim.Location

	err = loc.Unmarshal(bts)

	return &loc, err
}

func KeySession(channel string) string {
	return fmt.Sprintf("login:sn:%s", channel)
}

func KeyLocation(account, device string) string {
	if device == "" {
		return fmt.Sprintf("login:loc:%s", account)
	}

	return fmt.Sprintf("login:loc:%s:%s", account, device)
}

func KeyLocations(accounts ...string) []string {
	arr := make([]string, len(accounts))

	for i, account := range accounts {
		arr[i] = KeyLocation(account, "")
	}

	return arr
}
