package pd

import (
	"fmt"
	"sort"

	"github.com/moussetc/mattermost-plugin-dice-roller/server/br"
)

// Probability distributions

type PD struct {
	// The PD struct represents a probability distribution by mapping an outcome
	// to a probability. Both outcome and probability are of type BR (big rational
	// number). However, a BR doesn't work as a map key, so we use the outcome's
	// string representation as the key. The value holds a struct containing both
	// the outcome itself as well as the probability.
	probMapOP  probMapOP
	normalized bool
}
type probMapOP = map[string]probOP
type probOP = struct {
	outcome     BR
	probability BR
}

type BR = br.BR

var n = br.New
var nan = br.Nan
var zero = br.Zero
var one = br.One
var two = n(2)

var ErrPD = PD{probMapOP{"NaN": {nan, nan}}, true}
var ZeroPD = Constant(zero)
var OnePD = Constant(one)

func (pd PD) EnsureNormalized() PD {
	if !pd.normalized {
		for _, v := range pd.probMapOP {
			// v.outcome is already normalized, since .String() was called when
			// creating the map entry.
			v.probability.EnsureNormalized()
		}
		pd.normalized = true
	}
	return pd
}

func (pd PD) Get(outcome BR) BR {
	return pd.getWithStr(outcome.String())
}

func (pd PD) getWithStr(outcomeStr string) BR {
	ret, ok := pd.probMapOP[outcomeStr]
	if !ok {
		return zero
	}
	return ret.probability
}

func (pd PD) Equals(pd2 PD) bool {
	for k, v := range pd.probMapOP {
		if !v.probability.Equals(pd2.getWithStr(k)) {
			return false
		}
	}
	for k, v := range pd2.probMapOP {
		if !v.probability.Equals(pd.getWithStr(k)) {
			return false
		}
	}
	return true
}

// Basic operations

func Constant(outcome BR) PD {
	outcomeTxt := outcome.String()
	// The .String call ensures that the outcome is normalized.
	return PD{probMapOP{outcomeTxt: {outcome, one}}, true}
}

func (pd PD) isConstant() (bool, BR) {
	if len(pd.probMapOP) == 1 {
		for _, v := range pd.probMapOP {
			return true, v.outcome
		}
	}
	return false, nan
}

// Given a probability distribution and a function, return a new probability
// distribution where each outcome is mapped to the result of applying the
// function to the original outcome. The function MUST be a bijection.
func mapOutcome1(pd PD, f func(BR) BR) PD {
	ret := PD{make(probMapOP), pd.normalized}
	for _, v := range pd.probMapOP {
		k := f(v.outcome)
		kStr := k.String()
		ret.probMapOP[kStr] = probOP{k, v.probability}
	}
	return ret
}

// Given two probability distributions and a function, return a new probability
// distribution that represents the probability distribution of the result of
// applying the function to the outcomes of the two input distributions.
func mapOutcome(pd, pd2 PD, f func(BR, BR) BR) PD {
	// Since we'll do a lot of operations on these numbers, it's worth it to
	// normalize them first.
	pd.EnsureNormalized()
	pd2.EnsureNormalized()
	ret := PD{make(probMapOP), false}
	for _, v1 := range pd.probMapOP {
		for _, v2 := range pd2.probMapOP {
			k := f(v1.outcome, v2.outcome)
			v := v1.probability.Times(v2.probability)
			kStr := k.String()
			ret.probMapOP[kStr] = probOP{k, ret.getWithStr(kStr).Plus(v)}
		}
	}
	return ret
}

func (pd PD) Plus(pd2 PD) PD {
	if isConst, a := pd.isConstant(); isConst {
		a.EnsureNormalized()
		if a.Equals(zero) {
			return pd2
		}
		return mapOutcome1(pd2, func(b BR) BR { return a.Plus(b) })
	}
	if isConst, b := pd2.isConstant(); isConst {
		b.EnsureNormalized()
		if b.Equals(zero) {
			return pd
		}
		return mapOutcome1(pd, func(a BR) BR { return a.Plus(b) })
	}
	return mapOutcome(pd, pd2, func(a, b BR) BR { return a.Plus(b) })
}

