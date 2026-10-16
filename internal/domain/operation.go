package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type OperationID = int64
type OperationType string

const (
	OperationBuy  OperationType = "buy"
	OperationSell OperationType = "sell"
)

type Operation struct {
	ID          OperationID
	PortfolioID PortfolioID
	AssetID     AssetID
	Type        OperationType
	ExecutedAt  time.Time
	Quantity    decimal.Decimal
	TotalUSD    decimal.Decimal
	FeeAmount   *decimal.Decimal
	FeeAssetID  *AssetID
	Note        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
