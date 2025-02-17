package main

import (
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost/server/public/model"
)

// configuration captures the plugin's external configuration as exposed in the Mattermost server
// configuration, as well as values computed from the configuration. Any public fields will be
// deserialized from the Mattermost server configuration in OnConfigurationChange.
//
// As plugins are inherently concurrent (hooks being called asynchronously), and the plugin
// configuration can change at any time, access to the configuration must be synchronized. The
// strategy used in this plugin is to guard a pointer to the configuration, and clone the entire
// struct whenever it changes. You may replace this with whatever strategy you choose.
//
// If you add non-reference types to your configuration struct, be sure to rewrite Clone as a deep
// copy appropriate for your types.
type configuration struct {
	EnableDnd5e bool `json:"enable_dnd5e"`
	EnableLatex bool `json:"enable_latex"`
}

// Clone shallow copies the configuration. Your implementation may require a deep copy if
// your configuration has reference types.
func (c *configuration) Clone() *configuration {
	var clone = *c
	return &clone
}

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{
			EnableDnd5e: true,
			EnableLatex: true,
		}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
//
// Do not call setConfiguration while holding the configurationLock, as sync.Mutex is not
// reentrant. In particular, avoid using the plugin API entirely, as this may in turn trigger a
// hook back into the plugin. If that hook attempts to acquire this lock, a deadlock may occur.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		// Ignore assignment if the configuration struct is empty. Go will optimize the
		// allocation for same to point at the same memory address, breaking the check
		// above.
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(configuration)

	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	p.setConfiguration(configuration)
	p.parser = GetParser(*configuration)

	return p.defineBot()
}

func (p *Plugin) defineBot() error {
	bot := model.Bot{
		Username:    "dicerollerbot",
		DisplayName: "Dice Roller",
		Description: "A bot account created by " + manifest.Name + " plugin.",
	}
	botID, ensureBotError := p.API.EnsureBotUser(&bot)
	if ensureBotError != nil {
		return errors.Wrap(ensureBotError, "failed to ensure dice bot")
	}

	p.diceBotID = botID

	// Set ../assets/icon.png as profile image for the bot account
	bundlePath, bpErr := p.API.GetBundlePath()
	if bpErr != nil {
		return bpErr
	}

	iconData, ioErr := ioutil.ReadFile(filepath.Join(bundlePath, "assets", "icon.png"))
	if ioErr != nil {
		return ioErr
	}

	appErr := p.API.SetProfileImage(botID, iconData)
	if appErr != nil {
		return appErr
	}

	return nil
}
