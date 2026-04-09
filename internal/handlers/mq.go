package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MqService interface {
	Read(ctx context.Context) (string, error)
}

type MqHandler struct {
	mqSvc MqService
}

func NewMqHandler(service MqService) *MqHandler {
	return &MqHandler{
		mqSvc: service,
	}
}

func (h *MqHandler) Read(ctx *gin.Context) {
	result, err := h.mqSvc.Read(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": result})
}
