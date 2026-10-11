package domain

import (
	"github.com/shopspring/decimal"
)

type Position struct {
	Asset   Asset
	Balance decimal.Decimal
}
