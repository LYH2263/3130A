package handler

import (
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"label3130/backend/internal/auth"
	"label3130/backend/internal/dto"
	"label3130/backend/internal/middleware"
	"label3130/backend/internal/models"
	"label3130/backend/internal/service"
)

type HTTPHandler struct {
	authSvc            *service.AuthService
	categorySvc        *service.CategoryService
	tagSvc             *service.TagService
	knowledgePointSvc  *service.KnowledgePointService
	questionSvc        *service.QuestionService
	attemptSvc         *service.AttemptService
	mistakeReviewSvc   *service.MistakeReviewService
	favoriteSvc        *service.FavoriteService
	questionSetSvc     *service.QuestionSetService
	tokens             *auth.TokenManager
	log                *slog.Logger
}

func New(
	authSvc *service.AuthService,
	categorySvc *service.CategoryService,
	tagSvc *service.TagService,
	knowledgePointSvc *service.KnowledgePointService,
	questionSvc *service.QuestionService,
	attemptSvc *service.AttemptService,
	mistakeReviewSvc *service.MistakeReviewService,
	favoriteSvc *service.FavoriteService,
	questionSetSvc *service.QuestionSetService,
	tokens *auth.TokenManager,
	log *slog.Logger,
) *HTTPHandler {
	return &HTTPHandler{
		authSvc:            authSvc,
		categorySvc:        categorySvc,
		tagSvc:             tagSvc,
		knowledgePointSvc:  knowledgePointSvc,
		questionSvc:        questionSvc,
		attemptSvc:         attemptSvc,
		mistakeReviewSvc:   mistakeReviewSvc,
		favoriteSvc:        favoriteSvc,
		questionSetSvc:     questionSetSvc,
		tokens:             tokens,
		log:                log,
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
				teacher.GET("/leaderboard", h.teacherLeaderboard)

				teacher.GET("/categories", h.listCategories)
				teacher.POST("/categories", h.createCategory)
				teacher.PUT("/categories/:id", h.updateCategory)
				teacher.DELETE("/categories/:id", h.deleteCategory)

				teacher.GET("/tags", h.listTags)
				teacher.POST("/tags", h.createTag)
				teacher.PUT("/tags/:id", h.updateTag)
				teacher.DELETE("/tags/:id", h.deleteTag)

				teacher.GET("/knowledge-points", h.listKnowledgePoints)
				teacher.POST("/knowledge-points", h.createKnowledgePoint)
				teacher.PUT("/knowledge-points/:id", h.updateKnowledgePoint)
				teacher.DELETE("/knowledge-points/:id", h.deleteKnowledgePoint)

				teacher.GET("/questions", h.listQuestions)
				teacher.GET("/questions/:id", h.getQuestion)
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
				student.GET("/attempts/:id", h.getAttemptDetail)
				student.GET("/mistake-review/quiz", h.mistakeReviewQuiz)
				student.POST("/mistake-review/submit", h.mistakeReviewSubmit)
				student.POST("/draft", h.saveDraft)
				student.GET("/draft", h.getDraft)
				student.DELETE("/draft", h.clearDraft)
				student.GET("/explanations", h.getExplanations)
				student.GET("/questions/:id/explanation", h.getQuestionExplanation)
				student.GET("/leaderboard", h.studentLeaderboard)

				student.POST("/favorites/:questionId", h.toggleFavorite)
				student.DELETE("/favorites/:questionId", h.toggleFavorite)
				student.GET("/favorites", h.listFavorites)
				student.GET("/favorites/status", h.getFavoriteStatus)
				student.GET("/favorites/quiz", h.favoriteQuiz)

				student.GET("/question-sets", h.listQuestionSets)
				student.POST("/question-sets", h.createQuestionSet)
				student.GET("/question-sets/:id", h.getQuestionSet)
				student.PUT("/question-sets/:id", h.updateQuestionSet)
				student.DELETE("/question-sets/:id", h.deleteQuestionSet)
				student.POST("/question-sets/:id/questions", h.addQuestionsToSet)
				student.DELETE("/question-sets/:id/questions/:questionId", h.removeQuestionFromSet)
				student.POST("/question-sets/:id/reorder", h.reorderSetQuestions)
				student.GET("/question-sets/:id/quiz", h.getSetQuiz)
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
	var query dto.QuestionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid query parameters"})
		return
	}

	result, err := h.questionSvc.QueryQuestions(query, h.categorySvc)
	if err != nil {
		h.log.Error("list questions failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load questions"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) getQuestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid question id"})
		return
	}
	question, err := h.questionSvc.GetQuestion(uint(id))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, question)
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
	question, err := h.questionSvc.CreateQuestion(req, claims.UserID, h.tagSvc, h.knowledgePointSvc)
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
	question, err := h.questionSvc.UpdateQuestion(uint(id), req, h.tagSvc, h.knowledgePointSvc)
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

	count, err := h.questionSvc.UploadFromJSON(data, claims.UserID, h.tagSvc, h.knowledgePointSvc)
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

	_ = h.attemptSvc.ClearDraft(claims.UserID, models.QuizModeNormal)

	c.JSON(http.StatusCreated, result)
}

