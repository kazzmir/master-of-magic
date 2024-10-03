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

    if !Make(6, 11).Divide(Make(2, 3)).Equals(Make(9, 11)){
        test.Errorf("Expected %v to equal %v", Make(6, 11).Divide(Make(2, 3)), Make(9, 11))
    }

    if !Make(1, 2).LessThanEqual(Make(3, 4)){
        test.Errorf("Expected %v to be less than or equal to %v", Make(1, 2), Make(3, 4))
    }

    if !Make(1, 2).LessThanEqual(Make(1, 2)){
        test.Errorf("Expected %v to be less than or equal to %v", Make(1, 2), Make(1, 2))
    }

    if Make(1, 2).LessThanEqual(Make(1, 3)){
        test.Errorf("Expected %v to not be less than or equal to %v. Diff %v", Make(1, 2), Make(1, 3), Make(1, 3).Subtract(Make(1, 2)))
    }

    if !Make(-1, 2).LessThanEqual(FromInt(0)){
        test.Errorf("Expected %v to be less than or equal to %v", Make(-1, 2), FromInt(0))
    }

    if Make(1, 2).NormalString() != "1/2" {
        test.Errorf("Expected %v to equal '1/2'", Make(1, 2).NormalString())
    }

    if Make(3, 2).NormalString() != "1 1/2" {
        test.Errorf("Expected %v to equal '1 1/2'", Make(3, 2).NormalString())
    }

    if Make(5, 2).NormalString() != "2 1/2" {
        test.Errorf("Expected %v to equal '2 1/2'", Make(5, 2).NormalString())
    }
}
