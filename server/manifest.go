// This file is automatically generated. Do not modify it manually.

package main

import (
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
)

var manifest *model.Manifest

const manifestStr = `
{
  "id": "com.github.moussetc.mattermost.plugin.diceroller",
  "name": "Dice Roller ⚄",
  "description": "Add a command for rolling dice ⚄",
  "version": "3.0.0",
  "min_server_version": "5.12.0",
  "server": {
    "executables": {
      "linux-amd64": "server/dist/plugin-linux-amd64",
      "darwin-amd64": "server/dist/plugin-darwin-amd64",
      "windows-amd64": "server/dist/plugin-windows-amd64.exe"
    },
    "executable": "server/dist/plugin-freebsd-amd64"
  },
  "settings_schema": {
    "header": "",
    "footer": "* To report an issue, make a suggestion or a contribution, [check the GitHub repository]i(https://github.com/moussetc/mattermost-plugin-spoiler/)",
    "settings": []
  }
}
`

func init() {
	manifest = model.ManifestFromJson(strings.NewReader(manifestStr))
}