func (h *HTTPHandler) studentMistakes(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	items, err := h.mistakeReviewSvc.GetMistakeReviews(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "load mistakes failed"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *HTTPHandler) mistakeReviewQuiz(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	questions, err := h.mistakeReviewSvc.GenerateReviewQuiz(claims.UserID, limit)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, questions)
}

func (h *HTTPHandler) mistakeReviewSubmit(c *gin.Context) {
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
	result, err := h.mistakeReviewSvc.SubmitReview(claims.UserID, *claims.ClassID, req)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}

	_ = h.attemptSvc.ClearDraft(claims.UserID, models.QuizModeReview)

	c.JSON(http.StatusCreated, result)
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

func (h *HTTPHandler) studentLeaderboard(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	var query dto.LeaderboardQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid query parameters"})
		return
	}

	query.ClassID = claims.ClassID

	result, err := h.attemptSvc.GetLeaderboard(query, &claims.UserID)
	if err != nil {
		h.log.Error("student leaderboard failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "load leaderboard failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) teacherLeaderboard(c *gin.Context) {
	_, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	var query dto.LeaderboardQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid query parameters"})
		return
	}

	result, err := h.attemptSvc.GetLeaderboard(query, nil)
	if err != nil {
		h.log.Error("teacher leaderboard failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "load leaderboard failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) getAttemptDetail(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid attempt id"})
		return
	}
	report, err := h.attemptSvc.GetAttemptDetail(claims.UserID, uint(id))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *HTTPHandler) saveDraft(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	var req dto.SaveDraftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid draft payload"})
		return
	}

	updatedAt, err := h.attemptSvc.SaveDraft(claims.UserID, req)
	if err != nil {
		if errors.Is(err, service.ErrDraftConflict) {
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		h.log.Error("save draft failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "save draft failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "draft saved", "updatedAt": updatedAt})
}

func (h *HTTPHandler) getDraft(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	quizMode := c.DefaultQuery("mode", "normal")
	if quizMode != "normal" && quizMode != "review" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid quiz mode"})
		return
	}

	draft, err := h.attemptSvc.GetDraft(claims.UserID, quizMode)
	if err != nil {
		if errors.Is(err, service.ErrDraftNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "draft not found"})
			return
		}
		h.log.Error("get draft failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "get draft failed"})
		return
	}
	c.JSON(http.StatusOK, draft)
}

func (h *HTTPHandler) clearDraft(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	quizMode := c.DefaultQuery("mode", "normal")
	if quizMode != "normal" && quizMode != "review" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid quiz mode"})
		return
	}

	if err := h.attemptSvc.ClearDraft(claims.UserID, quizMode); err != nil {
		h.log.Error("clear draft failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "clear draft failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "draft cleared"})
}

func (h *HTTPHandler) listCategories(c *gin.Context) {
	categories, err := h.categorySvc.ListTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load categories"})
		return
	}
	c.JSON(http.StatusOK, categories)
}

func (h *HTTPHandler) createCategory(c *gin.Context) {
	var req dto.CategoryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid category payload"})
		return
	}
	category, err := h.categorySvc.CreateCategory(req.Name, req.ParentID, req.Sort)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (h *HTTPHandler) updateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid category id"})
		return
	}
	var req dto.CategoryInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid category payload"})
		return
	}
	category, err := h.categorySvc.UpdateCategory(uint(id), req.Name, req.ParentID, req.Sort)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, category)
}

