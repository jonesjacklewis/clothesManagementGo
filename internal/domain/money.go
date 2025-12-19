package domain

import (
	"strings"

	"github.com/dustin/go-humanize"
)

type Pence int64

func (p Pence) AsPounds() string {

	isNegative := p < 0

	if isNegative {
		p = -p
	}

	poundsValue := float64(p) / 100.0

	poundString := "£" + humanize.CommafWithDigits(poundsValue, 2)

	if !strings.Contains(poundString, ".") {
		poundString = poundString + ".00"
	}

	if isNegative {
		poundString = "-" + poundString
	}

	return poundString
}
