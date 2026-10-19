package http

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/service"
)

type DocumentHandler struct {
	documents  *service.DocumentService
	users      domain.UserRepository
	uploadsDir string
}

func NewDocumentHandler(documents *service.DocumentService, users domain.UserRepository, uploadsDir string) *DocumentHandler {
	return &DocumentHandler{documents: documents, users: users, uploadsDir: uploadsDir}
}

func (h *DocumentHandler) GetMyStatus(c *gin.Context) {
	instructorID := c.GetString("user_id")
	status, err := h.documents.GetStatus(instructorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load document status"})
		return
	}
	c.JSON(http.StatusOK, status)
}

var validDocTypes = map[domain.DocumentType]bool{
	domain.DocLicense:             true,
	domain.DocIdentity:            true,
	domain.DocVehicleRegistration: true,
	domain.DocInsurance:           true,
}

// Submit takes a multipart form with "doc_type" (license|identity|
// vehicle_registration|insurance), an optional "expiry_date" (YYYY-MM-DD),
// and the "photo" file.
// Like exams, this always goes through admin review — no auto-approve.
func (h *DocumentHandler) Submit(c *gin.Context) {
	instructorID := c.GetString("user_id")

	docType := domain.DocumentType(c.PostForm("doc_type"))
	if !validDocTypes[docType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "doc_type must be license, identity, vehicle_registration, or insurance"})
		return
	}

	var expiryDate *time.Time
	if raw := c.PostForm("expiry_date"); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "expiry_date must be in YYYY-MM-DD format"})
			return
		}
		expiryDate = &parsed
	}

	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photo file is required"})
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedPhotoExt[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photo must be jpg, png, webp, or heic"})
		return
	}

	docsDir := filepath.Join(h.uploadsDir, "documents")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store photo"})
		return
	}
	filename := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	dst := filepath.Join(docsDir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store photo"})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	photoURL := scheme + "://" + c.Request.Host + "/uploads/documents/" + filename

	doc, err := h.documents.Submit(instructorID, docType, photoURL, expiryDate)
	if errors.Is(err, service.ErrDocumentAlreadySubmitted) {
		c.JSON(http.StatusConflict, gin.H{"error": "ya tenés una entrega para este documento"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not submit document"})
		return
	}
	c.JSON(http.StatusCreated, doc)
}

type pendingDocumentResponse struct {
	*domain.InstructorDocument
	InstructorName string `json:"instructor_name"`
}

func (h *DocumentHandler) ListPending(c *gin.Context) {
	docs, err := h.documents.ListPendingReview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list pending documents"})
		return
	}

	out := make([]pendingDocumentResponse, 0, len(docs))
	for _, d := range docs {
		name := ""
		if instructor, err := h.users.GetByID(d.InstructorID); err == nil {
			name = instructor.FullName
		}
		out = append(out, pendingDocumentResponse{InstructorDocument: d, InstructorName: name})
	}
	c.JSON(http.StatusOK, out)
}

type reviewDocumentRequest struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason"`
}

func (h *DocumentHandler) Review(c *gin.Context) {
	adminID := c.GetString("user_id")
	var req reviewDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.documents.Review(c.Param("id"), adminID, req.Approve, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not review document"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
