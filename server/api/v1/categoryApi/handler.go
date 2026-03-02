package categoryApi

import (
	"2_Go/internal/category"
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CategoryHandler struct {
	CatSvc *category.CategoryService
	Logger logrus.FieldLogger
}

func NewHandler(svc *category.CategoryService, logger logrus.FieldLogger) *CategoryHandler {
	return &CategoryHandler{CatSvc: svc, Logger: logger}
}

type categoryRequest struct {
	Name  string `json:"name"  binding:"required"`
	Color string `json:"color"`
}

// GET /categories
func (h *CategoryHandler) HandleList(c *gin.Context) {
	userID := mustUserID(c)
	cats, err := h.CatSvc.GetCategoriesByUser(userID)
	if err != nil {
		h.Logger.Errorf("list categories: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch categories"})
		return
	}
	if cats == nil {
		cats = []category.Category{}
	}
	c.JSON(http.StatusOK, cats)
}

// POST /categories
func (h *CategoryHandler) HandleCreate(c *gin.Context) {
	userID := mustUserID(c)
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := h.CatSvc.CreateCategory(userID, req.Name, req.Color)
	if err != nil {
		h.Logger.Errorf("create category: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create category"})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

// PATCH /categories/:guid
func (h *CategoryHandler) HandleUpdate(c *gin.Context) {
	userID := mustUserID(c)
	guid := c.Param("guid")
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.CatSvc.UpdateCategory(guid, userID, req.Name, req.Color); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		h.Logger.Errorf("update category: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}
	c.Status(http.StatusNoContent)
}

// DELETE /categories/:guid
func (h *CategoryHandler) HandleDelete(c *gin.Context) {
	userID := mustUserID(c)
	guid := c.Param("guid")
	if err := h.CatSvc.DeleteCategory(guid, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		h.Logger.Errorf("delete category: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

func mustUserID(c *gin.Context) int {
	v, _ := c.Get("userId")
	return v.(int)
}
