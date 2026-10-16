package usecase

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type CreateOperation struct {
	PortfolioID domain.PortfolioID
	AssetID     domain.AssetID
	Type        domain.OperationType
	ExecutedAt  time.Time
	Quantity    decimal.Decimal
	TotalUSD    decimal.Decimal
	FeeAmount   *decimal.Decimal
	FeeAssetID  *domain.AssetID
	Note        string
}

type CreateOperationUseCase struct {
	operations interfaces.OperationRepository
}

func NewCreateOperationUseCase(operations interfaces.OperationRepository) *CreateOperationUseCase {
	return &CreateOperationUseCase{operations: operations}
}

func (usecase *CreateOperationUseCase) Execute(ctx context.Context, cmd CreateOperation) error {
	operation := domain.Operation{
		PortfolioID: cmd.PortfolioID,
		AssetID:     cmd.AssetID,
		Type:        cmd.Type,
		ExecutedAt:  cmd.ExecutedAt,
		Quantity:    cmd.Quantity,
		TotalUSD:    cmd.TotalUSD,
		FeeAmount:   cmd.FeeAmount,
		FeeAssetID:  cmd.FeeAssetID,
		Note:        cmd.Note,
	}

	return usecase.operations.Create(ctx, operation)
}
