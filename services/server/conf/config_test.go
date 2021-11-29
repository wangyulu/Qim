package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	config, err := Init("../conf.yaml")

	t.Log(config.RedisClusterAddrs)

	assert.Nil(t, err)

	assert.Equal(t, 3, len(config.RedisClusterAddrs))
}
