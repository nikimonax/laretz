package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type DeleteOperation struct {
	ID domain.OperationID
}

type DeleteOperationUseCase struct {
	operations interfaces.OperationRepository
}

func NewDeleteOperationUseCase(operations interfaces.OperationRepository) *DeleteOperationUseCase {
	return &DeleteOperationUseCase{operations: operations}
}

func (usecase *DeleteOperationUseCase) Execute(ctx context.Context, cmd DeleteOperation) error {
	return usecase.operations.Delete(ctx, cmd.ID)
}
