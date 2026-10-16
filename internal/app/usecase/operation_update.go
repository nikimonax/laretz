package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type UpdateOperation struct {
	CreateOperation
	ID domain.OperationID
}

type UpdateOperationUseCase struct {
	operations interfaces.OperationRepository
}

func NewUpdateOperationUseCase(operations interfaces.OperationRepository) *UpdateOperationUseCase {
	return &UpdateOperationUseCase{operations: operations}
}

func (usecase *UpdateOperationUseCase) Execute(ctx context.Context, cmd UpdateOperation) error {
	operation := domain.Operation{
		ID:          cmd.ID,
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

	return usecase.operations.Update(ctx, operation)
}