func (h *HTTPHandler) deleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid category id"})
		return
	}
	if err := h.categorySvc.DeleteCategory(uint(id)); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}

func (h *HTTPHandler) listTags(c *gin.Context) {
	tags, err := h.tagSvc.ListTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load tags"})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func (h *HTTPHandler) createTag(c *gin.Context) {
	var req dto.TagInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid tag payload"})
		return
	}
	tag, err := h.tagSvc.CreateTag(req.Name)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, tag)
}

func (h *HTTPHandler) updateTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid tag id"})
		return
	}
	var req dto.TagInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid tag payload"})
		return
	}
	tag, err := h.tagSvc.UpdateTag(uint(id), req.Name)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, tag)
}

func (h *HTTPHandler) deleteTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid tag id"})
		return
	}
	if err := h.tagSvc.DeleteTag(uint(id)); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "tag deleted"})
}

func (h *HTTPHandler) listKnowledgePoints(c *gin.Context) {
	kps, err := h.knowledgePointSvc.ListKnowledgePoints()
	if err != nil {
		h.log.Error("list knowledge points failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load knowledge points"})
		return
	}
	c.JSON(http.StatusOK, kps)
}

func (h *HTTPHandler) createKnowledgePoint(c *gin.Context) {
	var req dto.KnowledgePointInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid knowledge point payload"})
		return
	}
	kp, err := h.knowledgePointSvc.CreateKnowledgePoint(req.Name, req.Sort)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, kp)
}

func (h *HTTPHandler) updateKnowledgePoint(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid knowledge point id"})
		return
	}
	var req dto.KnowledgePointInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid knowledge point payload"})
		return
	}
	kp, err := h.knowledgePointSvc.UpdateKnowledgePoint(uint(id), req.Name, req.Sort)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, kp)
}

func (h *HTTPHandler) deleteKnowledgePoint(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid knowledge point id"})
		return
	}
	if err := h.knowledgePointSvc.DeleteKnowledgePoint(uint(id)); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "knowledge point deleted"})
}

func (h *HTTPHandler) getQuestionExplanation(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid question id"})
		return
	}
	explanation, err := h.questionSvc.GetExplanationForStudent(uint(id), claims.UserID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, explanation)
}

func (h *HTTPHandler) getExplanations(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	idsStr := c.QueryArray("questionIds")
	questionIDs := make([]uint, 0, len(idsStr))
	for _, s := range idsStr {
		id, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			questionIDs = append(questionIDs, uint(id))
		}
	}

	explanations, err := h.questionSvc.BatchGetExplanationsForStudent(questionIDs, claims.UserID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, explanations)
}

func (h *HTTPHandler) respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserExists):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrInvalidCredential):
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrClassNotFound),
		errors.Is(err, service.ErrQuestionNotFound),
		errors.Is(err, service.ErrCategoryNotFound),
		errors.Is(err, service.ErrTagNotFound),
		errors.Is(err, service.ErrKnowledgePointNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrTagExists),
		errors.Is(err, service.ErrKnowledgePointExists):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrNoQuestions):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrExplanationUnauthorized):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrInvalidQuestion),
		errors.Is(err, service.ErrInvalidSubmission),
		errors.Is(err, service.ErrInvalidQuestionType),
		errors.Is(err, service.ErrInvalidSingleOptionCount),
		errors.Is(err, service.ErrInvalidSingleCorrect),
		errors.Is(err, service.ErrInvalidMultipleOptionCount),
		errors.Is(err, service.ErrInvalidMultipleCorrect),
		errors.Is(err, service.ErrInvalidJudgeOptionCount),
		errors.Is(err, service.ErrInvalidJudgeCorrect),
		errors.Is(err, service.ErrInvalidBlankAnswerCount),
		errors.Is(err, service.ErrInvalidBlankAnswer),
		errors.Is(err, service.ErrInvalidOptionContent),
		errors.Is(err, service.ErrCategoryNameEmpty),
		errors.Is(err, service.ErrCategoryParentInvalid),
		errors.Is(err, service.ErrTagNameEmpty),
		errors.Is(err, service.ErrKnowledgePointNameEmpty),
		errors.Is(err, service.ErrQuestionSetNameEmpty):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrQuestionSetNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrQuestionSetForbidden):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrAttemptNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrAttemptForbidden):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrDraftConflict):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, service.ErrNoValidQuestions):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
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

