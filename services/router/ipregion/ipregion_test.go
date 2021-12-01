package ipregion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Ip2region_Search(t *testing.T) {
	ip, err := NewIp2region("../data/ip2region.db")
	assert.Nil(t, err)

	info, err := ip.Search("58.34.163.186")
	t.Log(info)
	assert.Nil(t, err)
	assert.Equal(t, "中国", info.Country)

	info, err = ip.Search("161.35.237.184")
	t.Log(info)
	assert.Nil(t, err)
	assert.Equal(t, "美国", info.Country)

	info, err = ip.Search("0.0.2222.1")
	t.Log(info)
	assert.Nil(t, err)
	assert.Equal(t, "0", info.Country)
}
