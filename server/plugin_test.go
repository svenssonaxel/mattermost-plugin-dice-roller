package main

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoodRequestHelp(t *testing.T) {
	p, _ := initTestPlugin()
	assert.Nil(t, p.OnActivate())

	command := &model.CommandArgs{
		Command: "/roll help",
		UserId:  "userid",
	}

	response, err := p.ExecuteCommand(&plugin.Context{}, command)
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Nil(t, response.Attachments)
}

func TestPluginBadInputs(t *testing.T) {
	p, _ := initTestPlugin()
	assert.Nil(t, p.OnActivate())

	testCases := []string{
		"/lolzies d20",
		"/roll ",
		"/roll d0",
		"/roll hahaha",
		"/roll 6d",
		"/roll 0d5",
	}
	for _, testCase := range testCases {
		// Wrong dice requests
		command := &model.CommandArgs{
			Command: testCase,
		}
		response, err := p.ExecuteCommand(&plugin.Context{}, command)

		assert.NotNil(t, err)
		assert.Nil(t, response)
	}
}

func TestPluginGoodInputs(t *testing.T) {
	p, api := initTestPlugin()
	var post *model.Post
	api.On("CreatePost", mock.AnythingOfType("*model.Post")).Return(nil, nil).Run(func(args mock.Arguments) {
		post = args.Get(0).(*model.Post)
	})
	assert.Nil(t, p.OnActivate())

	testCases := []struct {
		inputDiceRequest string
		expectedText     string
	}{
		{inputDiceRequest: "1", expectedText: "**User** rolls 1 = **1**"},
		{inputDiceRequest: "2+3*5", expectedText: "**User** rolls 2+3×5 = **17**"},
		{inputDiceRequest: "(2+3)*5", expectedText: "**User** rolls (2+3)×5 = **25**"},
	}
	for _, testCase := range testCases {
		command := &model.CommandArgs{
			Command: "/roll " + testCase.inputDiceRequest,
			UserId:  "userid",
		}
		response, err := p.ExecuteCommand(&plugin.Context{}, command)
		testLabel := "Testing " + testCase.inputDiceRequest
		assert.Nil(t, err, testLabel)
		assert.NotNil(t, response, testLabel)
		assert.NotNil(t, post, testLabel)
		assert.NotNil(t, post.Message, testLabel)
		assert.Equal(t, testCase.expectedText, strings.TrimSpace(post.Message), testLabel)
	}
}

func initTestPlugin() (*Plugin, *plugintest.API) {
	api := &plugintest.API{}
	api.On("RegisterCommand", mock.Anything).Return(nil)
	api.On("UnregisterCommand", mock.Anything, mock.Anything).Return(nil)
	api.On("GetUser", mock.Anything).Return(&model.User{
		Id:       "userid",
		Nickname: "User",
	}, (*model.AppError)(nil))

	p := Plugin{
		configuration: &configuration{
			EnableDnd5e: true,
			EnableLatex: false,
		},
	}
	p.parser = GetParser(*p.configuration)
	p.SetAPI(api)

	return &p, api
}
