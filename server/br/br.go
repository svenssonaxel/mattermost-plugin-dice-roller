package br

import (
	"fmt"
	"math/big"
	"regexp"
)

// Big rationals implemented on top of math/big.Int, to provide
// - Immutable rational numbers
// - Markdown and LaTeX rendering in mixed numeral form
// - Improved performance over math/big.Rat by defering normalization until
//   needed.
// - Special NaN handling:
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
	status int
	num    *big.Int
	den    *big.Int
}

const (
	// r.status == statusNaN guarantees that r1.num and r1.den are null, and that
	// r1 represents NaN.
	// r.status == statusUnnormalized guarantees that r1.num and r1.den are not
	// null and not shared with other BRs, but that they might not be normalized.
	// r.status == statusNormalized guarantees that r1.num and r1.den are not
	// null, that they are normalized, but that they might be shared with other
	// BRs.
	statusNaN          = 0
	statusUnnormalized = 1
	statusNormalized   = 2
)

func New(n int) BR {
	return BR{
		status: statusNormalized,
		num:    big.NewInt(int64(n)),
		den:    oneBI,
	}
}

func FromBigInt(num, den *big.Int) BR {
	if den.Sign() == 0 {
		return Nan
	}
	return BR{
		status: statusUnnormalized,
		num:    new(big.Int).Set(num),
		den:    new(big.Int).Set(den),
	}
}

func (r1 BR) EnsureNormalized() {
	// Normalize if needed, destructively.
	if r1.status == statusUnnormalized {
		gcd := new(big.Int).GCD(nil, nil, r1.num, r1.den)
		r1.num.Div(r1.num, gcd)
		r1.den.Div(r1.den, gcd)
		if r1.den.Sign() == -1 {
			panic("unreachable: negative denominator")
		}
		r1.status = statusNormalized
	}
}

const normBitLimit = 1200

func (r1 BR) maybeNormalize() {
	if r1.status == statusUnnormalized && r1.num.BitLen() > normBitLimit && r1.den.BitLen() > normBitLimit {
		r1.EnsureNormalized()
	}
}

var Nan = BR{}
var negativeOne = New(-1)
var Zero = New(0)
var zeroBI = big.NewInt(0)
var One = New(1)
var oneBI = big.NewInt(1)
var ten = New(10)
var tenBI = big.NewInt(10)
var oneHundred = New(100)
var tenThousand = New(10000)
var hundredMillion = New(100000000)

func (r1 BR) IsNaN() bool {
	return r1.status == statusNaN
}

// If the rational is an integer, normalize it destructively and return true.
func (r1 BR) IsInt() bool {
	switch r1.status {
	case statusNaN:
		return false
	case statusUnnormalized:
		rem := new(big.Int).Mod(r1.num, r1.den)
		if rem.Sign() == 0 {
			// Since we happen to aleady have the gcd, we
			// might as well normalize.
			r1.num = new(big.Int).Div(r1.num, r1.den)
			r1.den = oneBI
			r1.status = statusNormalized
			return true
		}
		return false
	case statusNormalized:
		return r1.den.Cmp(oneBI) == 0
	}
	panic("unreachable")
}

// If the rational is a non-negative integer, normalize it destructively and
// return true.
func (r1 BR) InN0() bool {
	return r1.IsInt() && r1.num.Sign() != -1
}

// Arithmetic

