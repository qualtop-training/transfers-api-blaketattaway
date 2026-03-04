package repositories

import (
	"context"
	"transfers-api/internal/config"
	"transfers-api/internal/models"
)

type TransfersMySQLRepo struct {
}

func NewTransfersMySQLRepository(cfg config.MySQLDB) *TransfersMySQLRepo {
	return &TransfersMySQLRepo{}
}

func (r *TransfersMySQLRepo) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	return "", nil
}

func (r *TransfersMySQLRepo) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	return models.Transfer{}, nil
}

func (r *TransfersMySQLRepo) Update(ctx context.Context, transfer models.Transfer) error {
	return nil
}

func (r *TransfersMySQLRepo) Delete(ctx context.Context, id string) error {
	return nil
}