func (pd PD) Minus(pd2 PD) PD {
	if isConst, a := pd.isConstant(); isConst {
		a.EnsureNormalized()
		return mapOutcome1(pd2, func(b BR) BR { return a.Minus(b) })
	}
	if isConst, b := pd2.isConstant(); isConst {
		b.EnsureNormalized()
		if b.Equals(zero) {
			return pd
		}
		return mapOutcome1(pd, func(a BR) BR { return a.Minus(b) })
	}
	return mapOutcome(pd, pd2, func(a, b BR) BR { return a.Minus(b) })
}

func (pd PD) Times(pd2 PD) PD {
	if isConst, a := pd.isConstant(); isConst {
		a.EnsureNormalized()
		if a.Equals(one) {
			return pd2
		}
		return mapOutcome1(pd2, func(b BR) BR { return a.Times(b) })
	}
	if isConst, b := pd2.isConstant(); isConst {
		b.EnsureNormalized()
		if b.Equals(one) {
			return pd
		}
		return mapOutcome1(pd, func(a BR) BR { return a.Times(b) })
	}
	return mapOutcome(pd, pd2, func(a, b BR) BR { return a.Times(b) })
}

func (pd PD) Div(pd2 PD) PD {
	if isConst, a := pd.isConstant(); isConst {
		a.EnsureNormalized()
		return mapOutcome1(pd2, func(b BR) BR { return a.Div(b) })
	}
	if isConst, b := pd2.isConstant(); isConst {
		b.EnsureNormalized()
		if b.Equals(one) {
			return pd
		}
		return mapOutcome1(pd, func(a BR) BR { return a.Div(b) })
	}
	return mapOutcome(pd, pd2, func(a, b BR) BR { return a.Div(b) })
}

func (pd PD) DivAndRoundTowardsZero(pd2 PD) PD {
	return mapOutcome(pd, pd2, func(a, b BR) BR { return a.DivAndRoundTowardsZero(b) })
}

func (pd PD) ExpectedValue() BR {
	ret := zero
	for _, v := range pd.probMapOP {
		ret = ret.Plus(v.outcome.Times(v.probability))
	}
	return ret
}

func LinearCombination(terms []LCTerm) PD {
	ret := PD{make(probMapOP), false}
	for _, term := range terms {
		for k, v := range term.PD.probMapOP {
			ret.probMapOP[k] = probOP{v.outcome, ret.getWithStr(k).Plus(term.Coeff.Times(v.probability))}
		}
	}
	return ret
}

type LCTerm struct {
	PD    PD
	Coeff BR
}

// Return a string describing the average value of the distribution and the
// probability of each outcome in a table.
//
// The options string is either "l" for allowing inline LaTeX, or "" otherwise.
func (pd PD) Render(options string) string {
	// Extract and sort the outcomes
	outcomes := make([]BR, 0, len(pd.probMapOP))
	for _, v := range pd.probMapOP {
		outcomes = append(outcomes, v.outcome)
	}
	sort.Slice(outcomes, func(i, j int) bool {
		return outcomes[i].LessThan(outcomes[j])
	})
	// Render table
	table := ""
	cumulative := one
	for _, k := range outcomes {
		v := pd.Get(k)
		if v.Equals(zero) {
			continue
		}
		var kRendered, vRendered, cumulativeRendered string
		kRendered = k.Render(options + "b")
		vRendered = v.Render(options + "p")
		cumulativeRendered = cumulative.Render(options + "p")
		table += fmt.Sprintf("\n|%s|%s|%s|", kRendered, vRendered, cumulativeRendered)
		cumulative = cumulative.Minus(v)
	}
	if !cumulative.Equals(zero) {
		table += fmt.Sprintf("\n|Probability unaccounted for|%s|**ERROR**|", cumulative.Render(options+"p"))
	}
	// Render header and expected value
	expectedValue := pd.ExpectedValue().Render(options + "b")
	return fmt.Sprintf("Average: %s\n\n|Outcome|Chance to get|Chance to get at least|\n|-|-|-|%s", expectedValue, table)
}
