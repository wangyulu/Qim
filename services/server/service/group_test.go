package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"jinv/kim/wire/rpc"
)

const app = "kim_test"

var groupService = NewGroupService("http://localhost:8080")

func Test_GroupService(t *testing.T) {
	resp, err := groupService.Create(app, &rpc.CreateGroupReq{
		Name:    "test",
		Owner:   "test1",
		Members: []string{"test1", "test2"},
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, resp.GroupId)
	t.Log(resp.GroupId)

	respMembers1, err := groupService.Members(app, &rpc.GroupMembersReq{
		GroupId: resp.GroupId,
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(respMembers1.GetUsers()))

	err = groupService.Join(app, &rpc.JoinGroupReq{
		Account: "test3",
		GroupId: resp.GroupId,
	})
	assert.Nil(t, err)

	respMembers2, err := groupService.Members(app, &rpc.GroupMembersReq{
		GroupId: resp.GroupId,
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(respMembers2.GetUsers()))
	assert.Equal(t, "test3", respMembers2.GetUsers()[2].Account)

	err = groupService.Quit(app, &rpc.QuitGroupReq{
		Account: "test2",
		GroupId: resp.GroupId,
	})
	assert.Nil(t, err)

	respMembers3, err := groupService.Members(app, &rpc.GroupMembersReq{
		GroupId: resp.GroupId,
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(respMembers3.GetUsers()))
	assert.Equal(t, "test3", respMembers3.GetUsers()[1].Account)
}
