package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/moussetc/mattermost-plugin-dice-roller/server/pd"
)

// Types
type Node struct {
	token       string
	child       []Node
	sp          NodeSpecialization
	rollComment string
}
type NodeSpecialization interface {
	roll(Node, Roller) NodeSpecialization
	// The roll comment can be
	// - ROLL_COMMENT_BLOCK_PARENT to indicate that no sum or product
	//   (grand)parent node is allowed to render a roll comment.
	// - ROLL_COMMENT_NOTHING to indicate that the node has no roll comment.
	// - The roll comment as a string, which is only to be used in the case
	//   that no (grand)parent has one.
	rollComment(n Node, c configuration) string
	value(Node) BR
	// Arguments:
	// 1. The node to render.
	// 2. The indentation prefix, e.g. "- " for the top level.
	// 3. The result role, one of RR_NONE, RR_TOP, RR_DETAIL.
	//	  (How the second return value will be used.)
	// 4. Whether roll comment is allowed to be rendered.
	// 5. The options string, which can be "l" for allowing inline latex, or "".
	// It returns three strings:
	// 1. An unformatted expression, used as (part of) the left side of the equals sign.
	// 2. A formatted result potentially used as (part of) the right side of the
	//    equals sign, or the empty string if there should be no equals sign.
	// 3. The details list, i.e. all subsequent rows, formatted and including the
	//    indentation prefix. The details list either starts with a newline or is the
	//    empty string.
	render(n Node, ind string, rr int, rcok bool, options string) (string, string, string)
	prob(n Node) PD
}
type Roller func(int) int
type GroupExpr struct{}
type Natural struct{ n int }
type Sum struct{ ops []string }
type Prod struct{ ops []string }
type Dice struct {
	n     int          // number of dice
	x     int          // number of sides
	l     int          // index in sorted results for first dice to keep, e.g. 0 to keep all
	h     int          // index in sorted results for first dice after the last to keep, e.g. n to keep all
	rolls []RollResult // roll results
}
type RollResult struct {
	result int
	use    bool
	order  int // order rolled
	rank   int // index when sorted by (result, order)
}
type Stats struct{}
type DeathSave struct{}
type Labeled struct {
	label string
}
type CommaList struct{}

// Constants
const (
	RR_NONE = iota + 1
	RR_TOP
	RR_DETAIL
)

// Roller
func (n Node) roll(roller Roller, conf configuration) Node {
	child := make([]Node, len(n.child))
	for i, c := range n.child {
		child[i] = c.roll(roller, conf)
	}
	sp := n.sp.roll(n, roller)
	node := Node{token: n.token, child: child, sp: sp}
	node.rollComment = sp.rollComment(node, conf)
	return node
}
func (sp GroupExpr) roll(_ Node, _ Roller) NodeSpecialization { return sp }
func (sp Natural) roll(_ Node, _ Roller) NodeSpecialization   { return sp }
func (sp Sum) roll(_ Node, _ Roller) NodeSpecialization       { return sp }
func (sp Prod) roll(_ Node, _ Roller) NodeSpecialization      { return sp }
func (sp Dice) roll(n Node, roller Roller) NodeSpecialization {
	rolls := make([]RollResult, sp.n)
	for i := 0; i < sp.n; i++ {
		rolls[i].result = roller(sp.x)
		rolls[i].use = false
		rolls[i].order = i
	}
	sort.Slice(rolls, func(i int, j int) bool {
		if rolls[i].result != rolls[j].result {
			return rolls[i].result < rolls[j].result
		}
		return rolls[i].order < rolls[j].order
	})
	for i := 0; i < sp.n; i++ {
		rolls[i].rank = i
		if sp.l <= i && i < sp.h {
			rolls[i].use = true
		}
	}
	sort.Slice(rolls, func(i int, j int) bool {
		return rolls[i].order < rolls[j].order
	})
	return Dice{n: sp.n, x: sp.x, l: sp.l, h: sp.h, rolls: rolls}
}
func (sp Stats) roll(_ Node, _ Roller) NodeSpecialization     { return sp }
func (sp DeathSave) roll(_ Node, _ Roller) NodeSpecialization { return sp }
func (sp Labeled) roll(_ Node, _ Roller) NodeSpecialization   { return sp }
func (sp CommaList) roll(_ Node, _ Roller) NodeSpecialization { return sp }

