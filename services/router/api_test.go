package router

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"jinv/kim/services/router/apis"
)

func Test_Lookup(t *testing.T) {
	cli := resty.New()
	cli.SetHeader("Content-Type", "application/json")

	domains := make(map[string]int)

	for i := 0; i < 1000; i++ {
		url := fmt.Sprintf("http://localhost:8081/api/lookup/%s", ksuid.New().String())

		var res apis.LookUpResp
		resp, err := cli.R().SetResult(&res).Get(url)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		if len(res.Domains) > 0 {
			domain := res.Domains[0]

			domains[domain]++
		}
	}

	for domain, hit := range domains {
		fmt.Printf("domain: %s; hit hit: %d \n", domain, hit)
	}
}
