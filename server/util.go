package main

import (
	"github.com/moussetc/mattermost-plugin-dice-roller/server/br"
	"github.com/moussetc/mattermost-plugin-dice-roller/server/pd"
)

// Useful short utilities from br and pd packages

type BR = br.BR
type PD = pd.PD

func itobr(i int) BR {
	return br.New(i)
}

var (
	zero     = br.Zero
	one      = br.One
	nine     = br.New(9)
	nineteen = br.New(19)
	twenty   = br.New(20)

	zeroPD = pd.ZeroPD
	onePD  = pd.OnePD
	errPD  = pd.ErrPD
)
