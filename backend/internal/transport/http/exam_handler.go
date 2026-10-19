package http

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/jmadrazo/clase-drive-backend/internal/domain"
	"github.com/jmadrazo/clase-drive-backend/internal/service"
)

type ExamHandler struct {
	exams      *service.ExamService
	users      domain.UserRepository
	uploadsDir string
}

func NewExamHandler(exams *service.ExamService, users domain.UserRepository, uploadsDir string) *ExamHandler {
	return &ExamHandler{exams: exams, users: users, uploadsDir: uploadsDir}
}

func (h *ExamHandler) GetMyStatus(c *gin.Context) {
	studentID := c.GetString("user_id")
	status, err := h.exams.GetStatus(studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load exam status"})
		return
	}
	c.JSON(http.StatusOK, status)
}

var allowedPhotoExt = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".heic": true}

// Submit takes a multipart form with "type" (theory|practice) and "photo"
// (the exam result image), saves the file to disk, and creates the exam
// submission. This paper is what the student's real license depends on, so
// it always goes through review — there's no auto-approve path.
func (h *ExamHandler) Submit(c *gin.Context) {
	studentID := c.GetString("user_id")

	examType := domain.ExamType(c.PostForm("type"))
	if examType != domain.ExamTheory && examType != domain.ExamPractice {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type must be 'theory' or 'practice'"})
		return
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

	examDir := filepath.Join(h.uploadsDir, "exams")
	if err := os.MkdirAll(examDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store photo"})
		return
	}
	filename := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	dst := filepath.Join(examDir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not store photo"})
		return
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	photoURL := scheme + "://" + c.Request.Host + "/uploads/exams/" + filename

	exam, err := h.exams.Submit(studentID, examType, photoURL)
	if errors.Is(err, service.ErrExamLocked) {
		c.JSON(http.StatusForbidden, gin.H{"error": "todavía no habilitaste este examen"})
		return
	}
	if errors.Is(err, service.ErrExamAlreadySubmitted) {
		c.JSON(http.StatusConflict, gin.H{"error": "ya tenés una entrega para este examen"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not submit exam"})
		return
	}
	c.JSON(http.StatusCreated, exam)
}

type pendingExamResponse struct {
	*domain.Exam
	StudentName string `json:"student_name"`
}

func (h *ExamHandler) ListPending(c *gin.Context) {
	exams, err := h.exams.ListPendingReview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list pending exams"})
		return
	}

	out := make([]pendingExamResponse, 0, len(exams))
	for _, e := range exams {
		name := ""
		if student, err := h.users.GetByID(e.StudentID); err == nil {
			name = student.FullName
		}
		out = append(out, pendingExamResponse{Exam: e, StudentName: name})
	}
	c.JSON(http.StatusOK, out)
}

type reviewExamRequest struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason"`
}

func (h *ExamHandler) Review(c *gin.Context) {
	adminID := c.GetString("user_id")
	var req reviewExamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.exams.Review(c.Param("id"), adminID, req.Approve, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not review exam"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
