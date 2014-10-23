package decimal

import (
	"code.google.com/p/godec/dec"
	"strings"
)

type Decimal struct {
	*dec.Dec
}

func (d *Decimal) Parts() []string {
	return strings.Split(d.Dec.String(), ".")
}

func (d *Decimal) IntegerPart() string {
	return d.Parts()[0]
}

func (d *Decimal) DecimalPart() string {
	parts := d.Parts()
	if len(parts) <= 1 {
		return ""
	}
	return d.Parts()[1]
}
