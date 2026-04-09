package services

import (
	"context"
	"errors"
	"testing"
	"time"
	"transfers-api/internal/config"
	"transfers-api/internal/enums"
	"transfers-api/internal/known_errors"
	"transfers-api/internal/models"
	"transfers-api/internal/services/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestService() (*TransfersService, *mocks.TransfersRepositoryMock, *mocks.TransfersRepositoryMock, *mocks.TransfersPublisherMock) {
	repo := new(mocks.TransfersRepositoryMock)
	cache := new(mocks.TransfersRepositoryMock)
	publisher := new(mocks.TransfersPublisherMock)
	svc := NewTransfersService(config.BusinessConfig{TransferMinAmount: 1}, repo, cache, publisher)
	return svc, repo, cache, publisher
}

func validTransfer() models.Transfer {
	return models.Transfer{
		SenderID:   "sender-1",
		ReceiverID: "receiver-1",
		Currency:   enums.CurrencyUSD,
		Amount:     100.0,
		State:      "pending",
	}
}

// ==================== Create ====================

func TestCreate_Success(t *testing.T) {
	svc, repo, cache, publisher := newTestService()
	transfer := validTransfer()

	repo.On("Create", mock.Anything, transfer).Return("id-123", nil)
	cache.On("Create", mock.Anything, mock.Anything).Return("id-123", nil)
	publisher.On("Publish", "created", "id-123").Return(nil)

	id, err := svc.Create(context.Background(), transfer)

	assert.NoError(t, err)
	assert.Equal(t, "id-123", id)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)

	// wait for goroutine to finish
	time.Sleep(50 * time.Millisecond)
	publisher.AssertExpectations(t)
}

func TestCreate_EmptySenderID(t *testing.T) {
	svc, _, _, _ := newTestService()
	transfer := validTransfer()
	transfer.SenderID = ""

	id, err := svc.Create(context.Background(), transfer)

	assert.Empty(t, id)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, known_errors.ErrBadRequest))
}

func TestCreate_EmptyReceiverID(t *testing.T) {
	svc, _, _, _ := newTestService()
	transfer := validTransfer()
	transfer.ReceiverID = ""

	id, err := svc.Create(context.Background(), transfer)

	assert.Empty(t, id)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, known_errors.ErrBadRequest))
}

func TestCreate_UnknownCurrency(t *testing.T) {
	svc, _, _, _ := newTestService()
	transfer := validTransfer()
	transfer.Currency = enums.CurrencyUnknown

	id, err := svc.Create(context.Background(), transfer)

	assert.Empty(t, id)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, known_errors.ErrBadRequest))
}

func TestCreate_ZeroAmount(t *testing.T) {
	svc, _, _, _ := newTestService()
	transfer := validTransfer()
	transfer.Amount = 0

	id, err := svc.Create(context.Background(), transfer)

	assert.Empty(t, id)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, known_errors.ErrBadRequest))
}

func TestCreate_NegativeAmount(t *testing.T) {
	svc, _, _, _ := newTestService()
	transfer := validTransfer()
	transfer.Amount = -50

	id, err := svc.Create(context.Background(), transfer)

	assert.Empty(t, id)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, known_errors.ErrBadRequest))
}

func TestCreate_EmptyState(t *testing.T) {
	svc, _, _, _ := newTestService()
	transfer := validTransfer()
	transfer.State = ""

	id, err := svc.Create(context.Background(), transfer)

	assert.Empty(t, id)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, known_errors.ErrBadRequest))
}

func TestCreate_RepoError(t *testing.T) {
	svc, repo, _, _ := newTestService()
	transfer := validTransfer()
	repoErr := errors.New("db connection failed")

	repo.On("Create", mock.Anything, transfer).Return("", repoErr)

	id, err := svc.Create(context.Background(), transfer)

	assert.Empty(t, id)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
	repo.AssertExpectations(t)
}

