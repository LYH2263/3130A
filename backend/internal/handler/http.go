package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"label3130/backend/internal/auth"
	"label3130/backend/internal/dto"
	"label3130/backend/internal/middleware"
	"label3130/backend/internal/models"
	"label3130/backend/internal/service"
)

type HTTPHandler struct {
	authSvc     *service.AuthService
	questionSvc *service.QuestionService
	attemptSvc  *service.AttemptService
	tokens      *auth.TokenManager
	log         *slog.Logger
}

func New(
	authSvc *service.AuthService,
	questionSvc *service.QuestionService,
	attemptSvc *service.AttemptService,
	tokens *auth.TokenManager,
	log *slog.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		authSvc:     authSvc,
		questionSvc: questionSvc,
		attemptSvc:  attemptSvc,
		tokens:      tokens,
		log:         log,
	}
}

func (h *HTTPHandler) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(h.requestLogger())
	r.Use(cors())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.GET("/classes", h.listClasses)
		api.POST("/auth/register", h.register)
		api.POST("/auth/login", h.login)

		authed := api.Group("", middleware.AuthRequired(h.tokens))
		{
			authed.GET("/me", h.me)

			teacher := authed.Group("/teacher", middleware.RequireRole(models.RoleTeacher))
			{
				teacher.GET("/overview", h.teacherOverview)
				teacher.GET("/class-stats", h.teacherClassStats)
				teacher.GET("/attempts", h.teacherAttempts)

				teacher.GET("/questions", h.listQuestions)
				teacher.POST("/questions", h.createQuestion)
				teacher.PUT("/questions/:id", h.updateQuestion)
				teacher.DELETE("/questions/:id", h.deleteQuestion)
				teacher.POST("/questions/upload", h.uploadQuestions)
			}

			student := authed.Group("/student", middleware.RequireRole(models.RoleStudent))
			{
				student.GET("/questions", h.studentQuestions)
				student.POST("/submit", h.submit)
				student.GET("/mistakes", h.studentMistakes)
				student.GET("/attempts", h.studentAttempts)
			}
		}
	}

	return r
}

func (h *HTTPHandler) register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid register payload"})
		return
	}
	result, err := h.authSvc.RegisterStudent(req)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *HTTPHandler) login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid login payload"})
		return
	}
	result, err := h.authSvc.Login(req)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) me(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	user, err := h.authSvc.GetUser(claims.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *HTTPHandler) listClasses(c *gin.Context) {
	classes, err := h.authSvc.ListClasses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load classes"})
		return
	}
	c.JSON(http.StatusOK, classes)
}

func (h *HTTPHandler) listQuestions(c *gin.Context) {
	questions, err := h.questionSvc.ListQuestions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load questions"})
		return
	}
	c.JSON(http.StatusOK, questions)
}

func (h *HTTPHandler) createQuestion(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	var req dto.QuestionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid question payload"})
		return
	}
	question, err := h.questionSvc.CreateQuestion(req, claims.UserID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, question)
}

func (h *HTTPHandler) updateQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid question id"})
		return
	}
	var req dto.QuestionInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid question payload"})
		return
	}
	question, err := h.questionSvc.UpdateQuestion(uint(id), req)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, question)
}

func (h *HTTPHandler) deleteQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid question id"})
		return
	}
	if err := h.questionSvc.DeleteQuestion(uint(id)); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "question deleted"})
}

func (h *HTTPHandler) uploadQuestions(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing file"})
		return
	}
	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "open file failed"})
		return
	}
	defer opened.Close()

	data, err := io.ReadAll(opened)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "read file failed"})
		return
	}

	count, err := h.questionSvc.UploadFromJSON(data, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "upload success", "count": count})
}

func (h *HTTPHandler) teacherOverview(c *gin.Context) {
	overview, err := h.attemptSvc.Overview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "load overview failed"})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *HTTPHandler) teacherClassStats(c *gin.Context) {
	stats, err := h.attemptSvc.ClassWrongStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "load class stats failed"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *HTTPHandler) teacherAttempts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "30"))
	items, err := h.attemptSvc.TeacherRecentAttempts(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "load attempts failed"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *HTTPHandler) studentQuestions(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	questions, err := h.questionSvc.GetQuizQuestions(limit)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, questions)
}

func (h *HTTPHandler) submit(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims.ClassID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid student context"})
		return
	}

	var req dto.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid submit payload"})
		return
	}
	result, err := h.attemptSvc.Submit(claims.UserID, *claims.ClassID, req)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *HTTPHandler) studentMistakes(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	items, err := h.attemptSvc.StudentMistakes(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "load mistakes failed"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *HTTPHandler) studentAttempts(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	items, err := h.attemptSvc.StudentAttempts(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "load attempts failed"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *HTTPHandler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserExists):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrInvalidCredential):
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrClassNotFound), errors.Is(err, service.ErrQuestionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrInvalidQuestion), errors.Is(err, service.ErrInvalidSubmission):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrNoQuestions):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	default:
		h.log.Error("service error", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
	}
}

func (h *HTTPHandler) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		h.log.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
		)
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
