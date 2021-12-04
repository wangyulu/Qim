package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ReadRoute(t *testing.T) {
	r, err := ReadRoute("../route.json")
	assert.Nil(t, err)

	t.Log(r)

	assert.Equal(t, "app", r.RouteBy)
	assert.Equal(t, "zone_ali_03", r.WhiteList["kim"])
}