func (r1 BR) Equals(r2 BR) bool {
	if !r1.IsNaN() && !r2.IsNaN() {
		p1 := new(big.Int).Mul(r1.num, r2.den)
		p2 := new(big.Int).Mul(r1.den, r2.num)
		return p1.Cmp(p2) == 0
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
	p1 := new(big.Int).Mul(r1.num, r2.den)
	p2 := new(big.Int).Mul(r1.den, r2.num)
	return p1.Cmp(p2) == -1
}

func (r1 BR) LessThanOrEquals(r2 BR) bool {
	return !r2.LessThan(r1)
}

func (r1 BR) Plus(r2 BR) BR {
	if r1.IsNaN() || r2.IsNaN() {
		return Nan
	}
	r1.maybeNormalize()
	r2.maybeNormalize()
	num := new(big.Int).Mul(r1.num, r2.den)
	den := new(big.Int).Mul(r1.den, r2.num)
	num.Add(num, den)
	den.Mul(r1.den, r2.den)
	return BR{
		status: statusUnnormalized,
		num:    num,
		den:    den,
	}
}

func (r1 BR) Minus(r2 BR) BR {
	if r1.IsNaN() || r2.IsNaN() {
		return Nan
	}
	r1.maybeNormalize()
	r2.maybeNormalize()
	num := new(big.Int).Mul(r1.num, r2.den)
	den := new(big.Int).Mul(r1.den, r2.num)
	num.Sub(num, den)
	den.Mul(r1.den, r2.den)
	return BR{
		status: statusUnnormalized,
		num:    num,
		den:    den,
	}
}

func (r1 BR) Times(r2 BR) BR {
	if r1.IsNaN() || r2.IsNaN() {
		return Nan
	}
	r1.maybeNormalize()
	r2.maybeNormalize()
	num := new(big.Int).Mul(r1.num, r2.num)
	den := new(big.Int).Mul(r1.den, r2.den)
	if den.Sign() == -1 {
		num.Neg(num)
		den.Neg(den)
	}
	return BR{
		status: statusUnnormalized,
		num:    num,
		den:    den,
	}
}

func (r1 BR) Div(r2 BR) BR {
	if r1.IsNaN() || r2.IsNaN() || r2.num.Sign() == 0 {
		return Nan
	}
	r1.maybeNormalize()
	r2.maybeNormalize()
	num := new(big.Int).Mul(r1.num, r2.den)
	den := new(big.Int).Mul(r1.den, r2.num)
	if den.Sign() == -1 {
		num.Neg(num)
		den.Neg(den)
	}
	return BR{
		status: statusUnnormalized,
		num:    num,
		den:    den,
	}
}

func (r1 BR) DivAndRoundTowardsZero(r2 BR) BR {
	r := r1.Div(r2)
	if r.IsNaN() || r.IsInt() {
		return r
	}
	r.num.Div(r.num, r.den)
	if r.num.Sign() == -1 {
		r.num.Add(r.num, oneBI)
	}
	r.den = oneBI
	r.status = statusNormalized
	return r
}

func (r1 BR) Pow(r2 BR) BR {
	if r1.IsNaN() || !r2.InN0() {
		return Nan
	}
	// It's worth normalizing r1, since the result might become big.
	r1.EnsureNormalized()
	// r2 is guaranteed to be normalized, due to the InN0 check above.
	num := new(big.Int).Exp(r1.num, r2.num, nil)
	den := new(big.Int).Exp(r1.den, r2.num, nil)
	return BR{
		// Since r1.num and r1.den are relatively prime, so must their powers be.
		status: statusNormalized,
		num:    num,
		den:    den,
	}
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
	// r1 and r2 are guaranteed to be normalized, due to the InN0 check above.
	num := new(big.Int).Binomial(r1.num.Int64(), r2.num.Int64())
	return BR{
		status: statusNormalized,
		num:    num,
		den:    oneBI,
	}
}

// String rendering and parsing

// Return a simple string representation of the rational number.
// - NaN is represented as "NaN".
// - Integers are represented as "-123".
// - Non-integers are represented as "-123/456".
func (r1 BR) String() string {
	if r1.IsNaN() {
		return "NaN"
	}
	r1.EnsureNormalized()
	if r1.IsInt() {
		return r1.num.String()
	}
	return fmt.Sprintf("%s/%s", r1.num.String(), r1.den.String())
}

// Return the rational number represented by the string argument, or NaN.
// Intended as an inverse of BR.String().
func FromString(s string) BR {
	// Match an integer or a rational number in the form "a/b". In order to avoid
	// CVE-2022-23772, we parse manually and use big.Int.SetString rather than
	// big.Rat.SetString.
	if m := regexp.MustCompile(`^(-?\d+)(?:/(\d+))?$`).FindStringSubmatch(s); m != nil {
		num, ok := new(big.Int).SetString(m[1], 10)
		if !ok {
			return Nan
		}
		if m[2] == "" {
			return BR{
				status: statusNormalized,
				num:    num,
				den:    oneBI,
			}
		}
		den, ok := new(big.Int).SetString(m[2], 10)
		if !ok {
			return Nan
		}
		return BR{
			status: statusUnnormalized,
			num:    num,
			den:    den,
		}
	}
	return Nan
}

// Return a traditional representation of the rational number, which depending
// on the numer is an integer, decimal, proper fraction or mixed numeral.
//
// The options parameter is a string of flags:
// "l": use Mattermost inline LaTeX syntax for rendering. This produces
// better-looking fractions. If not present, use Markdown syntax.
// "p": render as a percentage, e.g. multiply the number by 100 and append a
// percent sign.
// "b": render in bold-face.
// "i": render in italics.
func (r1 BR) Render(options string) string {
	r := r1
	// Parse options.
	var latex, percent, bold, italics bool
	for _, c := range options {
		switch c {
		case 'l':
			latex = true
		case 'p':
			percent = true
		case 'b':
			bold = true
		case 'i':
			italics = true
		}
	}
	// Render NaN.
	if r.IsNaN() {
		if latex {
			if bold && italics {
				return "$\\pmb{\\mathit{NaN}}$"
			}
			if bold {
				return "$\\mathbf{NaN}$"
			}
			if italics {
				return "$\\mathit{NaN}$"
			}
			return "$NaN$"
		}
		ret := "NaN"
		if bold {
			ret = "**" + ret + "**"
		}
		if italics {
			ret = "*" + ret + "*"
		}
		return ret
	}
	// Calculations
	r.EnsureNormalized()
	if percent {
		r = r.Times(oneHundred)
	}
	negative := false
	if r.num.Sign() == -1 {
		r = Zero.Minus(r)
		negative = true
	}
	if r.den.Sign() == -1 {
		panic("denominator is negative")
	}
	r.EnsureNormalized()
	num := r.num
	den := r.den
	intpart, rem := new(big.Int).DivMod(num, den, new(big.Int))
	mixedNum := rem
	mixedIntpart := intpart
	hasFracpart := mixedNum.Sign() != 0
	decimalStr := ""
	hasDecimals := false
	if hasFracpart {
		decimals := 0
		decpartBR := BR{
			status: statusNormalized,
			num:    mixedNum,
			den:    den,
		}
		for new(big.Int).GCD(nil, nil, decpartBR.den, tenBI).Cmp(oneBI) != 0 {
			decimals++
			decpartBR = decpartBR.Times(ten)
			decpartBR.EnsureNormalized()
		}
		if decpartBR.IsInt() {
			hasDecimals = true
			hasFracpart = false
			decimalStr = decpartBR.num.String()
			for len(decimalStr) < decimals {
				decimalStr = "0" + decimalStr
			}
			decimalStr = "." + decimalStr
		}
	}
	hasIntpart := mixedIntpart.Sign() != 0 || !hasFracpart
	// Render
	ret := ""
	if negative {
		ret = "-"
		if latex {
			if bold {
				ret = "\\textbf" + ret
			}
			ret = "\\textrm{" + ret + "}"
		}
	}
	if hasIntpart {
		ret += mixedIntpart.String()
		if hasFracpart && !latex {
			ret += " "
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
	if hasDecimals {
		ret += decimalStr
	}
	if percent {
		if latex {
			ret += "\\%"
		} else {
			ret += " %"
		}
	}
	if latex {
		switch {
		case bold && italics:
			ret = fmt.Sprintf("\\pmb{\\mathit{%s}}", ret)
		case bold:
			ret = fmt.Sprintf("\\mathbf{%s}", ret)
		case italics:
			ret = fmt.Sprintf("\\mathit{%s}", ret)
		}
		ret = fmt.Sprintf("$%s$", ret)
	} else {
		if bold {
			ret = fmt.Sprintf("**%s**", ret)
		}
		if italics {
			ret = fmt.Sprintf("*%s*", ret)
		}
	}
	return ret
}
