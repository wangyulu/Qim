package apis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"jinv/kim"
	"jinv/kim/naming"
	"jinv/kim/services/router/conf"
)

func Test_SelectIdc(t *testing.T) {
	idc := selectIdc("test1", &conf.Region{
		Idcs: []conf.IDC{
			{ID: "SH_ALI"},
			{ID: "HZ_ALI"},
			{ID: "SH_TENCENT"},
		},
		Slots: []byte{0, 0, 1, 1, 2, 2, 2},
	})

	t.Log(idc)
}

func Test_SelectGateways(t *testing.T) {
	gateways := selectGateways("test1", []kim.ServiceRegistration{
		&naming.DefaultService{Id: "g1"},
		&naming.DefaultService{Id: "g2"},
	}, 3)
	assert.Equal(t, 2, len(gateways))

	gateways = selectGateways("test1", []kim.ServiceRegistration{
		&naming.DefaultService{Id: "g1"},
		&naming.DefaultService{Id: "g2"},
		&naming.DefaultService{Id: "g3"},
	}, 3)
	assert.Equal(t, 3, len(gateways))

	gateways = selectGateways("test2", []kim.ServiceRegistration{
		&naming.DefaultService{Id: "g1"},
		&naming.DefaultService{Id: "g2"},
		&naming.DefaultService{Id: "g3"},
		&naming.DefaultService{Id: "g4"},
		&naming.DefaultService{Id: "g5"},
	}, 3)

	t.Log(gateways)

	assert.Equal(t, 3, len(gateways))
}
