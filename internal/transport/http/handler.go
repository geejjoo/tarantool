package http

import (
	"net/http"
	"strconv"

	"kv-storage/internal/domain"
	"kv-storage/internal/interfaces"
	"kv-storage/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.KVService
	logger  interfaces.Logger
}

func NewHandler(service *service.KVService, logger interfaces.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Create godoc
// @Summary Create a new key-value pair
// @Description Create a new key-value pair in the storage
// @Tags kv
// @Accept json
// @Produce json
// @Param kv body domain.CreateKVRequest true "Key-value pair to create"
// @Success 201 {object} domain.KV
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/kv [post]
func (h *Handler) Create(c *gin.Context) {
	var req domain.CreateKVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	kv, err := h.service.Create(&req)
	if err != nil {
		switch err {
		case domain.ErrInvalidKey, domain.ErrInvalidValue:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrKeyExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case domain.ErrKeyAlreadyExists:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			h.logger.Error("Failed to create KV", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, kv)
}

// Get godoc
// @Summary Get a key-value pair by key
// @Description Retrieve a key-value pair from the storage by its key
// @Tags kv
// @Produce json
// @Param key path string true "Key to retrieve"
// @Success 200 {object} domain.KV
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/kv/{key} [get]
func (h *Handler) Get(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	kv, err := h.service.Get(key)
	h.logger.Info("Get result", "key", key, "kv", kv, "err", err)

	if err != nil {
		switch err {
		case domain.ErrInvalidKey:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrKeyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			h.logger.Error("Failed to get KV", "key", key, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, kv)
}

// Update godoc
// @Summary Update a key-value pair
// @Description Update an existing key-value pair in the storage
// @Tags kv
// @Accept json
// @Produce json
// @Param key path string true "Key to update"
// @Param kv body domain.UpdateKVRequest true "New value for the key"
// @Success 200 {object} domain.KV
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/kv/{key} [put]
func (h *Handler) Update(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	var req domain.UpdateKVRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	kv, err := h.service.Update(key, &req)
	if err != nil {
		switch err {
		case domain.ErrInvalidKey, domain.ErrInvalidValue:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrKeyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			h.logger.Error("Failed to update KV", "key", key, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, kv)
}

// Delete godoc
// @Summary Delete a key-value pair
// @Description Delete a key-value pair from the storage (hard delete by default, soft delete if specified)
// @Tags kv
// @Accept json
// @Produce json
// @Param key path string true "Key to delete"
// @Param delete body domain.DeleteKVRequest false "Delete options"
// @Success 200 {object} domain.KV
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/kv/{key} [delete]
func (h *Handler) Delete(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	var deleteReq domain.DeleteKVRequest
	if err := c.ShouldBindJSON(&deleteReq); err != nil && err.Error() != "EOF" {
		h.logger.Error("Failed to bind JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var kv *domain.KV
	var err error

	if !deleteReq.SoftDelete {
		kv, err = h.service.SoftDelete(key)
	} else {
		kv, err = h.service.Delete(key)
	}

	if err != nil {
		switch err {
		case domain.ErrInvalidKey:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrKeyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			h.logger.Error("Failed to delete KV", "key", key, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, kv)
}

// Restore godoc
// @Summary Restore a soft-deleted key-value pair
// @Description Restore a soft-deleted key-value pair
// @Tags kv
// @Produce json
// @Param key path string true "Key to restore"
// @Success 200 {object} domain.KV
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/kv/{key}/restore [post]
func (h *Handler) Restore(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	kv, err := h.service.Restore(key)
	if err != nil {
		switch err {
		case domain.ErrInvalidKey:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case domain.ErrKeyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			h.logger.Error("Failed to restore KV", "key", key, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, kv)
}

// List godoc
// @Summary List key-value pairs
// @Description Get a paginated list of key-value pairs from the storage (excluding soft-deleted)
// @Tags kv
// @Produce json
// @Param limit query int false "Number of items to return (default: 10, max: 100)"
// @Param offset query int false "Number of items to skip (default: 0)"
// @Success 200 {object} domain.ListKVResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/kv [get]
func (h *Handler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	response, err := h.service.List(limit, offset)
	if err != nil {
		h.logger.Error("Failed to list KV", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListIncludingDeleted godoc
// @Summary List all key-value pairs including deleted
// @Description Get a paginated list of all key-value pairs including soft-deleted ones
// @Tags kv
// @Produce json
// @Param limit query int false "Number of items to return (default: 10, max: 100)"
// @Param offset query int false "Number of items to skip (default: 0)"
// @Success 200 {object} domain.ListKVResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/kv/all [get]
func (h *Handler) ListIncludingDeleted(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	response, err := h.service.ListIncludingDeleted(limit, offset)
	if err != nil {
		h.logger.Error("Failed to list KV including deleted", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck godoc
// @Summary Health check
// @Description Returns the health status of the service
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: "kv-storage",
	})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}
