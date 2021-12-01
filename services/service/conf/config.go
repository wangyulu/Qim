package conf

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/kataras/iris/v12/middleware/accesslog"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
	"jinv/kim"
	"jinv/kim/logger"
)

type Config struct {
	ServiceID     string
	NodeID        int64
	Listen        string
	PublicAddress string
	PublicPort    int `default:"8080"`
	Tags          []string
	ConsulURL     string
	RedisAddrs    string
	BaseDb        string
	MessageDb     string
	LogLevel      string `default:"INFO"`
}

func (c *Config) String() string {
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
		config.ServiceID = fmt.Sprintf("royal_%s", strings.ReplaceAll(localIP, ".", ""))

		localIPArr := strings.Split(localIP, ".")
		if len(localIPArr) == 4 {
			suffix, _ := strconv.Atoi(localIPArr[3])
			config.NodeID = int64(suffix)
		}
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

func MakeAccessLog() *accesslog.AccessLog {
	// Initialize a new access log middleware.
	ac := accesslog.File("./data/access.log")
	// Remove this line to disable logging to console:
	ac.AddOutput(os.Stdout)

	// The default configuration:
	ac.Delim = '|'
	ac.TimeFormat = "2006-01-02 15:04:05"
	ac.Async = false
	ac.IP = true
	ac.BytesReceivedBody = true
	ac.BytesSentBody = true
	ac.BytesReceived = false
	ac.BytesSent = false
	ac.BodyMinify = true
	ac.RequestBody = true
	ac.ResponseBody = false
	ac.KeepMultiLineError = true
	ac.PanicLog = accesslog.LogHandler

	// Default line format if formatter is missing:
	// Time|Latency|Code|Method|Path|IP|Path Params Query Fields|Bytes Received|Bytes Sent|Request|Response|
	//
	// Set Custom Formatter:
	// ac.SetFormatter(&accesslog.JSON{
	// 	Indent:    "  ",
	// 	HumanTime: true,
	// })
	// ac.SetFormatter(&accesslog.CSV{})
	// ac.SetFormatter(&accesslog.Template{Text: "{{.Code}}"})

	return ac
}
