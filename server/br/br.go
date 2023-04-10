package br

import (
	"fmt"
	"math/big"
)

// Wrapper around math/big.Rat, to provide
// - immutable rational numbers
// - Markdown and LaTeX rendering in mixed numeral form
// - special NaN handling:
//   - NaN is rendered as "NaN".
//   - NaN is propagated through all operations.
//   - NaN equals only itself.
//   - NaN is not an integer.
//   - NaN sorts first.
//   - NaN is produced by:
//     - arithmetic operations with NaN operands.
//     - division by zero.
//     - exponentiation with negative or non-integer exponent.
//     - parsing a string that does not represent a rational number.

type BR struct {
	r *big.Rat
}

func New(n int) BR {
	return BR{big.NewRat(int64(n), 1)}
}

var Nan = BR{nil}
var Zero = New(0)
var One = New(1)
var oneHundred = New(100)
var tenThousand = New(10000)

func (r BR) IsNaN() bool {
	return r.r == nil
}

func (r BR) IsInt() bool {
	if r.IsNaN() {
		return false
	}
	return r.r.IsInt()
}

func (r BR) InN0() bool {
	return r.IsInt() && r.r.Sign() != -1
}

// Arithmetic

func (r1 BR) Equals(r2 BR) bool {
	if !r1.IsNaN() && !r2.IsNaN() {
		return r1.r.Cmp(r2.r) == 0
	}
	return r1.IsNaN() == r2.IsNaN()
}

func (r1 BR) LessThan(r2 BR) bool {
	if r2.IsNaN() {
		return false
	}
	if r1.IsNaN() {
		return true
	}
	return r1.r.Cmp(r2.r) == -1
}

func (r1 BR) LessThanOrEquals(r2 BR) bool {
	return r1.LessThan(r2) || r1.Equals(r2)
}

func (r1 BR) Plus(r2 BR) BR {
	if r1.IsNaN() || r2.IsNaN() {
		return Nan
	}
	return BR{new(big.Rat).Add(r1.r, r2.r)}
}

func (r1 BR) Minus(r2 BR) BR {
	if r1.IsNaN() || r2.IsNaN() {
		return Nan
	}
	return BR{new(big.Rat).Sub(r1.r, r2.r)}
}

func (r1 BR) Times(r2 BR) BR {
	if r1.IsNaN() || r2.IsNaN() {
		return Nan
	}
	return BR{new(big.Rat).Mul(r1.r, r2.r)}
}

func (r1 BR) Div(r2 BR) BR {
	if r1.IsNaN() || r2.IsNaN() || r2.r.Sign() == 0 {
		return Nan
	}
	return BR{new(big.Rat).Quo(r1.r, r2.r)}
}

func (r1 BR) DivAndRoundTowardsZero(r2 BR) BR {
	r := r1.Div(r2)
	if r.IsNaN() || r.IsInt() {
		return r
	}
	intpart, _ := new(big.Int).DivMod(r.r.Num(), r.r.Denom(), big.NewInt(1))
	ret := BR{new(big.Rat).SetInt(intpart)}
	if intpart.Sign() == -1 {
		ret = ret.Plus(One)
	}
	return ret
}

func (r1 BR) Pow(r2 BR) BR {
	if r1.IsNaN() || !r2.InN0() {
		return Nan
	}
	num := r1.r.Num()
	den := r1.r.Denom()
	exp := r2.r.Num()
	num = new(big.Int).Exp(num, exp, nil)
	den = new(big.Int).Exp(den, exp, nil)
	return BR{new(big.Rat).SetFrac(num, den)}
}

func (r1 BR) Binomial(r2 BR) BR {
	if !r1.InN0() || !r2.InN0() {
		return Nan
	}
	if r1.LessThan(r2) {
		return Nan
	}
	if tenThousand.LessThan(r1) {
		return Nan
	}
	num := new(big.Int).Binomial(r1.r.Num().Int64(), r2.r.Num().Int64())
	return BR{new(big.Rat).SetInt(num)}
}

// String rendering and parsing

// Return a simple string representation of the rational number.
// - NaN is represented as "NaN".
// - Integers are represented as "-123".
// - Non-integers are represented as "-123/456".
func (r BR) String() string {
	if r.IsNaN() {
		return "NaN"
	}
	if r.r.IsInt() {
		return r.r.Num().String()
	}
	return r.r.String()
}

// Return the rational number represented by the string argument, or NaN.
// Intended as an inverse of BR.String().
func FromString(s string) BR {
	// Match an integer or a rational number in the form "a/b"
	// matches := regexp.MustCompile(`^(0|-?[1-9][0-9]*)(/[1-9][0-9]*)?$`).FindStringSubmatch(s)
	// if matches != nil {
	// 	r, ok := new(big.Rat).SetString(s)
	// 	if !ok {
	// 		return Nan
	// 	}
	// 	return BR{r}
	// }
	r, ok := new(big.Rat).SetString(s)
	if ok {
		return BR{r}
	}
	return Nan
}

// Return a traditional representation of the rational number, which depending
// on the numer is an integer, proper fraction or mixed numeral representation.
// - latex: if true, use inline LaTeX syntax for rendering. This produces
//   better-looking fractions.
// - percent: if true, multiply the number by 100 and render it as a percentage.
// - bold: if true, render the number in bold-face, using either LaTeX or
//   Markdown syntax.
func (r BR) Render(latex, percent, bold bool) string {
	if r.IsNaN() {
		if latex && bold {
			return "$\\mathbf{NaN}$"
		}
		if latex {
			return "$NaN$"
		}
		if bold {
			return "**NaN**"
		}
		return "NaN"
	}
	if percent {
		r = r.Times(oneHundred)
	}
	num := r.r.Num()
	den := r.r.Denom()
	intpart, rem := new(big.Int).DivMod(num, den, big.NewInt(1))
	mixedNum := rem
	mixedIntpart := intpart
	if intpart.Sign() == -1 && rem.Sign() != 0 {
		mixedIntpart = new(big.Int).Add(intpart, big.NewInt(1))
		mixedNum = new(big.Int).Sub(den, rem)
	}
	hasFracpart := mixedNum.Sign() != 0
	hasIntpart := mixedIntpart.Sign() != 0 || !hasFracpart
	ret := ""
	if hasIntpart {
		ret += mixedIntpart.String()
		if hasFracpart && !latex {
			ret += " "
		}
	} else {
		if intpart.Sign() == -1 {
			ret += "-"
		}
	}
	if hasFracpart {
		if latex {
			ret += "\\frac{"
		}
		ret += mixedNum.String()
		if latex {
			ret += "}{"
		} else {
			ret += "/"
		}
		ret += den.String()
		if latex {
			ret += "}"
		}
	}
	if percent {
		if latex {
			ret += "\\%"
		} else {
			ret += " %"
		}
	}
	if bold {
		if latex {
			ret = fmt.Sprintf("\\mathbf{%s}", ret)
		} else {
			ret = fmt.Sprintf("**%s**", ret)
		}
	}
	if latex {
		ret = fmt.Sprintf("$%s$", ret)
	}
	return ret
}
