package maintenanceApi

import (
	"2_Go/internal/maintenance"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const msgNotAuthenticated = "not authenticated"

type MaintenanceHandler struct {
	MntSvc *maintenance.MaintenanceService
	Logger logrus.FieldLogger
}

func NewHandler(svc *maintenance.MaintenanceService, logger logrus.FieldLogger) *MaintenanceHandler {
	return &MaintenanceHandler{MntSvc: svc, Logger: logger}
}

type repairRequest struct {
	Issues []maintenance.IssueRef `json:"issues" binding:"required,min=1,max=1000,dive"`
}

// GET /maintenance/verify – lists file issues of the authenticated user's documents.
func (h *MaintenanceHandler) HandleVerify(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		return
	}

	issues, err := h.MntSvc.VerifyUserDocuments(c.Request.Context(), userID)
	if err != nil {
		h.Logger.Errorf("verify documents user=%d: %s", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify documents"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"issues": issues})
}

// POST /maintenance/repair – repairs the selected issues of the authenticated user's documents.
func (h *MaintenanceHandler) HandleRepair(c *gin.Context) {
	userID, ok := userIDFrom(c)
	if !ok {
		return
	}

	var req repairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "select at least one issue to repair"})
		return
	}

	result, err := h.MntSvc.RepairUserDocuments(c.Request.Context(), userID, req.Issues)
	if err != nil {
		h.Logger.Errorf("repair documents user=%d: %s", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to repair documents"})
		return
	}

	for _, o := range append(result.Repaired, result.Skipped...) {
		if o.Err != nil {
			h.Logger.Errorf("repair user=%d obj=%d code=%d: %s", userID, o.ObjId, o.Code, o.Err)
		}
	}
	c.JSON(http.StatusOK, result)
}

func userIDFrom(c *gin.Context) (int, bool) {
	v, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": msgNotAuthenticated})
		return 0, false
	}
	return v.(int), true
}
