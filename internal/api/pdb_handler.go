package api

import (
	"context"
	"net/http"

	"kube-tide/internal/core/k8s"

	"github.com/gin-gonic/gin"
)

type PDBHandler struct {
	service *k8s.PDBService
}

func NewPDBHandler(service *k8s.PDBService) *PDBHandler {
	return &PDBHandler{service: service}
}

func (h *PDBHandler) ListPDBs(c *gin.Context) {
	namespace := namespaceFromRequest(c)
	items, err := h.service.ListPDBs(context.Background(), c.Param("cluster"), namespace)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pdb.listFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pdbs": items})
}

func (h *PDBHandler) GetPDB(c *gin.Context) {
	item, err := h.service.GetPDB(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("pdb"))
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pdb.getFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pdb": item})
}

func (h *PDBHandler) CreatePDB(c *gin.Context) {
	var req k8s.CreatePDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "pdb.invalidRequest", err.Error())
		return
	}
	item, err := h.service.CreatePDB(context.Background(), c.Param("cluster"), c.Param("namespace"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pdb.createFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pdb": item})
}

func (h *PDBHandler) UpdatePDB(c *gin.Context) {
	var req k8s.UpdatePDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "pdb.invalidRequest", err.Error())
		return
	}
	item, err := h.service.UpdatePDB(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("pdb"), req)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, "pdb.updateFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"pdb": item})
}

func (h *PDBHandler) DeletePDB(c *gin.Context) {
	if err := h.service.DeletePDB(context.Background(), c.Param("cluster"), c.Param("namespace"), c.Param("pdb")); err != nil {
		ResponseError(c, http.StatusInternalServerError, "pdb.deleteFailed", err.Error())
		return
	}
	ResponseSuccess(c, gin.H{"message": "PDB deleted successfully"})
}
