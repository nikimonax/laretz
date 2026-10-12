package pages

import (
	"github.com/shopspring/decimal"
)

// money renders an amount in the Quote Currency for display, rounded to
// cents.
func money(amount decimal.Decimal) string {
	return amount.StringFixed(2)
}
