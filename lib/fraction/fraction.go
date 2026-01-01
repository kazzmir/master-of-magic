package fraction

import (
    "fmt"
    "encoding/json"
)

// represents exact fractions, instead of using floats

type Fraction struct {
    Numerator int
    Denominator int
}

func GCD(a, b int) int {
    for b != 0 {
        a, b = b, a % b
    }
    return a
}

func Zero() Fraction {
    return Fraction{}
}

func Make(numerator int, denominator int) Fraction {
    if denominator == 0 {
        // maybe panic?
        return Fraction{}
    }

    return Fraction{
        Numerator: numerator,
        Denominator: denominator,
    }.Reduce()
}

func FromInt(numerator int) Fraction {
    return Fraction{
        Numerator: numerator,
        Denominator: 1,
    }
}

func (fraction Fraction) MarshalJSON() ([]byte, error) {
    return json.Marshal(map[string]int{
        "n": fraction.Numerator,
        "d": fraction.Denominator,
    })
}

func (fraction Fraction) ToInt() int {
    return int(fraction.ToFloat())
}

func (fraction Fraction) ToFloat() float64 {
    if fraction.Numerator == 0 {
        return 0
    }
    return float64(fraction.Numerator) / float64(fraction.Denominator)
}

func (fraction Fraction) Reduce() Fraction {
    if fraction.IsZero(){
        return fraction
    }

    gcd := GCD(fraction.Numerator, fraction.Denominator)
    if gcd < 0 {
        gcd = -gcd
    }
    return Fraction{
        Numerator: fraction.Numerator / gcd,
        Denominator: fraction.Denominator / gcd,
    }
}

func (fraction Fraction) IsZero() bool {
    return fraction.Numerator == 0
}

func (fraction Fraction) Add(other Fraction) Fraction {
    if fraction.IsZero(){
        return other
    }
    if other.IsZero() {
        return fraction
    }
    return Fraction{
        Numerator: fraction.Numerator * other.Denominator + other.Numerator * fraction.Denominator,
        Denominator: fraction.Denominator * other.Denominator,
    }.Reduce()
}

func (fraction Fraction) Equals(other Fraction) bool {
    if fraction.IsZero() && other.IsZero() {
        return true
    }

    return fraction.Numerator == other.Numerator && fraction.Denominator == other.Denominator
}

func (fraction Fraction) Subtract(other Fraction) Fraction {
    if other.IsZero() {
        return fraction
    }

    if fraction.IsZero(){
        return other.Negate()
    }

    return Fraction{
        Numerator: fraction.Numerator * other.Denominator - other.Numerator * fraction.Denominator,
        Denominator: fraction.Denominator * other.Denominator,
    }.Reduce()
}

// return whichever fraction is larger
func (fraction Fraction) Max(other Fraction) Fraction {
    if fraction.GreaterThan(other) {
        return fraction
    }
    return other
}

func (fraction Fraction) Negate() Fraction {
    return Fraction{
        Numerator: -fraction.Numerator,
        Denominator: fraction.Denominator,
    }
}

func (fraction Fraction) Invert() Fraction {
    if fraction.IsZero() {
        return Fraction{}
    }

    return Fraction{
        Numerator: fraction.Denominator,
        Denominator: fraction.Numerator,
    }
}

func (fraction Fraction) Multiply(other Fraction) Fraction {
    if fraction.IsZero() || other.IsZero() {
        return Fraction{}
    }

    return Fraction{
        Numerator: fraction.Numerator * other.Numerator,
        Denominator: fraction.Denominator * other.Denominator,
    }.Reduce()
}

func (fraction Fraction) Divide(other Fraction) Fraction {
    return fraction.Multiply(other.Invert())
}

func (fraction Fraction) LessThanEqual(other Fraction) bool {
    return fraction.Equals(other) || fraction.LessThan(other)
}

func (fraction Fraction) GreaterThanEqual(other Fraction) bool {
    return fraction.Equals(other) || fraction.GreaterThan(other)
}

func (fraction Fraction) LessThan(other Fraction) bool {
    rest := other.Subtract(fraction)
    return rest.Numerator > 0
}

// returns the minimum of two fractions
func (fraction Fraction) Min(other Fraction) Fraction {
    if fraction.LessThan(other) {
        return fraction
    }
    return other
}

func (fraction Fraction) GreaterThan(other Fraction) bool {
    return !fraction.LessThanEqual(other)
}

func (fraction Fraction) String() string {
    return fmt.Sprintf("%v/%v", fraction.Numerator, fraction.Denominator)
}

// convert 3/2 into '1 1/2'
// 1/2 -> '1/2'
func (fraction Fraction) NormalString() string {
    if fraction.Numerator == 0 {
        return "0"
    }

    if fraction.Numerator > fraction.Denominator {
        v := fraction.Numerator / fraction.Denominator
        if fraction.Numerator % fraction.Denominator == 0 {
            return fmt.Sprintf("%v", v)
        }

        return fmt.Sprintf("%v %v/%v", v, fraction.Numerator % fraction.Denominator, fraction.Denominator)
    }

    return fmt.Sprintf("%v/%v", fraction.Numerator, fraction.Denominator)
}
