package storage

import (
	"github.com/go-redis/redis/v7"
	"github.com/golang/protobuf/proto"
	"jinv/kim"
	"jinv/kim/wire/pkt"
)

type RedisClusterStorageV2 struct {
	cli *redis.ClusterClient
}

func NewRedisClusterStorageV2(cli *redis.ClusterClient) kim.SessionStorage {
	return &RedisClusterStorageV2{
		cli: cli,
	}
}

func (r *RedisClusterStorageV2) Add(session *pkt.Session) error {
	loc := &kim.Location{
		ChannelId: session.ChannelId,
		GateId:    session.GateId,
	}

	locKey := KeyLocation(session.Account, "")
	err := r.cli.Do("SET", locKey, loc.Bytes(), "EX", int(LocationExpired)).Err()
	if err != nil {
		return err
	}

	snKey := KeySession(session.ChannelId)
	buf, _ := proto.Marshal(session)
	err = r.cli.Do("SET", snKey, buf, "EX", int(LocationExpired)).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClusterStorageV2) Delete(account string, channelId string) error {
	locKey := KeyLocation(account, "")
	err := r.cli.Do("DEL", locKey).Err()
	if err != nil {
		return err
	}

	snKey := KeySession(channelId)
	err = r.cli.Do("DEL", snKey).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClusterStorageV2) Get(ChannelId string) (*pkt.Session, error) {
	snKey := KeySession(ChannelId)
	bts, err := r.cli.Get(snKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, kim.ErrSessionNil
		}

		return nil, err
	}

	var session pkt.Session

	_ = proto.Unmarshal(bts, &session)

	return &session, nil
}

func (r *RedisClusterStorageV2) GetLocations(accounts ...string) ([]*kim.Location, error) {
	locKeys := KeyLocations(accounts...)

	pipe := r.cli.Pipeline()

	cmds := make([]*redis.StringCmd, len(locKeys))

	for i, val := range locKeys {
		cmds[i] = pipe.Get(val)
	}

	_, err := pipe.Exec()
	if err != nil {
		if err == redis.Nil {
			return nil, kim.ErrSessionNil
		}

		return nil, err
	}

	result := make([]*kim.Location, 0)

	for _, l := range cmds {
		if l == nil {
			continue
		}

		var loc kim.Location
		_ = loc.Unmarshal([]byte(l.Val()))
		result = append(result, &loc)
	}

	if len(result) == 0 {
		return nil, kim.ErrSessionNil
	}

	return result, nil
}

func (r *RedisClusterStorageV2) GetLocation(account string, device string) (*kim.Location, error) {
	locKey := KeyLocation(account, device)
	bts, err := r.cli.Get(locKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, kim.ErrSessionNil
		}

		return nil, err
	}

	var loc kim.Location

	_ = loc.Unmarshal(bts)

	return &loc, nil
}
