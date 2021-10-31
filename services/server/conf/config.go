package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"jinv/kim"
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
	RoyalURL      string   `envconfig:"royalURL"`
	LogLevel      string   `envconfig:"logLevel",default:"INFO"`
}

func (c Config) Stirng() string {
	bts, _ := json.Marshal(c)

	return string(bts)
}

func Init(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/conf")

	var config Config
	if err := viper.ReadInConfig(); err != nil {
		logger.Warn(err)
	} else {
		if err := viper.Unmarshal(&config); err != nil {
			return nil, err
		}
	}

	if err := envconfig.Process("kim", &config); err != nil {
		return nil, err
	}

	if config.ServiceID == "" {
		localIP := kim.GetLocalIP()
		config.ServiceID = fmt.Sprintf("server_%s", strings.ReplaceAll(localIP, ".", ""))
	}

	if config.PublicAddress == "" {
		config.PublicAddress = kim.GetLocalIP()
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
