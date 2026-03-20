package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bradfitz/gomemcache/memcache"

	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/models"
)

type TransfersMemcachedRepo struct {
	client     *memcache.Client
	ttlSeconds int32
}

type transferCacheDAO struct {
	ID         string  `json:"id"`
	SenderID   string  `json:"sender_id"`
	ReceiverID string  `json:"receiver_id"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	State      string  `json:"state"`
}

func NewTransfersMemcachedRepository(cfg config.Memcached) *TransfersMemcachedRepo {
	client := memcache.New(fmt.Sprintf("%s:%d", cfg.Hostname, cfg.Port))
	return &TransfersMemcachedRepo{client: client, ttlSeconds: int32(cfg.TTLSeconds)}
}

func (r *TransfersMemcachedRepo) Create(ctx context.Context, transfer models.Transfer) (string, error) {

	if transfer.ID == "" {
		return "", fmt.Errorf("transfer ID required for cache create")
	}

	dao := transferCacheDAO{
		ID:         transfer.ID,
		SenderID:   transfer.SenderID,
		ReceiverID: transfer.ReceiverID,
		Currency:   transfer.Currency.String(),
		Amount:     transfer.Amount,
		State:      string(transfer.State),
	}

	data, err := json.Marshal(dao)
	if err != nil {
		return "", fmt.Errorf("error marshalling transfer for cache: %v", err)
	}

	item := &memcache.Item{
		Key:        transfer.ID,
		Value:      data,
		Expiration: r.ttlSeconds,
	}

	err = r.client.Set(item)
	if err != nil {
		return "", fmt.Errorf("error setting transfer in cache: %v", err)
	}

	return transfer.ID, nil
}

func (r *TransfersMemcachedRepo) GetByID(ctx context.Context, id string) (models.Transfer, error) {
	item, err := r.client.Get(id)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return models.Transfer{}, fmt.Errorf("transfer not found: %w", known_errors.ErrNotFound)
		}
		return models.Transfer{}, fmt.Errorf("error getting transfer from cache: %v", err)
	}

	var dao transferCacheDAO
	err = json.Unmarshal(item.Value, &dao)
	if err != nil {
		return models.Transfer{}, fmt.Errorf("error unmarshalling transfer from cache: %v", err)
	}

	return models.Transfer{
		ID:         dao.ID,
		SenderID:   dao.SenderID,
		ReceiverID: dao.ReceiverID,
		Currency:   enums.ParseCurrency(dao.Currency),
		Amount:     dao.Amount,
		State:      dao.State,
	}, nil
}

func (r *TransfersMemcachedRepo) Update(ctx context.Context, transfer models.Transfer) error {
	// Overwriting the cache entry
	_, err := r.Create(ctx, transfer)
	return err
}

func (r *TransfersMemcachedRepo) Delete(ctx context.Context, id string) error {
	err := r.client.Delete(id)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return fmt.Errorf("transfer not found: %w", known_errors.ErrNotFound)
		}
		return fmt.Errorf("error deleting transfer from cache: %v", err)
	}
	return nil
}

func (r *TransfersMemcachedRepo) GetAll(ctx context.Context) ([]models.Transfer, error) {
	return nil, fmt.Errorf("GetAll not implemented for Memcached repository")
}

func (r *TransfersMemcachedRepo) GetBySenderID(ctx context.Context, senderID string) ([]models.Transfer, error) {
	return nil, fmt.Errorf("GetBySenderID not implemented for Memcached repository")
}

func (r *TransfersMemcachedRepo) GetByReceiverID(ctx context.Context, receiverID string) ([]models.Transfer, error) {
	return nil, fmt.Errorf("GetByReceiverID not implemented for Memcached repository")
}
