package pd_test

import (
	"sort"
	"testing"

	"github.com/moussetc/mattermost-plugin-dice-roller/server/br"
	"github.com/moussetc/mattermost-plugin-dice-roller/server/pd"
	"github.com/stretchr/testify/assert"
)

var n = br.New

func TestProb(t *testing.T) {
	d6 := pd.UniformInt(1, 6)
	assert.Equal(t, "3 1/2", d6.ExpectedValue().Render(false, false, false))
}

func TestDice(t *testing.T) {
	d6_1 := pd.UniformInt(1, 6)
	d6_2 := pd.Dice(1, 6, 0, 0)
	assert.True(t, d6_1.Equals(d6_2))
	assert.Equal(t, d6_1.Render(false), d6_2.Render(false))
	// Model checking.
	for numberOfDice := 0; numberOfDice <= 4; numberOfDice++ {
		for sides := 0; sides <= 4; sides++ {
			for dropLow := 0; dropLow <= numberOfDice; dropLow++ {
				for dropHigh := 0; dropHigh <= (numberOfDice - dropLow); dropHigh++ {
					expectedDist := getCorrectDiceDistribution(numberOfDice, sides, dropLow, dropHigh)
					actualDist := pd.Dice(numberOfDice, sides, dropLow, dropHigh)
					assert.True(t, expectedDist.Equals(actualDist), "numberOfDice=%d, sides=%d, dropLow=%d, dropHigh=%d", numberOfDice, sides, dropLow, dropHigh)
					assert.Equal(t, expectedDist.Render(false), actualDist.Render(false), "numberOfDice=%d, sides=%d, dropLow=%d, dropHigh=%d", numberOfDice, sides, dropLow, dropHigh)
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
