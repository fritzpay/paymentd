package decimal

import (
	"code.google.com/p/godec/dec"
	"testing"
)

func TestDecimalParts(t *testing.T) {
	decimal := dec.NewDecInt64(123)
	decimal.SetScale(dec.Scale(2))
	mDec := &Decimal{Dec: *decimal}
	expect := "1.23"
	if mDec.String() != expect {
		t.Errorf("expect %s, got %s", expect, mDec.String())
		return
	}
	parts := mDec.Parts()
	expectLen := 2
	if len(parts) != expectLen {
		t.Errorf("expect parts len %d, got %d", expectLen, len(parts))
		return
	}
	expect = "1"
	if parts[0] != expect {
		t.Errorf("expect integer part %s, got %s", expect, parts[0])
		return
	}
	if mDec.IntegerPart() != expect {
		t.Errorf("expect integer part IntegerPart() %s, got %s", expect, mDec.IntegerPart())
		return
	}
	expect = "23"
	if parts[1] != expect {
		t.Errorf("expect decimal part %s, got %s", expect, parts[1])
	}
	if mDec.DecimalPart() != expect {
		t.Errorf("expect decimal part DecimalPart() %s, got %s", expect, mDec.DecimalPart())
	}
}

func TestDecimalPartsIntegerOnly(t *testing.T) {
	decimal := dec.NewDecInt64(123)
	mDec := &Decimal{Dec: *decimal}
	expect := "123"
	if mDec.IntegerPart() != expect {
		t.Errorf("expect integer part %s, got %s", expect, mDec.IntegerPart())
		return
	}
	expect = ""
	if mDec.DecimalPart() != expect {
		t.Errorf("expect empty decimal part, got %s", mDec.DecimalPart())
		return
	}
}
