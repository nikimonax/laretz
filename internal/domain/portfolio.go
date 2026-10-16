package domain

import (
	"time"
)

type PortfolioID = int64

type Portfolio struct {
	ID        PortfolioID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
