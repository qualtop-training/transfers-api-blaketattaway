package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/karlseguin/ccache/v3"

	"transfers-api/internal/config"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/models"
)

type TransfersCCacheRepo struct {
	cache *ccache.Cache[models.Transfer]
	ttl   time.Duration
}

func NewTransfersCCacheRepository(cfg config.CCache) *TransfersCCacheRepo {
	cache := ccache.New(ccache.Configure[models.Transfer]().
		MaxSize(int64(cfg.MaxSize)).
		PercentToPrune(uint8(cfg.PercentToPrune)),
	)
	return &TransfersCCacheRepo{
		cache: cache,
		ttl:   time.Duration(cfg.TTLSeconds) * time.Second,
	}
}

func (r *TransfersCCacheRepo) Create(ctx context.Context, transfer models.Transfer) (string, error) {
	if transfer.ID == "" {
		return "", fmt.Errorf("transfer ID required for cache create")
	}

	r.cache.Set(transfer.ID, transfer, r.ttl)
	return transfer.ID, nil
}

func (r *TransfersCCacheRepo) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	item := r.cache.Get(id)
	if item == nil || item.Expired() {
		return models.Transfer{}, fmt.Errorf("transfer not found: %w", known_errors.ErrNotFound)
	}
	return item.Value(), nil
}

func (r *TransfersCCacheRepo) Update(ctx context.Context, transfer models.Transfer) error {
	_, err := r.Create(ctx, transfer)
	return err
}

func (r *TransfersCCacheRepo) Delete(ctx context.Context, id string) error {
	r.cache.Delete(id)
	return nil
}

func (r *TransfersCCacheRepo) GetAll(ctx context.Context) ([]models.Transfer, error) {
	return nil, fmt.Errorf("GetAll not implemented for ccache repository")
}

func (r *TransfersCCacheRepo) GetBySenderID(ctx context.Context, senderID string) ([]models.Transfer, error) {
	return nil, fmt.Errorf("GetBySenderID not implemented for ccache repository")
}

func (r *TransfersCCacheRepo) GetByReceiverID(ctx context.Context, receiverID string) ([]models.Transfer, error) {
	return nil, fmt.Errorf("GetByReceiverID not implemented for ccache repository")
}
