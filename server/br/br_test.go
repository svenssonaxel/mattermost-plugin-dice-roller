package br_test

import (
	"strconv"
	"testing"

	"github.com/moussetc/mattermost-plugin-dice-roller/server/br"
	"github.com/stretchr/testify/assert"
)

var n = br.New

func TestBigRat(t *testing.T) {
	// Test constants
	zero := n(0)
	one := n(1)
	negOne := zero.Minus(one)
	half := one.Div(n(2))
	negHalf := zero.Minus(half)
	sevenThirds := n(7).Div(n(3))
	negSevenThirds := zero.Minus(sevenThirds)
	nan1 := one.Div(zero)
	nan2 := nan1.Plus(one)
	nan3 := one.Minus(nan2)
	two := n(2)
	three := n(1).Div(n(16)).Times(n(48))

	assert.True(t, br.Nan.Equals(nan1))
	assert.True(t, br.Zero.Equals(zero))
	assert.True(t, br.One.Equals(one))

	// Test cases, in sorted order
	testCases := []struct {
		bigrat br.BR
		txt    string
		isInt  bool
	}{
		{nan1, "NaN", false},
		{nan2, "NaN", false},
		{nan3, "NaN", false},
		{negSevenThirds, "-7/3", false},
		{negOne, "-1", true},
		{negHalf, "-1/2", false},
		{zero, "0", true},
		{half, "1/2", false},
		{one, "1", true},
		{two, "2", true},
		{sevenThirds, "7/3", false},
		{three, "3", true},
	}

	for idx_a, tc_a := range testCases {
		a := tc_a.bigrat
		msg := "Test case: " + strconv.Itoa(idx_a)
		// Test IsNaN
		assert.Equal(t, tc_a.txt == "NaN", a.IsNaN(), msg)
		// Test IsInt
		assert.Equal(t, tc_a.isInt, a.IsInt(), msg)
		// Test most Arithmetic: Equals, LessThan, LessThanOrEquals, Plus, Minus, Times, Div, DivAndRoundTowardsZero
		assert.Equal(t, tc_a.txt, a.Plus(zero).String(), msg)
		assert.Equal(t, tc_a.txt, a.Minus(zero).String(), msg)
		assert.Equal(t, turner(a.IsNaN(), "NaN", "0"), a.Minus(a).Render(false, false, false), msg)
		assert.Equal(t, tc_a.txt, a.Times(one).String(), msg)
		assert.Equal(t, turner(a.IsNaN(), "NaN", "0"), a.Times(zero).Render(false, false, false), msg)
		assert.Equal(t, a.Plus(a).String(), a.Times(two).String(), msg)
		assert.Equal(t, tc_a.txt, a.Div(one).String(), msg)
		assert.Equal(t, "NaN", a.Div(zero).String(), msg)
		assert.Equal(t, turner(a.IsNaN() || tc_a.txt == "0", "NaN", "1"), a.Div(a).Render(false, false, false), msg)
		for idx_b, tc_b := range testCases {
			b := tc_b.bigrat
			msg := msg + ", " + strconv.Itoa(idx_b)
			assert.Equal(t, tc_a.txt == tc_b.txt, a.Equals(b), msg)
			assert.Equal(t, tc_a.txt != tc_b.txt && idx_a < idx_b, a.LessThan(b), msg)
			assert.Equal(t, idx_a <= idx_b || tc_a.txt == tc_b.txt, a.LessThanOrEquals(b), msg)
			assert.Equal(t, a.Plus(b).String(), b.Plus(a).String(), msg)
			assert.Equal(t, a.Minus(b).String(), a.Plus(zero.Minus(b)).String(), msg)
			assert.Equal(t, a.Times(b).String(), b.Times(a).String(), msg)
			bIsNanOrZero := b.IsNaN() || b.Render(false, false, false) == "0"
			assert.Equal(t, turner(bIsNanOrZero, "NaN", tc_a.txt), a.Div(b).Times(b).String(), msg)
			assert.Equal(t, turner(bIsNanOrZero, "NaN", tc_a.txt), a.Times(b).Div(b).String(), msg)
			rat := a.Div(b)
			intrat := a.DivAndRoundTowardsZero(b)
			assert.Equal(t, rat.IsNaN(), intrat.IsNaN(), msg)
			if !rat.IsNaN() {
				if rat.IsInt() {
					assert.True(t, rat.Equals(intrat), msg)
				} else {
					if rat.LessThan(zero) {
						assert.True(t, rat.LessThan(intrat), msg)
						assert.True(t, intrat.Minus(one).LessThan(rat), msg)
					} else {
						assert.True(t, intrat.LessThan(rat), msg)
						assert.True(t, rat.LessThan(intrat.Plus(one)), msg)
					}
				}
			}
		}
		// Test string representation
		assert.Equal(t, tc_a.txt, a.String(), msg)
		assert.Equal(t, tc_a.txt, br.FromString(tc_a.txt).String(), msg)
	}

	// Test Pow
	assert.Equal(t, "1", br.FromString("1/3").Pow(zero).String())
	assert.Equal(t, "1/3", br.FromString("1/3").Pow(one).String())
	assert.Equal(t, "1/9", br.FromString("1/3").Pow(two).String())
	assert.Equal(t, "1/27", br.FromString("1/3").Pow(three).String())
	assert.Equal(t, "1", br.FromString("-1/3").Pow(zero).String())
	assert.Equal(t, "-1/3", br.FromString("-1/3").Pow(one).String())
	assert.Equal(t, "1/9", br.FromString("-1/3").Pow(two).String())
	assert.Equal(t, "-1/27", br.FromString("-1/3").Pow(three).String())
	assert.Equal(t, "1", zero.Pow(zero).String())
	assert.Equal(t, "0", zero.Pow(one).String())
	assert.Equal(t, "0", zero.Pow(two).String())

	// Test Binomial
	assert.Equal(t, "1", zero.Binomial(zero).String())
	assert.Equal(t, "1", one.Binomial(zero).String())
	assert.Equal(t, "1", one.Binomial(one).String())
	assert.Equal(t, "1", two.Binomial(zero).String())
	assert.Equal(t, "2", two.Binomial(one).String())
	assert.Equal(t, "1", two.Binomial(two).String())
	assert.Equal(t, "1", three.Binomial(zero).String())
	assert.Equal(t, "3", three.Binomial(one).String())
	assert.Equal(t, "3", three.Binomial(two).String())
	assert.Equal(t, "1", three.Binomial(three).String())

	// Test Render
	oneThird := br.FromString("1/3")
	negOneThird := br.FromString("-1/3")
	renderTestCases := []struct {
		bigrat  br.BR
		md      string
		mdB     string
		mdP     string
		mdPB    string
		latex   string
		latexB  string
		latexP  string
		latexPB string
	}{
		{nan1, "NaN", "**NaN**", "NaN", "**NaN**", "$NaN$", "$\\mathbf{NaN}$", "$NaN$", "$\\mathbf{NaN}$"},
		{negSevenThirds, "-2 1/3", "**-2 1/3**", "-233 1/3 %", "**-233 1/3 %**", "$-2\\frac{1}{3}$", "$\\mathbf{-2\\frac{1}{3}}$", "$-233\\frac{1}{3}\\%$", "$\\mathbf{-233\\frac{1}{3}\\%}$"},
		{negOne, "-1", "**-1**", "-100 %", "**-100 %**", "$-1$", "$\\mathbf{-1}$", "$-100\\%$", "$\\mathbf{-100\\%}$"},
		{negHalf, "-1/2", "**-1/2**", "-50 %", "**-50 %**", "$-\\frac{1}{2}$", "$\\mathbf{-\\frac{1}{2}}$", "$-50\\%$", "$\\mathbf{-50\\%}$"},
		{negOneThird, "-1/3", "**-1/3**", "-33 1/3 %", "**-33 1/3 %**", "$-\\frac{1}{3}$", "$\\mathbf{-\\frac{1}{3}}$", "$-33\\frac{1}{3}\\%$", "$\\mathbf{-33\\frac{1}{3}\\%}$"},
		{zero, "0", "**0**", "0 %", "**0 %**", "$0$", "$\\mathbf{0}$", "$0\\%$", "$\\mathbf{0\\%}$"},
		{oneThird, "1/3", "**1/3**", "33 1/3 %", "**33 1/3 %**", "$\\frac{1}{3}$", "$\\mathbf{\\frac{1}{3}}$", "$33\\frac{1}{3}\\%$", "$\\mathbf{33\\frac{1}{3}\\%}$"},
		{half, "1/2", "**1/2**", "50 %", "**50 %**", "$\\frac{1}{2}$", "$\\mathbf{\\frac{1}{2}}$", "$50\\%$", "$\\mathbf{50\\%}$"},
		{one, "1", "**1**", "100 %", "**100 %**", "$1$", "$\\mathbf{1}$", "$100\\%$", "$\\mathbf{100\\%}$"},
		{two, "2", "**2**", "200 %", "**200 %**", "$2$", "$\\mathbf{2}$", "$200\\%$", "$\\mathbf{200\\%}$"},
		{sevenThirds, "2 1/3", "**2 1/3**", "233 1/3 %", "**233 1/3 %**", "$2\\frac{1}{3}$", "$\\mathbf{2\\frac{1}{3}}$", "$233\\frac{1}{3}\\%$", "$\\mathbf{233\\frac{1}{3}\\%}$"},
		{three, "3", "**3**", "300 %", "**300 %**", "$3$", "$\\mathbf{3}$", "$300\\%$", "$\\mathbf{300\\%}$"},
	}
	for _, tc := range renderTestCases {
		assert.Equal(t, tc.md, tc.bigrat.Render(false, false, false))
		assert.Equal(t, tc.mdB, tc.bigrat.Render(false, false, true))
		assert.Equal(t, tc.mdP, tc.bigrat.Render(false, true, false))
		assert.Equal(t, tc.mdPB, tc.bigrat.Render(false, true, true))
		assert.Equal(t, tc.latex, tc.bigrat.Render(true, false, false))
		assert.Equal(t, tc.latexB, tc.bigrat.Render(true, false, true))
		assert.Equal(t, tc.latexP, tc.bigrat.Render(true, true, false))
		assert.Equal(t, tc.latexPB, tc.bigrat.Render(true, true, true))
	}
}

func turner(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