// Evaluate
func (n Node) value() BR { return n.sp.value(n) }
func (GroupExpr) value(n Node) BR {
	return n.child[0].value()
}
func (sp Natural) value(_ Node) BR {
	return itobr(sp.n)
}
func (sp Sum) value(n Node) BR {
	var ret = zero
	for i, c := range n.child {
		switch sp.ops[i] {
		case "+", "":
			ret = ret.Plus(c.value())
		case "-":
			ret = ret.Minus(c.value())
		}
	}
	return ret
}
func (sp Prod) value(n Node) BR {
	var ret = one
	for i, c := range n.child {
		switch sp.ops[i] {
		case "*", "×", "":
			ret = ret.Times(c.value())
		case "/":
			ret = ret.DivAndRoundTowardsZero(c.value())
		case "//", "÷":
			ret = ret.Div(c.value())
		}
	}
	return ret
}
func (sp Dice) value(_ Node) BR {
	var ret = zero
	for _, rr := range sp.rolls {
		if rr.use {
			ret = ret.Plus(itobr(rr.result))
		}
	}
	return ret
}
func (Stats) value(_ Node) BR { return zero }
func (DeathSave) value(n Node) BR {
	return n.child[0].value()
}
func (sp Labeled) value(n Node) BR {
	return n.child[0].value()
}
func (sp CommaList) value(n Node) BR {
	if len(n.child) == 1 {
		return n.child[0].value()
	} else {
		return zero
	}
}

