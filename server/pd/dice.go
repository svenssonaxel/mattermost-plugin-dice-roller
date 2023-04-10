package pd

import (
	"fmt"
	"math/big"

	"github.com/moussetc/mattermost-plugin-dice-roller/server/br"
)

// Memoization wrapper for dice
func Dice(numberOfDice, sides, dropLow, dropHigh int) PD {
	if numberOfDice < 0 || sides < 0 || dropLow < 0 || dropHigh < 0 || numberOfDice < dropLow+dropHigh {
		return ErrPD
	}
	if numberOfDice == 0 || sides == 0 || numberOfDice == dropLow+dropHigh {
		return ZeroPD
	}
	// When not dropping any dice, cache if
	// A) the number of dice and sides are small, since these are common; or
	// B) the number of dice is a multiple of or less than 4, to use as starting
	//    points when calculating the others.
	// C) the number of sides or dice is 1, since these will be frequently used
	//    as starting points for other calculations.
	//
	// Never cache when dropping low dice, since the dice function will use a
	// symmetry to reduce the number of cache entries.
	//
	// When dropping dice, cache only if dropping an even number of high dice and
	// no low dice. Odd-numbered cases will depend only on the cacheable
	// even-numbered cases, and caching less than this would be wasteful since the
	// sub-problems begin overlapping at the second level of recursion.
	cacheAble := (dropLow == 0 && dropHigh == 0 && ((numberOfDice <= 5 && sides <= 20) || (numberOfDice%4 == 0 || numberOfDice < 4) || sides == 1 || numberOfDice == 1)) || (dropLow == 0 && dropHigh%2 == 0)
	var cacheKey string
	if cacheAble {
		cacheKey = fmt.Sprintf("%d,%d,%d,%d", numberOfDice, sides, dropLow, dropHigh)
		if ret, ok := diceCache[cacheKey]; ok {
			// Since this is already in the cache, it's the second time we need it,
			// and we want popular cache entries to be normalized.
			ret.EnsureNormalized()
			return ret
		}
	}
	ret := dice(numberOfDice, sides, dropLow, dropHigh)
	if cacheAble {
		if len(diceCache) > 10000 {
			diceCache = make(map[string]PD)
		}
		diceCache[cacheKey] = ret
	}
	return ret
}

var diceCache = make(map[string]PD)

// Probability distribution for dice rolls

func dice(numberOfDice, sides, dropLow, dropHigh int) PD {
	// With 0 dice, the result is a constant 0.
	if numberOfDice == 0 {
		return ZeroPD
	}
	// With only 1 side, the result is a constant.
	if sides == 1 {
		return Constant(n(sides * (numberOfDice - dropLow - dropHigh)))
	}
	// When not dropping any dice, it's possible to use the more efficient
	// combinatorial method.
	if dropLow == 0 && dropHigh == 0 {
		return diceCombinatorial(numberOfDice, sides)
	}
	// When dropping dice, we have to use the recursive method.
	return diceRecursiveWithDrops(numberOfDice, sides, dropLow, dropHigh)
}

// Cache for binomial coefficients, used by fastBinomial

type binomialCacheT = struct {
	binomials map[struct{ n, k int }]*big.Int
}

var binomialCache binomialCacheT

func resetBinomialCache() {
	binomialCache.binomials = make(map[struct{ n, k int }]*big.Int)
}

func init() {
	resetBinomialCache()
}

func binomialAvailable(n, k int) bool {
	if k > (n >> 1) {
		k = n - k
	}
	if n < 0 {
		return false
	}
	if k <= 2 || n-k <= 2 {
		return true
	}
	_, ok := binomialCache.binomials[struct{ n, k int }{n, k}]
	return ok
}

// Compute the binomial coefficient `n` choose `k`, faster than
// big.Int.Binomial. The speedup is achieved by
// A) Caching
// B) Derivation of values from already computed values
func fastBinomial(target, tmp *big.Int, n, k, s int) {
	if k > (n >> 1) {
		k = n - k
	}
	if k < 0 {
		target.SetInt64(0)
		return
	}
	if k == 0 {
		target.SetInt64(1)
		return
	}
	if k == 1 {
		target.SetInt64(int64(n))
		return
	}
	if k == 2 {
		target.SetInt64(int64(n * (n - 1) / 2))
		return
	}
	key := struct{ n, k int }{n, k}
	if cached, ok := binomialCache.binomials[key]; ok {
		target.Set(cached)
		return
	}
	// Compute the binomial coefficient, using existing results as available. We
	// only recurse after checking availability, so we don't need to pass along
	// the tmp parameter.
	result := new(big.Int)
	binomialAvailable_nMinusOne_k := binomialAvailable(n-1, k)
	switch {
	case binomialAvailable(n-1, k-1) && binomialAvailable_nMinusOne_k:
		// (n choose k) = (n-1 choose k-1) + (n-1 choose k)
		fastBinomial(result, nil, n-1, k-1, s)
		fastBinomial(tmp, nil, n-1, k, s)
		result.Add(result, tmp)
	case binomialAvailable_nMinusOne_k: // This is the most common case when computing large binomials.
		// (n choose k) = (n-1 choose k) * n / (n-k)
		fastBinomial(result, nil, n-1, k, s)
		result.Mul(result, tmp.SetInt64(int64(n)))
		result.Div(result, tmp.SetInt64(int64(n-k)))
	case binomialAvailable(n+1, k) && binomialAvailable(n, k-1):
		// (n choose k) = (n+1 choose k) - (n choose k-1)
		fastBinomial(result, nil, n+1, k, s)
		fastBinomial(tmp, nil, n, k-1, s)
		result.Sub(result, tmp)
	case binomialAvailable(n+1, k+1) && binomialAvailable(n, k+1):
		// (n choose k) = (n+1 choose k+1) - (n choose k+1)
		fastBinomial(result, nil, n+1, k+1, s)
		fastBinomial(tmp, nil, n, k+1, s)
		result.Sub(result, tmp)
	case binomialAvailable(n, k-1):
		// (n choose k) = (n choose k-1) * (n-k+1) / k
		fastBinomial(result, nil, n, k-1, s)
		result.Mul(result, tmp.SetInt64(int64(n-k+1)))
		result.Div(result, tmp.SetInt64(int64(k)))
	case binomialAvailable(n-1, k-1):
		// (n choose k) = (n-1 choose k-1) * n / k
		fastBinomial(result, nil, n-1, k-1, s)
		result.Mul(result, tmp.SetInt64(int64(n)))
		result.Div(result, tmp.SetInt64(int64(k)))
	default:
		// Due to the pattern of invocation from diceCombinatorial, this should
		// never happen.
		result.Binomial(int64(n), int64(k))
	}
	// Cache the result
	binomialCache.binomials[key] = result
	target.Set(result)
}

