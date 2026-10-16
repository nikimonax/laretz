package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

type GetOperationUseCase struct {
	operations interfaces.OperationRepository
}

func NewGetOperationUseCase(operations interfaces.OperationRepository) *GetOperationUseCase {
	return &GetOperationUseCase{operations: operations}
}

func (usecase *GetOperationUseCase) Execute(ctx context.Context, id domain.OperationID) (domain.Operation, error) {
	return usecase.operations.Get(ctx, id)
}
