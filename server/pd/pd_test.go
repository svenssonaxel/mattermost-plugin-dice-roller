package pd_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moussetc/mattermost-plugin-dice-roller/server/br"
	"github.com/moussetc/mattermost-plugin-dice-roller/server/pd"
)

var n = br.New

func TestDice(t *testing.T) {
	// Model checking.
	for numberOfDice := 0; numberOfDice <= 4; numberOfDice++ {
		for sides := 0; sides <= 4; sides++ {
			for dropLow := 0; dropLow <= numberOfDice; dropLow++ {
				for dropHigh := 0; dropHigh <= (numberOfDice - dropLow); dropHigh++ {
					expectedDist := getCorrectDiceDistribution(numberOfDice, sides, dropLow, dropHigh)
					actualDist := pd.Dice(numberOfDice, sides, dropLow, dropHigh)
					msg := fmt.Sprintf("numberOfDice=%d, sides=%d, dropLow=%d, dropHigh=%d", numberOfDice, sides, dropLow, dropHigh)
					assert.True(t, expectedDist.Equals(actualDist), msg)
					assert.Equal(t, expectedDist.Render(""), actualDist.Render(""), msg)
					assert.Equal(t, expectedDist.ExpectedValue().Render(""), actualDist.ExpectedValue().Render(""), msg)
				}
			}
		}
	}
}

// A very explicit and inefficient reference implementation for pd.Dice.
func getCorrectDiceDistribution(numberOfDice, sides, dropLow, dropHigh int) pd.PD {
	if numberOfDice == 0 || sides == 0 || dropLow+dropHigh >= numberOfDice {
		return pd.Constant(br.Zero)
	}
	distribution := make([]int, numberOfDice*sides+1)
	count := 0
	// Create a slice of the correct size and fill it with 1.
	outcome := make([]int, numberOfDice)
	for i := range outcome {
		outcome[i] = 1
	}
	// Iterate over all possible outcomes.
	dobreak := false
	for {
		sum := 0
		sortedOutcome := make([]int, len(outcome))
		copy(sortedOutcome, outcome)
		sort.Ints(sortedOutcome)
		for i := dropLow; i < len(sortedOutcome)-dropHigh; i++ {
			sum += sortedOutcome[i]
		}
		distribution[sum]++
		count++
		// Increment the outcome.
		for i := len(outcome) - 1; i >= 0; i-- {
			if outcome[i] < sides {
				outcome[i]++
				break
			} else {
				outcome[i] = 1
				if i == 0 {
					dobreak = true
				}
			}
		}
		if dobreak {
			break
		}
	}
	// Convert the distribution to a probability distribution.
	den := n(count)
	var lcTerms []pd.LCTerm
	for i, v := range distribution {
		if v != 0 {
			lcTerms = append(lcTerms, pd.LCTerm{Coeff: n(v).Div(den), PD: pd.Constant(n(i))})
		}
	}
	return pd.LinearCombination(lcTerms)
}

// Test high number of dice/sides.
func TestDiceHigh(t *testing.T) {
	for n := 0; n <= 20; n++ {
		for s := 0; s <= 20; s++ {
			pd.Dice(n, s, 0, 0)
		}
	}
	pd_100d100 := pd.Dice(100, 100, 0, 0)
	assert.Equal(t, "5050", pd_100d100.ExpectedValue().Render(""))
	pd_2d2000 := pd.Dice(2, 2000, 0, 0)
	assert.Equal(t, "2001", pd_2d2000.ExpectedValue().Render(""))
	pd_2000d2 := pd.Dice(1000, 2, 0, 0)
	assert.Equal(t, "1500", pd_2000d2.ExpectedValue().Render(""))
	pd_4d500 := pd.Dice(4, 500, 0, 0)
	assert.Equal(t, "1002", pd_4d500.ExpectedValue().Render(""))
	pd_500d4 := pd.Dice(500, 4, 0, 0)
	assert.Equal(t, "1250", pd_500d4.ExpectedValue().Render(""))
	pd_1d10000 := pd.Dice(1, 10000, 0, 0)
	assert.Equal(t, "5000.5", pd_1d10000.ExpectedValue().Render(""))
	pd_1000d1 := pd.Dice(1000, 1, 0, 0)
	assert.Equal(t, "1000", pd_1000d1.ExpectedValue().Render(""))
}

// Test complex dice expressions.
func TestDiceComplex(t *testing.T) {
	res := pd.Dice(30, 30, 0, 0)
	for i := 0; i < 30; i++ {
		res = res.Plus(pd.OnePD)
		res = res.Times(pd.OnePD)
		res = res.Minus(pd.OnePD)
		res = res.Div(pd.OnePD)
	}
	assert.Equal(t, "465", res.ExpectedValue().Render(""))

	expr_a := pd.Dice(20, 12, 0, 0)
	expr_b := pd.Dice(30, 6, 2, 0)
	expr_c := pd.Constant(n(7))
	expr_d := pd.Constant(n(2))
	res2 := expr_a.Plus(expr_b).Plus(expr_c).Div(expr_d)
	assert.Equal(t, "119 1697959580431178797867/1727139997818229358592", res2.ExpectedValue().Render(""))
}

// Test high number of dropped dice.
func TestDiceDrop(t *testing.T) {
	d6_1 := pd.Dice(12, 12, 5, 0)
	d6_2 := pd.Dice(12, 12, 0, 5)
	d6_3 := pd.Dice(12, 12, 3, 3)
	assert.Equal(t, "61 423896133343/743008370688", d6_1.ExpectedValue().Render(""))
	assert.Equal(t, "29 319112237345/743008370688", d6_2.ExpectedValue().Render(""))
	assert.Equal(t, "39", d6_3.ExpectedValue().Render(""))
}
