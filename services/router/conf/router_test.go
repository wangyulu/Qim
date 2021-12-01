package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LoadMappings(t *testing.T) {
	mapping, err := LoadMapping("../data/mapping.json")
	assert.Nil(t, err)

	t.Log(mapping)

	assert.Equal(t, "EC", mapping["中国"])

	assert.Equal(t, "JP", mapping["日本"])
	assert.Equal(t, "JP", mapping["韩国"])
}

func Test_LoadRegions(t *testing.T) {
	regions, err := LoadRegions("../data/regions.json")
	assert.Nil(t, err)

	t.Log(regions["EC"])
	t.Log(regions["JP"])

	assert.Equal(t, "SH_ALI", regions["EC"].Idcs[0].ID)
}
