package storage

import (
	"github.com/gitstliu/go-redis-cluster"
	"github.com/golang/protobuf/proto"
	"jinv/kim"
	"jinv/kim/wire/pkt"
)

type RedisClusterStorage struct {
	cli *redis.Cluster
}

func NewRedisClusterStorage(cli *redis.Cluster) kim.SessionStorage {
	return &RedisClusterStorage{
		cli: cli,
	}
}

func (r *RedisClusterStorage) Add(session *pkt.Session) error {
	loc := &kim.Location{
		ChannelId: session.ChannelId,
		GateId:    session.GateId,
	}

	locKey := KeyLocation(session.Account, "")
	_, err := r.cli.Do("SET", locKey, loc.Bytes(), "EX", int(LocationExpired))
	if err != nil {
		return err
	}

	snKey := KeySession(session.ChannelId)
	buf, _ := proto.Marshal(session)
	_, err = r.cli.Do("SET", snKey, buf, "EX", int(LocationExpired))
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClusterStorage) Delete(account string, channelId string) error {
	locKey := KeyLocation(account, "")
	_, err := r.cli.Do("DEL", locKey)
	if err != nil {
		return err
	}

	snKey := KeySession(channelId)
	_, err = r.cli.Do("DEL", snKey)
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClusterStorage) Get(ChannelId string) (*pkt.Session, error) {
	snKey := KeySession(ChannelId)
	bts, err := redis.Bytes(r.cli.Do("GET", snKey))
	if err != nil {
		if err == redis.ErrNil {
			return nil, kim.ErrSessionNil
		}

		return nil, err
	}

	var session pkt.Session

	_ = proto.Unmarshal(bts, &session)

	return &session, nil
}

func (r *RedisClusterStorage) GetLocations(accounts ...string) ([]*kim.Location, error) {
	locKeys := KeyLocationsBytes(accounts...)

	list, err := redis.Values(r.cli.Do("MGET", locKeys...))
	if err != nil {
		return nil, err
	}

	result := make([]*kim.Location, 0)

	for _, l := range list {
		if l == nil {
			continue
		}

		var loc kim.Location
		_ = loc.Unmarshal(l.([]byte))
		result = append(result, &loc)
	}

	if len(result) == 0 {
		return nil, kim.ErrSessionNil
	}

	return result, nil
}

func (r *RedisClusterStorage) GetLocation(account string, device string) (*kim.Location, error) {
	locKey := KeyLocation(account, device)
	bts, err := redis.Bytes(r.cli.Do("GET", locKey))
	if err != nil {
		if err == redis.ErrNil {
			return nil, kim.ErrSessionNil
		}

		return nil, err
	}

	var loc kim.Location

	_ = loc.Unmarshal(bts)

	return &loc, nil
}
