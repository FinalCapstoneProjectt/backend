package documentations

import (
	"backend/internal/auth"
	"backend/pkg/response"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler { return &Handler{service: s} }

func (h *Handler) GetProjectDocs(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Param("id"), 10, 32) 
	docs, err := h.service.GetDocs(uint(projectID))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error", err.Error())
		return
	}
	response.Success(c, docs)
}

func (h *Handler) Submit(c *gin.Context) {
	claims, _ := c.Get("claims")
	userClaims := claims.(*auth.TokenClaims)
	projectID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var docType, url string

	// 1. Try to get from JSON (For Links)
	var jsonReq struct {
		DocumentType string `json:"document_type"`
		URL          string `json:"url"`
	}
	
	// If it's JSON, bind it
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&jsonReq); err == nil {
			docType = jsonReq.DocumentType
			url = jsonReq.URL
		}
	} else {
		// 2. Otherwise get from Form (For Files)
		docType = c.PostForm("document_type")
		url = c.PostForm("url")
	}

	file, _ := c.FormFile("file")

	// 3. Call Service
	doc, err := h.service.SubmitDoc(uint(projectID), userClaims.UserID, docType, url, file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.JSON(c, http.StatusCreated, "Success", doc)
}

func (h *Handler) Delete(c *gin.Context) {
	claims, _ := c.Get("claims")
	userClaims := claims.(*auth.TokenClaims)
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.service.DeleteDoc(uint(docID), userClaims.UserID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.JSON(c, http.StatusOK, "Deleted", nil)
}

func (h *Handler) Review(c *gin.Context) {
	claims, _ := c.Get("claims")
	userClaims := claims.(*auth.TokenClaims)
	docID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req struct {
		Status  string `json:"status"` // approved, rejected
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&req)

	if err := h.service.ReviewDoc(uint(docID), userClaims.UserID, req.Status, req.Comment); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	response.JSON(c, http.StatusOK, "Review recorded", nil)
}