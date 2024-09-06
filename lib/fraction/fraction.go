package fraction

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

func Make(numerator int, denominator int) Fraction {
    return Fraction{
        Numerator: numerator,
        Denominator: denominator,
    }.Reduce()
}

func (fraction Fraction) ToFloat() float64 {
    return float64(fraction.Numerator) / float64(fraction.Denominator)
}

func (fraction Fraction) Reduce() Fraction {
    if fraction.IsZero(){
        return fraction
    }

    gcd := GCD(fraction.Numerator, fraction.Denominator)
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

func (fraction Fraction) Negate() Fraction {
    return Fraction{
        Numerator: -fraction.Numerator,
        Denominator: fraction.Denominator,
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