// Render
// Arguments named "options" can be "l" for allowing inline latex, or "".
func renderRollComment(n Node, rr int, rcok bool) (string, bool) {
	if rcok && rr != RR_NONE && n.rollComment != ROLL_COMMENT_NOTHING && n.rollComment != ROLL_COMMENT_BLOCK_PARENT {
		return n.rollComment, false
	} else {
		return "", rcok
	}
}
func (n Node) renderToplevel(options string) string {
	r1, r2, r3 := n.render("- ", RR_TOP, true, options)
	if r2 != "" {
		return fmt.Sprintf("%s = %s%s", r1, r2, r3)
	} else {
		return fmt.Sprintf("%s%s", r1, r3)
	}
}
func (n Node) render(ind string, rr int, rcok bool, options string) (string, string, string) {
	return n.sp.render(n, ind, rr, rcok, options)
}
func (GroupExpr) render(n Node, ind string, rr int, rcok bool, options string) (string, string, string) {
	r1, r2, r3 := n.child[0].render(ind, rr, rcok, options)
	return fmt.Sprintf("(%s)", r1), r2, r3
}
func renderNumber(n BR, rr int, options string) string {
	if rr == RR_TOP {
		return n.Render(options + "b")
	} else {
		return n.Render(options + "ib")
	}
}
func (sp Natural) render(_ Node, _ string, rr int, _ bool, options string) (string, string, string) {
	return fmt.Sprintf("%d", sp.n), renderNumber(itobr(sp.n), rr, options), ""
}
func (sp Sum) render(n Node, ind string, rr int, rcok bool, options string) (string, string, string) {
	return renderSumProd(n, ind, sp.ops, rr, rcok, options)
}
func (sp Prod) render(n Node, ind string, rr int, rcok bool, options string) (string, string, string) {
	return renderSumProd(n, ind, sp.ops, rr, rcok, options)
}
func renderSumProd(n Node, ind string, ops []string, rr int, rcok bool, options string) (string, string, string) {
	if len(n.child) == 1 {
		return n.child[0].sp.render(n.child[0], ind, rr, rcok, options)
	}
	r1, r2, r3 := "", renderNumber(n.value(), rr, options), ""
	rComment, c_rcok := renderRollComment(n, rr, rcok)
	r2 += rComment
	for i, c := range n.child {
		r1a, _, r3a := c.render(ind, RR_NONE, c_rcok, options)
		effectiveOp := ops[i]
		if effectiveOp == "*" {
			effectiveOp = "×"
		}
		if effectiveOp == "//" {
			effectiveOp = "÷"
		}
		r1 += effectiveOp + r1a
		r3 += r3a
	}
	return r1, r2, r3
}
func (sp Dice) render(n Node, ind string, rr int, rcok bool, options string) (string, string, string) {
	needsRollStr := !(sp.n == 1 && len(sp.rolls) == 1 && sp.rolls[0].use)
	needsDetail := rr == RR_NONE || (rr != RR_DETAIL && needsRollStr)
	rollStr := ""
	if needsRollStr {
		rollsStrs := make([]string, len(sp.rolls))
		for i, rr := range sp.rolls {
			if rr.use {
				rollsStrs[i] = fmt.Sprintf("%d", rr.result)
			} else {
				rollsStrs[i] = fmt.Sprintf("~~%d~~", rr.result)
			}
		}
		rollStr = fmt.Sprintf(" (%s)", strings.Join(rollsStrs, " "))
	}
	detail := ""
	if needsDetail {
		detail = fmt.Sprintf("\n%s*%s%s =* %s", ind, n.token, rollStr, n.value().Render(options+"ib"))
	}
	token := n.token
	if needsRollStr && !needsDetail {
		token += rollStr
	}
	result := renderNumber(n.value(), rr, options)
	if rr == RR_NONE {
		rComment, _ := renderRollComment(n, RR_DETAIL, rcok)
		detail += rComment
	} else {
		rComment, _ := renderRollComment(n, RR_DETAIL, rcok)
		result += rComment
	}
	return token, result, detail
}
func (sp Stats) render(n Node, ind string, _ int, _ bool, options string) (string, string, string) {
	intro := "up a new character! Adventure awaits. In the meanwhile, here are your ability scores:"
	// Extract values and sort them descending
	values := make([]BR, len(n.child))
	for i, c := range n.child {
		values[i] = c.value()
	}
	sort.Slice(values, func(i int, j int) bool {
		return values[j].LessThan(values[i])
	})
	// Render the scores
	scoreText := ""
	for _, v := range values {
		scoreText += fmt.Sprintf("%s, ", v.Render(options+"b"))
	}
	scoreText = scoreText[:len(scoreText)-2]
	// Render details
	details := ""
	for _, c := range n.child {
		_, _, detail := c.render(ind, RR_NONE, false, options)
		details += detail
	}
	return fmt.Sprintf("%s\n%s", intro, scoreText), "", details
}
func (sp DeathSave) render(n Node, ind string, rr int, _ bool, options string) (string, string, string) {
	event := ""
	value := n.value()
	switch {
	case value.Equals(one):
		event = "suffers **A CRITICAL FAIL!** :coffin:"
	case value.LessThanOrEquals(nine):
		event = "**FAILS** :skull:"
	case value.LessThanOrEquals(nineteen):
		event = "**SUCCEEDS** :thumbsup:"
	default:
		event = "**REGAINS 1 HP!** :star-struck:"
	}
	_, _, details := n.child[0].render(ind, RR_NONE, false, options)
	return fmt.Sprintf("a death saving throw, and %s", event), "", details
}
func (sp Labeled) render(n Node, ind string, rr int, rcok bool, options string) (string, string, string) {
	if sp.label == "" {
		return n.child[0].render(ind, rr, rcok, options)
	}
	rComment, c_rcok := renderRollComment(n, rr, rcok)
	switch rr {
	case RR_TOP:
		r1, r2, r3 := n.child[0].render(ind, rr, c_rcok, options)
		return r1, fmt.Sprintf("%s %s%s", r2, sp.label, rComment), r3
	case RR_NONE:
		r1none, _, _ := n.child[0].render("  "+ind, RR_NONE, false, options)
		r1, r2, r3 := n.child[0].render("  "+ind, RR_DETAIL, false, options)
		return r1none, r2, fmt.Sprintf("\n%s*%s =* %s *%s*%s%s", ind, r1, r2, sp.label, rComment, r3)
	case RR_DETAIL:
		r1, r2, r3 := n.child[0].render("  "+ind, rr, false, options)
		return r1, fmt.Sprintf("%s *%s*%s", r2, sp.label, rComment), r3
	default:
		panic("invalid render request in Labeled.render")
	}
}
func (sp CommaList) render(n Node, ind string, rr int, rcok bool, options string) (string, string, string) {
	r1, r2, r3 := "", "", ""
	for i, c := range n.child {
		r1a, r2a, r3a := c.render(ind, rr, rcok, options)
		if i > 0 {
			r1 += ", "
			r2 += ", "
		}
		r1 += r1a
		r2 += r2a
		r3 += r3a
	}
	return r1, r2, r3
}

