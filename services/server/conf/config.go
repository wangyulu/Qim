package conf

import (
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"jinv/kim/logger"
)

type Config struct {
	ServiceID     string   `envconfig:"serviceId"`
	Namespace     string   `envconfig:"namespace"`
	Listen        string   `envconfig:"listen"`
	PublicAddress string   `envconfig:"publicAddress"`
	PublicPort    int      `envconfig:"publicPort"`
	Tags          []string `evnconfig:"tags"`
	ConsulURL     string   `envconfig:"consulURL"`
	RedisAddrs    string   `envconfig:"redisAddrs"`
	RpcURL        string   `envconfig:"rpcURL"`
}

func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config file not found: %s", err)
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	if err := envconfig.Process("", &config); err != nil {
		return nil, err
	}

	logger.Info(config)

	return &config, nil
}

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
