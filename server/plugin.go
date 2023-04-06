package main

import (
	_ "embed"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
)

const (
	trigger string = "roll"
)

//go:embed helptext.md
var helpText string

//go:embed helptext-dnd5e.md
var helpTextDnd5e string

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	// function used to parse dice rolls
	parser func(input string) (*Node, error)

	// BotId of the created bot account for dice rolling
	diceBotID string
}

func (p *Plugin) OnActivate() error {
	rand.Seed(time.Now().UnixNano())

	return p.API.RegisterCommand(&model.Command{
		Trigger:          trigger,
		Description:      "Roll one or more dice",
		DisplayName:      "Dice roller ⚄",
		AutoComplete:     true,
		AutoCompleteDesc: "Roll one or several dice. ⚁ ⚄ Try /roll help for a list of possibilities.",
		AutoCompleteHint: "(3d20+4)/2",
	})
}

func (p *Plugin) GetHelpMessage() *model.CommandResponse {
	text := helpText
	if p.getConfiguration().EnableDnd5e {
		text += helpTextDnd5e
	}
	text += "⚅ ⚂ Let's get rolling! ⚁ ⚄"

	props := map[string]interface{}{
		"from_webhook": "true",
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         text,
		Props:        props,
	}
}

// ExecuteCommand returns a post that displays the result of the dice rolls
func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	if p.API == nil {
		return nil, appError("Cannot access the plugin API.", nil)
	}

	cmd := "/" + trigger
	if strings.HasPrefix(args.Command, cmd) {
		query := strings.TrimSpace((strings.Replace(args.Command, cmd, "", 1)))

		lQuery := strings.ToLower(query)
		if lQuery == "help" || lQuery == "--help" || lQuery == "h" || lQuery == "-h" {
			return p.GetHelpMessage(), nil
		}

		// Suppress lint error
		// > G404: Use of weak random number generator (math/rand instead of crypto/rand) (gosec)
		// because dice rolls don't need to be cryptographically secure.
		//#nosec G404
		roller := func(x int) int { return 1 + rand.Intn(x) }
		post, generatePostError := p.generateDicePost(query, args.UserId, args.ChannelId, args.RootId, roller, p.parser)
		if generatePostError != nil {
			return nil, generatePostError
		}
		_, createPostError := p.API.CreatePost(post)
		if createPostError != nil {
			return nil, createPostError
		}

		return &model.CommandResponse{}, nil
	}

	return nil, appError("Expected trigger "+cmd+" but got "+args.Command, nil)
}

func (p *Plugin) generateDicePost(query, userID, channelID, rootID string, roller Roller, parse func(input string) (*Node, error)) (*model.Post, *model.AppError) {
	// Get the user to display their name
	user, userErr := p.API.GetUser(userID)
	if userErr != nil {
		return nil, userErr
	}
	displayName := user.Nickname
	if displayName == "" {
		displayName = user.Username
	}

	parsedNode, err := parse(query)
	if err != nil {
		return nil, appError(fmt.Sprintf("%s: See `/roll help` for examples.", err.Error()), err)
	}

	rolledNode := parsedNode.roll(roller, *p.configuration)
	renderResult := rolledNode.renderToplevel()

	text := fmt.Sprintf("**%s** rolls %s", displayName, renderResult)

	return &model.Post{
		UserId:    p.diceBotID,
		ChannelId: channelID,
		RootId:    rootID,
		Message:   text,
	}, nil
}

func appError(message string, err error) *model.AppError {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	return model.NewAppError("Dice Roller Plugin", message, nil, errorMessage, http.StatusBadRequest)
}
