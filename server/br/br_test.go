package br_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moussetc/mattermost-plugin-dice-roller/server/br"
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
		assert.Equal(t, ternaryStr(a.IsNaN(), "NaN", "0"), a.Minus(a).Render(""), msg)
		assert.Equal(t, tc_a.txt, a.Times(one).String(), msg)
		assert.Equal(t, ternaryStr(a.IsNaN(), "NaN", "0"), a.Times(zero).Render(""), msg)
		assert.Equal(t, a.Plus(a).String(), a.Times(two).String(), msg)
		assert.Equal(t, tc_a.txt, a.Div(one).String(), msg)
		assert.Equal(t, "NaN", a.Div(zero).String(), msg)
		assert.Equal(t, ternaryStr(a.IsNaN() || tc_a.txt == "0", "NaN", "1"), a.Div(a).Render(""), msg)
		for idx_b, tc_b := range testCases {
			b := tc_b.bigrat
			msg2 := msg + ", " + strconv.Itoa(idx_b)
			assert.Equal(t, tc_a.txt == tc_b.txt, a.Equals(b), msg2)
			assert.Equal(t, tc_a.txt != tc_b.txt && idx_a < idx_b, a.LessThan(b), msg2)
			assert.Equal(t, idx_a <= idx_b || tc_a.txt == tc_b.txt, a.LessThanOrEquals(b), msg2)
			assert.Equal(t, a.Plus(b).String(), b.Plus(a).String(), msg2)
			assert.Equal(t, a.Minus(b).String(), a.Plus(zero.Minus(b)).String(), msg2)
			assert.Equal(t, a.Times(b).String(), b.Times(a).String(), msg2)
			bIsNanOrZero := b.IsNaN() || b.Render("") == "0"
			assert.Equal(t, ternaryStr(bIsNanOrZero, "NaN", tc_a.txt), a.Div(b).Times(b).String(), msg2)
			assert.Equal(t, ternaryStr(bIsNanOrZero, "NaN", tc_a.txt), a.Times(b).Div(b).String(), msg2)
			rat := a.Div(b)
			intrat := a.DivAndRoundTowardsZero(b)
			assert.Equal(t, rat.IsNaN(), intrat.IsNaN(), msg2)
			if !rat.IsNaN() {
				assert.True(t, intrat.IsInt(), msg2)
				if rat.IsInt() {
					assert.True(t, rat.Equals(intrat), msg2)
				} else {
					if rat.LessThan(zero) {
						assert.True(t, rat.LessThan(intrat), msg2)
						assert.True(t, intrat.Minus(one).LessThan(rat), msg2)
					} else {
						assert.True(t, intrat.LessThan(rat), msg2)
						assert.True(t, rat.LessThan(intrat.Plus(one)), msg2)
					}
				}
			}
		}
		// Test string representation
		assert.Equal(t, tc_a.txt, a.String(), msg)
		assert.Equal(t, tc_a.txt, br.FromString(tc_a.txt).String(), msg)
	}

	// Test IsInt
	anotherOne := three.Div(three)
	assert.True(t, anotherOne.IsInt())
	assert.Equal(t, "1", anotherOne.String())

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
	negDec := br.FromString("-2345/1000")
	renderTestCases := []struct {
		options  string
		examples map[br.BR]string
	}{
		{
			"",
			map[br.BR]string{
				nan1:           "NaN",
				negDec:         "-2.345",
				negSevenThirds: "-2 1/3",
				negOne:         "-1",
				negHalf:        "-0.5",
				negOneThird:    "-1/3",
				zero:           "0",
				oneThird:       "1/3",
				half:           "0.5",
				one:            "1",
				two:            "2",
				sevenThirds:    "2 1/3",
				three:          "3",
			},
		},
		{
			"b",
			map[br.BR]string{
				nan1:           "**NaN**",
				negDec:         "**-2.345**",
				negSevenThirds: "**-2 1/3**",
				negOne:         "**-1**",
				negHalf:        "**-0.5**",
				negOneThird:    "**-1/3**",
				zero:           "**0**",
				oneThird:       "**1/3**",
				half:           "**0.5**",
				one:            "**1**",
				two:            "**2**",
				sevenThirds:    "**2 1/3**",
				three:          "**3**",
			},
		},
		{
			"i",
			map[br.BR]string{
				nan1:           "*NaN*",
				negDec:         "*-2.345*",
				negSevenThirds: "*-2 1/3*",
				negOne:         "*-1*",
				negHalf:        "*-0.5*",
				negOneThird:    "*-1/3*",
				zero:           "*0*",
				oneThird:       "*1/3*",
				half:           "*0.5*",
				one:            "*1*",
				two:            "*2*",
				sevenThirds:    "*2 1/3*",
				three:          "*3*",
			},
		},
		{
			"ib",
			map[br.BR]string{
				nan1:           "***NaN***",
				negDec:         "***-2.345***",
				negSevenThirds: "***-2 1/3***",
				negOne:         "***-1***",
				negHalf:        "***-0.5***",
				negOneThird:    "***-1/3***",
				zero:           "***0***",
				oneThird:       "***1/3***",
				half:           "***0.5***",
				one:            "***1***",
				two:            "***2***",
				sevenThirds:    "***2 1/3***",
				three:          "***3***",
			},
		},
		{
			"p",
			map[br.BR]string{
				nan1:           "NaN",
				negDec:         "-234.5 %",
				negSevenThirds: "-233 1/3 %",
				negOne:         "-100 %",
				negHalf:        "-50 %",
				negOneThird:    "-33 1/3 %",
				zero:           "0 %",
				oneThird:       "33 1/3 %",
				half:           "50 %",
				one:            "100 %",
				two:            "200 %",
				sevenThirds:    "233 1/3 %",
				three:          "300 %",
			},
		},
		{
			"pb",
			map[br.BR]string{
				nan1:           "**NaN**",
				negDec:         "**-234.5 %**",
				negSevenThirds: "**-233 1/3 %**",
				negOne:         "**-100 %**",
				negHalf:        "**-50 %**",
				negOneThird:    "**-33 1/3 %**",
				zero:           "**0 %**",
				oneThird:       "**33 1/3 %**",
				half:           "**50 %**",
				one:            "**100 %**",
				two:            "**200 %**",
				sevenThirds:    "**233 1/3 %**",
				three:          "**300 %**",
			},
		},
		{
			"pi",
			map[br.BR]string{
				nan1:           "*NaN*",
				negDec:         "*-234.5 %*",
				negSevenThirds: "*-233 1/3 %*",
				negOne:         "*-100 %*",
				negHalf:        "*-50 %*",
				negOneThird:    "*-33 1/3 %*",
				zero:           "*0 %*",
				oneThird:       "*33 1/3 %*",
				half:           "*50 %*",
				one:            "*100 %*",
				two:            "*200 %*",
				sevenThirds:    "*233 1/3 %*",
				three:          "*300 %*",
			},
		},
		{
			"pib",
			map[br.BR]string{
				nan1:           "***NaN***",
				negDec:         "***-234.5 %***",
				negSevenThirds: "***-233 1/3 %***",
				negOne:         "***-100 %***",
				negHalf:        "***-50 %***",
				negOneThird:    "***-33 1/3 %***",
				zero:           "***0 %***",
				oneThird:       "***33 1/3 %***",
				half:           "***50 %***",
				one:            "***100 %***",
				two:            "***200 %***",
				sevenThirds:    "***233 1/3 %***",
				three:          "***300 %***",
			},
		},
		{
			"l",
			map[br.BR]string{
				nan1:           "$NaN$",
				negDec:         "$\\textrm{-}2.345$",
				negSevenThirds: "$\\textrm{-}2\\frac{1}{3}$",
				negOne:         "$\\textrm{-}1$",
				negHalf:        "$\\textrm{-}0.5$",
				negOneThird:    "$\\textrm{-}\\frac{1}{3}$",
				zero:           "$0$",
				oneThird:       "$\\frac{1}{3}$",
				half:           "$0.5$",
				one:            "$1$",
				two:            "$2$",
				sevenThirds:    "$2\\frac{1}{3}$",
				three:          "$3$",
			},
		},
		{
			"lb",
			map[br.BR]string{
				nan1:           "$\\mathbf{NaN}$",
				negDec:         "$\\mathbf{\\textrm{\\textbf-}2.345}$",
				negSevenThirds: "$\\mathbf{\\textrm{\\textbf-}2\\frac{1}{3}}$",
				negOne:         "$\\mathbf{\\textrm{\\textbf-}1}$",
				negHalf:        "$\\mathbf{\\textrm{\\textbf-}0.5}$",
				negOneThird:    "$\\mathbf{\\textrm{\\textbf-}\\frac{1}{3}}$",
				zero:           "$\\mathbf{0}$",
				oneThird:       "$\\mathbf{\\frac{1}{3}}$",
				half:           "$\\mathbf{0.5}$",
				one:            "$\\mathbf{1}$",
				two:            "$\\mathbf{2}$",
				sevenThirds:    "$\\mathbf{2\\frac{1}{3}}$",
				three:          "$\\mathbf{3}$",
			},
		},
		{
			"li",
			map[br.BR]string{
				nan1:           "$\\mathit{NaN}$",
				negDec:         "$\\mathit{\\textrm{-}2.345}$",
				negSevenThirds: "$\\mathit{\\textrm{-}2\\frac{1}{3}}$",
				negOne:         "$\\mathit{\\textrm{-}1}$",
				negHalf:        "$\\mathit{\\textrm{-}0.5}$",
				negOneThird:    "$\\mathit{\\textrm{-}\\frac{1}{3}}$",
				zero:           "$\\mathit{0}$",
				oneThird:       "$\\mathit{\\frac{1}{3}}$",
				half:           "$\\mathit{0.5}$",
				one:            "$\\mathit{1}$",
				two:            "$\\mathit{2}$",
				sevenThirds:    "$\\mathit{2\\frac{1}{3}}$",
				three:          "$\\mathit{3}$",
			},
		},
		{
			"lib",
			map[br.BR]string{
				nan1:           "$\\pmb{\\mathit{NaN}}$",
				negDec:         "$\\pmb{\\mathit{\\textrm{\\textbf-}2.345}}$",
				negSevenThirds: "$\\pmb{\\mathit{\\textrm{\\textbf-}2\\frac{1}{3}}}$",
				negOne:         "$\\pmb{\\mathit{\\textrm{\\textbf-}1}}$",
				negHalf:        "$\\pmb{\\mathit{\\textrm{\\textbf-}0.5}}$",
				negOneThird:    "$\\pmb{\\mathit{\\textrm{\\textbf-}\\frac{1}{3}}}$",
				zero:           "$\\pmb{\\mathit{0}}$",
				oneThird:       "$\\pmb{\\mathit{\\frac{1}{3}}}$",
				half:           "$\\pmb{\\mathit{0.5}}$",
				one:            "$\\pmb{\\mathit{1}}$",
				two:            "$\\pmb{\\mathit{2}}$",
				sevenThirds:    "$\\pmb{\\mathit{2\\frac{1}{3}}}$",
				three:          "$\\pmb{\\mathit{3}}$",
			},
		},
		{
			"lp",
			map[br.BR]string{
				nan1:           "$NaN$",
				negDec:         "$\\textrm{-}234.5\\%$",
				negSevenThirds: "$\\textrm{-}233\\frac{1}{3}\\%$",
				negOne:         "$\\textrm{-}100\\%$",
				negHalf:        "$\\textrm{-}50\\%$",
				negOneThird:    "$\\textrm{-}33\\frac{1}{3}\\%$",
				zero:           "$0\\%$",
				oneThird:       "$33\\frac{1}{3}\\%$",
				half:           "$50\\%$",
				one:            "$100\\%$",
				two:            "$200\\%$",
				sevenThirds:    "$233\\frac{1}{3}\\%$",
				three:          "$300\\%$",
			},
		},
		{
			"lpb",
			map[br.BR]string{
				nan1:           "$\\mathbf{NaN}$",
				negDec:         "$\\mathbf{\\textrm{\\textbf-}234.5\\%}$",
				negSevenThirds: "$\\mathbf{\\textrm{\\textbf-}233\\frac{1}{3}\\%}$",
				negOne:         "$\\mathbf{\\textrm{\\textbf-}100\\%}$",
				negHalf:        "$\\mathbf{\\textrm{\\textbf-}50\\%}$",
				negOneThird:    "$\\mathbf{\\textrm{\\textbf-}33\\frac{1}{3}\\%}$",
				zero:           "$\\mathbf{0\\%}$",
				oneThird:       "$\\mathbf{33\\frac{1}{3}\\%}$",
				half:           "$\\mathbf{50\\%}$",
				one:            "$\\mathbf{100\\%}$",
				two:            "$\\mathbf{200\\%}$",
				sevenThirds:    "$\\mathbf{233\\frac{1}{3}\\%}$",
				three:          "$\\mathbf{300\\%}$",
			},
		},
		{
			"lpi",
			map[br.BR]string{
				nan1:           "$\\mathit{NaN}$",
				negDec:         "$\\mathit{\\textrm{-}234.5\\%}$",
				negSevenThirds: "$\\mathit{\\textrm{-}233\\frac{1}{3}\\%}$",
				negOne:         "$\\mathit{\\textrm{-}100\\%}$",
				negHalf:        "$\\mathit{\\textrm{-}50\\%}$",
				negOneThird:    "$\\mathit{\\textrm{-}33\\frac{1}{3}\\%}$",
				zero:           "$\\mathit{0\\%}$",
				oneThird:       "$\\mathit{33\\frac{1}{3}\\%}$",
				half:           "$\\mathit{50\\%}$",
				one:            "$\\mathit{100\\%}$",
				two:            "$\\mathit{200\\%}$",
				sevenThirds:    "$\\mathit{233\\frac{1}{3}\\%}$",
				three:          "$\\mathit{300\\%}$",
			},
		},
		{
			"lpib",
			map[br.BR]string{
				nan1:           "$\\pmb{\\mathit{NaN}}$",
				negDec:         "$\\pmb{\\mathit{\\textrm{\\textbf-}234.5\\%}}$",
				negSevenThirds: "$\\pmb{\\mathit{\\textrm{\\textbf-}233\\frac{1}{3}\\%}}$",
				negOne:         "$\\pmb{\\mathit{\\textrm{\\textbf-}100\\%}}$",
				negHalf:        "$\\pmb{\\mathit{\\textrm{\\textbf-}50\\%}}$",
				negOneThird:    "$\\pmb{\\mathit{\\textrm{\\textbf-}33\\frac{1}{3}\\%}}$",
				zero:           "$\\pmb{\\mathit{0\\%}}$",
				oneThird:       "$\\pmb{\\mathit{33\\frac{1}{3}\\%}}$",
				half:           "$\\pmb{\\mathit{50\\%}}$",
				one:            "$\\pmb{\\mathit{100\\%}}$",
				two:            "$\\pmb{\\mathit{200\\%}}$",
				sevenThirds:    "$\\pmb{\\mathit{233\\frac{1}{3}\\%}}$",
				three:          "$\\pmb{\\mathit{300\\%}}$",
			},
		},
	}
	for _, tc := range renderTestCases {
		options := tc.options
		for br, result := range tc.examples {
			assert.Equal(t, result, br.Render(options))
		}
	}
}

func ternaryStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