// Probability distribution for the sum of `n` dice with `s` sides. Implemented
// with a combinatorial approach.
func diceCombinatorial(n, s int) PD {
	mapOP := make(probMapOP)
	low := n
	high := n * s
	sPowN := new(big.Int).Exp(big.NewInt(int64(s)), big.NewInt(int64(n)), nil)
	if sPowN.Sign() != 1 {
		return ErrPD
	}
	sum := new(big.Int)
	a := new(big.Int)
	tmp := new(big.Int)
	tmp2 := new(big.Int)
	for {
		T := low
		// Compute the probability of rolling `T` with `n` dice with `s` sides each.
		// [1] https://towardsdatascience.com/modelling-the-probability-distributions-of-dice-b6ecf87b24ea
		// According to [1], this probability is
		// P(n, s, T) = 1/s^n * sum_{k=0}^{floor((T-n)/s)} (-1)^k * a(k, T, n, s) where
		// a(k, T, n, s) = Binomial(n, k) * Binomial(T-s*k-1, n-1)
		// Todo: Translate the optimization that: if n <= s, then some of the factors in the numerator and denominator are the same, so we can cancel them out.
		limit := (T - n) / s
		var prob BR
		if limit < 0 {
			prob = br.Zero
		} else {
			sum.SetInt64(0)
			for k := 0; k <= limit; k++ {
				fastBinomial(a, tmp2, n, k, s)
				fastBinomial(tmp, tmp2, T-s*k-1, n-1, s)
				a.Mul(a, tmp)
				if k%2 == 1 {
					a.Neg(a)
				}
				sum.Add(sum, a)
			}
			prob = br.FromBigInt(sum, sPowN)
		}
		// Populate the map with the probability of rolling `low` and `high`. Due to
		// symmetry, prob is the same for both.
		lowBR := br.New(low)
		mapOP[lowBR.String()] = probOP{lowBR, prob}
		highBR := br.New(high)
		mapOP[highBR.String()] = probOP{highBR, prob}
		if high-low <= 1 {
			break
		}
		// Prepare for next iteration.
		low++
		high--
	}
	return PD{mapOP, false}
}

func diceRecursiveWithDrops(numberOfDice, sides, dropLow, dropHigh int) PD {
	// When we drop more low than high dice, we can use a symmetry to reduce the
	// number of cache entries. In practice, at least one of dropHigh and dropLow
	// will be 0.
	if dropHigh < dropLow {
		return Constant(n((sides + 1) * (numberOfDice - dropLow - dropHigh))).Minus(Dice(numberOfDice, sides, dropHigh, dropLow))
	}
	// When dropping more high than low dice, we use a dynamic programming
	// technique to calculate the probability distribution. We break the problem
	// down into sub-problems based on the number of dice that roll above average
	// value, k. For each k, we calculate the probability of k dice rolling above
	// average value and the rest not. We then calculate the probability
	// distribution of the roll in the given situation. This probability
	// distribution depends on two sub-problems with less dice sides.
	lcTerms := make([]LCTerm, numberOfDice+1)
	numberOfDiceBR := n(numberOfDice)
	sidesBR := n(sides)
	limit := (sides + 1) / 2
	limitBR := n(limit)
	totalProbability := zero
	probabilityRollNotHigherThanLimit := limitBR.Div(sidesBR)
	probabilityRollHigherThanLimit := one.Minus(probabilityRollNotHigherThanLimit)
	for k := 0; k <= numberOfDice; k++ {
		rest := numberOfDice - k
		kBR := n(k)
		restBR := n(rest)
		// A_k = "k dice roll above the limit, the rest do not."
		// Calculate the probability of A_k.
		probabilityKRollHigherThanLimit := probabilityRollHigherThanLimit.Pow(kBR)
		probabilityRestNotHigherThanLimit := probabilityRollNotHigherThanLimit.Pow(restBR)
		combos := numberOfDiceBR.Binomial(kBR)
		coeff := probabilityKRollHigherThanLimit.Times(probabilityRestNotHigherThanLimit).Times(combos)
		// Calculate the probability distribution of the roll, given A_k.
		effectiveDropLow := max(0, dropLow-rest)
		effectiveDropHigh := min(k, dropHigh)
		probHigher := Dice(k, sides-limit, effectiveDropLow, effectiveDropHigh).Plus(Constant(n(limit * (k - effectiveDropLow - effectiveDropHigh))))
		probLower := Dice(rest, limit, min(rest, dropLow), max(0, dropHigh-k))
		prob := probHigher.Plus(probLower)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
