package fraction

import (
    "testing"
)

func TestBasic(test *testing.T){
    fraction1 := Fraction{Numerator: 1, Denominator: 2}
    fraction2 := Fraction{Numerator: 1, Denominator: 2}

    if !fraction1.Equals(fraction2){
        test.Errorf("Expected %v to equal %v", fraction1, fraction2)
    }

    if !fraction1.Add(fraction2).Equals(Fraction{Numerator: 1, Denominator: 1}){
        test.Errorf("Expected %v to equal %v", fraction1.Add(fraction2), Fraction{Numerator: 1, Denominator: 1})
    }
}

func TestMore(test *testing.T){
    f1 := Make(2, 3)
    f2 := Make(3, 5)

    f3 := f1.Multiply(f2)

    if !f3.Equals(Make(2, 5)){
        test.Errorf("Expected %v to equal %v", f3, Make(2, 5))
    }

    if !f1.Multiply(Make(0, 0)).IsZero() {
        test.Errorf("Expected %v to be zero", f1.Multiply(Make(0, 0)))
    }
}