func TestCreate_CacheError_StillSucceeds(t *testing.T) {
	svc, repo, cache, publisher := newTestService()
	transfer := validTransfer()

	repo.On("Create", mock.Anything, transfer).Return("id-123", nil)
	cache.On("Create", mock.Anything, mock.Anything).Return("", errors.New("cache down"))
	publisher.On("Publish", "created", "id-123").Return(nil)

	id, err := svc.Create(context.Background(), transfer)

	assert.NoError(t, err)
	assert.Equal(t, "id-123", id)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// ==================== GetByID ====================

func TestGetByID_CacheHit(t *testing.T) {
	svc, _, cache, _ := newTestService()
	expected := models.Transfer{ID: "id-123", SenderID: "sender-1", Amount: 100}

	cache.On("GetByID", mock.Anything, "id-123").Return(expected, nil)

	result, err := svc.GetByID(context.Background(), "id-123")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	cache.AssertExpectations(t)
}

func TestGetByID_CacheMiss_RepoHit(t *testing.T) {
	svc, repo, cache, _ := newTestService()
	expected := models.Transfer{ID: "id-123", SenderID: "sender-1", Amount: 100}

	cache.On("GetByID", mock.Anything, "id-123").Return(models.Transfer{}, errors.New("not in cache"))
	repo.On("GetByID", mock.Anything, "id-123").Return(expected, nil)

	result, err := svc.GetByID(context.Background(), "id-123")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	cache.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestGetByID_CacheMiss_RepoError(t *testing.T) {
	svc, repo, cache, _ := newTestService()
	repoErr := errors.New("db error")

	cache.On("GetByID", mock.Anything, "id-123").Return(models.Transfer{}, errors.New("not in cache"))
	repo.On("GetByID", mock.Anything, "id-123").Return(models.Transfer{}, repoErr)

	result, err := svc.GetByID(context.Background(), "id-123")

	assert.Equal(t, models.Transfer{}, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
}

// ==================== Update ====================

func TestUpdate_Success(t *testing.T) {
	svc, repo, _, _ := newTestService()
	transfer := models.Transfer{ID: "id-123", SenderID: "sender-updated"}

	repo.On("Update", mock.Anything, transfer).Return(nil)

	err := svc.Update(context.Background(), transfer)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdate_EmptyID(t *testing.T) {
	svc, _, _, _ := newTestService()
	transfer := models.Transfer{SenderID: "sender-1"}

	err := svc.Update(context.Background(), transfer)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, known_errors.ErrBadRequest))
}

func TestUpdate_NoFieldsToUpdate(t *testing.T) {
	svc, _, _, _ := newTestService()
	transfer := models.Transfer{ID: "id-123", Currency: enums.CurrencyUnknown}

	err := svc.Update(context.Background(), transfer)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, known_errors.ErrBadRequest))
}

func TestUpdate_RepoError(t *testing.T) {
	svc, repo, _, _ := newTestService()
	transfer := models.Transfer{ID: "id-123", SenderID: "sender-updated"}
	repoErr := errors.New("db error")

	repo.On("Update", mock.Anything, transfer).Return(repoErr)

	err := svc.Update(context.Background(), transfer)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
	repo.AssertExpectations(t)
}

// ==================== Delete ====================

func TestDelete_Success(t *testing.T) {
	svc, repo, _, _ := newTestService()

	repo.On("Delete", mock.Anything, "id-123").Return(nil)

	err := svc.Delete(context.Background(), "id-123")

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_RepoError(t *testing.T) {
	svc, repo, _, _ := newTestService()
	repoErr := errors.New("db error")

	repo.On("Delete", mock.Anything, "id-123").Return(repoErr)

	err := svc.Delete(context.Background(), "id-123")

	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
	repo.AssertExpectations(t)
}

// ==================== GetAll ====================

func TestGetAll_Success(t *testing.T) {
	svc, repo, _, _ := newTestService()
	expected := []models.Transfer{
		{ID: "id-1", SenderID: "sender-1"},
		{ID: "id-2", SenderID: "sender-2"},
	}

	repo.On("GetAll", mock.Anything).Return(expected, nil)

	result, err := svc.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestGetAll_RepoError(t *testing.T) {
	svc, repo, _, _ := newTestService()
	repoErr := errors.New("db error")

	repo.On("GetAll", mock.Anything).Return(nil, repoErr)

	result, err := svc.GetAll(context.Background())

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
	repo.AssertExpectations(t)
}

// ==================== GetBySenderID ====================

func TestGetBySenderID_Success(t *testing.T) {
	svc, repo, _, _ := newTestService()
	expected := []models.Transfer{
		{ID: "id-1", SenderID: "sender-1"},
	}

	repo.On("GetBySenderID", mock.Anything, "sender-1").Return(expected, nil)

	result, err := svc.GetBySenderID(context.Background(), "sender-1")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestGetBySenderID_RepoError(t *testing.T) {
	svc, repo, _, _ := newTestService()
	repoErr := errors.New("db error")

	repo.On("GetBySenderID", mock.Anything, "sender-1").Return(nil, repoErr)

	result, err := svc.GetBySenderID(context.Background(), "sender-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
	repo.AssertExpectations(t)
}

// ==================== GetByReceiverID ====================

func TestGetByReceiverID_Success(t *testing.T) {
	svc, repo, _, _ := newTestService()
	expected := []models.Transfer{
		{ID: "id-1", ReceiverID: "receiver-1"},
	}

	repo.On("GetByReceiverID", mock.Anything, "receiver-1").Return(expected, nil)

	result, err := svc.GetByReceiverID(context.Background(), "receiver-1")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestGetByReceiverID_RepoError(t *testing.T) {
	svc, repo, _, _ := newTestService()
	repoErr := errors.New("db error")

	repo.On("GetByReceiverID", mock.Anything, "receiver-1").Return(nil, repoErr)

	result, err := svc.GetByReceiverID(context.Background(), "receiver-1")

	assert.Nil(t, result)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, repoErr))
	repo.AssertExpectations(t)
}
