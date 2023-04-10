package pd

import (
	"fmt"
	"sort"

	"github.com/moussetc/mattermost-plugin-dice-roller/server/br"
)

// Probability distributions

type PD struct {
	probs probs
}
type probs = map[string]probKV
type probKV = struct {
	key   BR
	value BR
}

type BR = br.BR

var n = br.New
var nan = br.Nan
var zero = br.Zero
var one = br.One

var ErrPD = PD{probs{"NaN": {nan, nan}}}
var zeroPD = Constant(zero)

func (pr PD) Get(outcome BR) BR {
	outcomeStr := outcome.String()
	ret, ok := pr.probs[outcomeStr]
	if !ok {
		return zero
	}
	return ret.value
}

func (r1 PD) Equals(r2 PD) bool {
	for _, v := range r1.probs {
		if !v.value.Equals(r2.Get(v.key)) {
			return false
		}
	}
	for _, v := range r2.probs {
		if !v.value.Equals(r1.Get(v.key)) {
			return false
		}
	}
	return true
}

// Basic operations

func Constant(outcome BR) PD {
	outcomeTxt := outcome.String()
	return PD{probs{outcomeTxt: {outcome, one}}}
}

func UniformInt(low, high int) PD {
	if low > high {
		return ErrPD
	}
	probs := make(probs)
	value := one.Div(n(high - low + 1))
	for i := low; i <= high; i++ {
		br := n(i)
		probs[br.String()] = probKV{br, value}
	}
	return PD{probs}
}

func mapOutcome(r1, r2 PD, f func(BR, BR) BR) PD {
	ret := PD{make(probs)}
	for _, v1 := range r1.probs {
		for _, v2 := range r2.probs {
			k := f(v1.key, v2.key)
			v := v1.value.Times(v2.value)
			ret.probs[k.String()] = probKV{k, ret.Get(k).Plus(v)}
		}
	}
	return ret
}

func (r1 PD) Plus(r2 PD) PD {
	return mapOutcome(r1, r2, func(a, b BR) BR { return a.Plus(b) })
}

func (r1 PD) Minus(r2 PD) PD {
	return mapOutcome(r1, r2, func(a, b BR) BR { return a.Minus(b) })
}

func (r1 PD) Times(r2 PD) PD {
	return mapOutcome(r1, r2, func(a, b BR) BR { return a.Times(b) })
}

func (r1 PD) Div(r2 PD) PD {
	return mapOutcome(r1, r2, func(a, b BR) BR { return a.Div(b) })
}

func (r1 PD) DivAndRoundTowardsZero(r2 PD) PD {
	return mapOutcome(r1, r2, func(a, b BR) BR { return a.DivAndRoundTowardsZero(b) })
}

func (r PD) ExpectedValue() BR {
	ret := zero
	for _, v := range r.probs {
		ret = ret.Plus(v.key.Times(v.value))
	}
	return ret
}

// Probability distribution for dice rolls

func dice(numberOfDice, sides, dropLow, dropHigh int) PD {
	if numberOfDice < 0 || sides < 0 || dropLow < 0 || dropHigh < 0 || numberOfDice < dropLow+dropHigh {
		return ErrPD
	}
	if numberOfDice == 0 || sides == 0 || numberOfDice == dropLow+dropHigh {
		return zeroPD
	}
	lcTerms := make([]LCTerm, numberOfDice+1)
	numberOfDiceBR := n(numberOfDice)
	sidesBR := n(sides)
	totalProbability := zero
	probabilityRollHighest := one.Div(sidesBR)
	probabilityRollNotHighest := one.Minus(probabilityRollHighest)
	for k := 0; k <= numberOfDice; k++ {
		kBR := n(k)
		restBR := numberOfDiceBR.Minus(kBR)
		// A_k = "k dice roll their highest value, the rest do not."
		// Calculate the probability of A_k.
		probabilityKRollHighest := probabilityRollHighest.Pow(kBR)
		probabilityRestNotHighest := probabilityRollNotHighest.Pow(restBR)
		combos := numberOfDiceBR.Binomial(kBR)
		coeff := probabilityKRollHighest.Times(probabilityRestNotHighest).Times(combos)
		// Calculate the probability distribution of the roll, given A_k.
		var prob PD
		if k <= dropHigh { // < is also correct
			prob = Dice(numberOfDice-k, sides-1, dropLow, dropHigh-k)
		} else if k < numberOfDice-dropLow { // <= is also correct
			prob = Constant(n(sides * (k - dropHigh))).Plus(Dice(numberOfDice-k, sides-1, dropLow, 0))
		} else {
			prob = Constant(n(sides * (numberOfDice - dropLow - dropHigh)))
		}
		lcTerms[k] = LCTerm{
			Coeff: coeff,
			PD:    prob,
		}
		totalProbability = totalProbability.Plus(coeff)
	}
	if !totalProbability.Equals(one) {
		panic("totalProbability != 1 in prob.dice")
	}
	return LinearCombination(lcTerms)
}

func LinearCombination(terms []LCTerm) PD {
	ret := PD{make(probs)}
	for _, term := range terms {
		for k, v := range term.PD.probs {
			ret.probs[k] = probKV{v.key, ret.Get(v.key).Plus(term.Coeff.Times(v.value))}
		}
	}
	return ret
}

type LCTerm struct {
	PD    PD
	Coeff BR
}

// Memoization wrapper for dice
func Dice(numberOfDice, sides, dropLow, dropHigh int) PD {
	key := fmt.Sprintf("%d,%d,%d,%d", numberOfDice, sides, dropLow, dropHigh)
	if ret, ok := diceCache[key]; ok {
		return ret
	}
	ret := dice(numberOfDice, sides, dropLow, dropHigh)
	if len(diceCache) > 10000 {
		diceCache = make(map[string]PD)
	}
	diceCache[key] = ret
	return ret
}

var diceCache = make(map[string]PD)

// String rendering and parsing

func (r PD) Render(latex bool) string {
	// Extract and sort the keys
	keys := make([]BR, 0, len(r.probs))
	for _, v := range r.probs {
		keys = append(keys, v.key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].LessThan(keys[j])
	})
	// Render table
	table := ""
	cumulative := one
	for _, k := range keys {
		v := r.Get(k)
		if v.Equals(zero) {
			continue
		}
		var kRendered, vRendered, cumulativeRendered string
		kRendered = k.Render(latex, false, true)
		vRendered = v.Render(latex, true, false)
		cumulativeRendered = cumulative.Render(latex, true, false)
		table += fmt.Sprintf("\n|%s|%s|%s|", kRendered, vRendered, cumulativeRendered)
		cumulative = cumulative.Minus(v)
	}
	if !cumulative.Equals(zero) {
		table += fmt.Sprintf("\n|Probability unaccounted for|%s|**ERROR**|", cumulative.Render(latex, true, false))
	}
	// Render header and expected value
	expectedValue := r.ExpectedValue().Render(latex, false, true)
	return fmt.Sprintf("Average: %s\n\n|Outcome|Chance to get|Chance to get at least|\n|-|-|-|%s", expectedValue, table)
}