func (h *HTTPHandler) toggleFavorite(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	questionID, err := strconv.ParseUint(c.Param("questionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid question id"})
		return
	}
	favorited, err := h.favoriteSvc.ToggleFavorite(claims.UserID, uint(questionID))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"favorited": favorited})
}

func (h *HTTPHandler) listFavorites(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	items, err := h.favoriteSvc.GetFavorites(claims.UserID)
	if err != nil {
		h.log.Error("list favorites failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load favorites"})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *HTTPHandler) getFavoriteStatus(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}

	idsStr := c.QueryArray("questionIds")
	questionIDs := make([]uint, 0, len(idsStr))
	for _, s := range idsStr {
		id, err := strconv.ParseUint(s, 10, 64)
		if err == nil {
			questionIDs = append(questionIDs, uint(id))
		}
	}

	status, err := h.favoriteSvc.GetFavoriteStatus(claims.UserID, questionIDs)
	if err != nil {
		h.log.Error("get favorite status failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load favorite status"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *HTTPHandler) favoriteQuiz(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	questions, err := h.favoriteSvc.GetFavoriteQuestions(claims.UserID)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	if len(questions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "no favorite questions"})
		return
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]service.StudentQuestion, 0, len(questions))
	for _, q := range questions {
		sq := service.StudentQuestion{
			ID:          q.ID,
			Type:        q.Type,
			Title:       q.Title,
			Description: q.Description,
		}
		if q.Type != models.QuestionTypeBlank {
			opts := make([]service.StudentOption, 0, len(q.Options))
			for _, opt := range q.Options {
				opts = append(opts, service.StudentOption{ID: opt.ID, Content: opt.Content})
			}
			r.Shuffle(len(opts), func(i, j int) {
				opts[i], opts[j] = opts[j], opts[i]
			})
			sq.Options = opts
		}
		result = append(result, sq)
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listQuestionSets(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	sets, err := h.questionSetSvc.ListSets(claims.UserID)
	if err != nil {
		h.log.Error("list question sets failed", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load question sets"})
		return
	}
	c.JSON(http.StatusOK, sets)
}

func (h *HTTPHandler) getQuestionSet(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid set id"})
		return
	}
	set, err := h.questionSetSvc.GetSet(claims.UserID, uint(id))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, set)
}

func (h *HTTPHandler) createQuestionSet(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	var req dto.CreateQuestionSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}
	set, err := h.questionSetSvc.CreateSet(claims.UserID, req)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, set)
}

func (h *HTTPHandler) updateQuestionSet(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid set id"})
		return
	}
	var req dto.UpdateQuestionSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}
	set, err := h.questionSetSvc.UpdateSet(claims.UserID, uint(id), req)
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, set)
}

func (h *HTTPHandler) deleteQuestionSet(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid set id"})
		return
	}
	if err := h.questionSetSvc.DeleteSet(claims.UserID, uint(id)); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "question set deleted"})
}

func (h *HTTPHandler) addQuestionsToSet(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	setID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid set id"})
		return
	}
	var req dto.AddQuestionsToSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}
	if err := h.questionSetSvc.AddQuestions(claims.UserID, uint(setID), req.QuestionIDs); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "questions added to set"})
}

func (h *HTTPHandler) removeQuestionFromSet(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	setID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid set id"})
		return
	}
	questionID, err := strconv.ParseUint(c.Param("questionId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid question id"})
		return
	}
	if err := h.questionSetSvc.RemoveQuestion(claims.UserID, uint(setID), uint(questionID)); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "question removed from set"})
}

func (h *HTTPHandler) reorderSetQuestions(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	setID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid set id"})
		return
	}
	var req dto.ReorderSetQuestionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request payload"})
		return
	}
	if err := h.questionSetSvc.ReorderQuestions(claims.UserID, uint(setID), req.QuestionIDs); err != nil {
		h.respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "questions reordered"})
}

func (h *HTTPHandler) getSetQuiz(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
		return
	}
	setID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid set id"})
		return
	}
	questions, err := h.questionSetSvc.GetSetQuestions(claims.UserID, uint(setID))
	if err != nil {
		h.respondServiceError(c, err)
		return
	}
	if len(questions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "no questions in set"})
		return
	}
	c.JSON(http.StatusOK, questions)
}