// roll comment
const ROLL_COMMENT_BLOCK_PARENT = "<ROLL_COMMENT_BLOCK_PARENT>"
const ROLL_COMMENT_NOTHING = "<ROLL_COMMENT_NOTHING>"

func (sp GroupExpr) rollComment(n Node, c configuration) string {
	return n.child[0].sp.rollComment(n.child[0], c)
}
func (sp Natural) rollComment(n Node, _ configuration) string { return ROLL_COMMENT_NOTHING }
func (sp Sum) rollComment(n Node, conf configuration) string  { return rollCommentSumProd(n, conf) }
func (sp Prod) rollComment(n Node, conf configuration) string { return rollCommentSumProd(n, conf) }
func rollCommentSumProd(n Node, conf configuration) string {
	if len(n.child) == 1 {
		return n.child[0].sp.rollComment(n.child[0], conf)
	}
	nofComments := 0
	comment := ""
	for _, c := range n.child {
		cComment := c.sp.rollComment(c, conf)
		if cComment == ROLL_COMMENT_BLOCK_PARENT {
			return ROLL_COMMENT_BLOCK_PARENT
		}
		if cComment != ROLL_COMMENT_NOTHING {
			nofComments++
			comment = cComment
		}
	}
	if nofComments == 1 {
		return comment
	}
	return ROLL_COMMENT_NOTHING
}
func (sp Dice) rollComment(n Node, conf configuration) string {
	if conf.EnableDnd5e && sp.x == 20 && (sp.h-sp.l) == 1 {
		if n.value().Equals(twenty) {
			return " (NAT20! :star-struck:)"
		}
		if n.value().Equals(one) {
			return " (NAT1! :grimacing:)"
		}
	}
	return ROLL_COMMENT_BLOCK_PARENT
}
func (sp Stats) rollComment(n Node, _ configuration) string     { return ROLL_COMMENT_NOTHING }
func (sp DeathSave) rollComment(n Node, _ configuration) string { return ROLL_COMMENT_NOTHING }
func (sp Labeled) rollComment(n Node, conf configuration) string {
	return n.child[0].sp.rollComment(n.child[0], conf)
}
func (sp CommaList) rollComment(n Node, conf configuration) string {
	if len(n.child) == 1 {
		return n.child[0].sp.rollComment(n.child[0], conf)
	}
	return ROLL_COMMENT_NOTHING
}

// Probability distributions
func (n Node) prob() PD {
	return n.sp.prob(n)
}
func (sp GroupExpr) prob(n Node) PD {
	return n.child[0].prob()
}
func (sp Natural) prob(n Node) PD {
	return pd.Constant(n.value())
}
func (sp Sum) prob(n Node) PD {
	var ret = zeroPD
	for i, c := range n.child {
		switch sp.ops[i] {
		case "+", "":
			ret = ret.Plus(c.prob())
		case "-":
			ret = ret.Minus(c.prob())
		}
	}
	return ret
}
func (sp Prod) prob(n Node) PD {
	var ret = onePD
	for i, c := range n.child {
		switch sp.ops[i] {
		case "*", "×", "":
			ret = ret.Times(c.prob())
		case "/":
			ret = ret.DivAndRoundTowardsZero(c.prob())
		case "//", "÷":
			ret = ret.Div(c.prob())
		}
	}
	return ret
}
func (sp Dice) prob(_ Node) PD {
	return pd.Dice(sp.n, sp.x, sp.l, sp.n-sp.h)
}
func (Stats) prob(n Node) PD {
	return errPD // todo: maybe list of probs after sorting
}
func (DeathSave) prob(n Node) PD {
	return errPD // todo: maybe constant string.
}
func (sp Labeled) prob(n Node) PD {
	return n.child[0].prob()
}
func (sp CommaList) prob(n Node) PD {
	if len(n.child) == 1 {
		return n.child[0].prob()
	} else {
		return errPD // todo: maybe list of probs
	}
}
