package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/usecase"

	"github.com/labstack/echo/v4"
)

type ApiHandler interface {
	GetDomain(c echo.Context) error
	GetDomains(c echo.Context) error
	UpdateDomain(c echo.Context) error
	DeleteDomain(c echo.Context) error
	FetchDomains(c echo.Context) error
	AnalyzeDomains(c echo.Context) error
	GetTargets(c echo.Context) error
	CreateTarget(c echo.Context) error
	UpdateTarget(c echo.Context) error
	DeleteTarget(c echo.Context) error
	GetLogs(c echo.Context) error
	CreateLog(c echo.Context) error
	GetTasks(c echo.Context) error
	CreateTask(c echo.Context) error
	UpdateTask(c echo.Context) error
	DeleteTask(c echo.Context) error
	ExecuteTasks(c echo.Context) error
	ExecuteTask(c echo.Context) error
}

type apiHandler struct {
	fetchUsecase  usecase.FetchUsecase
	domainUsecase usecase.DomainUsecase
	targetUsecase usecase.TargetUsecase
	logUsecase    usecase.LogUsecase
	gptUsecase    usecase.GptUsecase
	taskUsecase   usecase.TaskUsecase
}

func NewApiHandler(
	fetchUsecase usecase.FetchUsecase,
	domainUsecase usecase.DomainUsecase,
	targetUsecase usecase.TargetUsecase,
	logUsecase usecase.LogUsecase,
	gptUsecase usecase.GptUsecase,
	taskUsecase usecase.TaskUsecase,
) ApiHandler {
	return &apiHandler{
		fetchUsecase:  fetchUsecase,
		domainUsecase: domainUsecase,
		targetUsecase: targetUsecase,
		logUsecase:    logUsecase,
		gptUsecase:    gptUsecase,
		taskUsecase:   taskUsecase,
	}
}

// GetDomain godoc
// @Summary Get domain
// @Description Get domain
// @Tags ドメイン
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {array} domain.Domain
// @Router /domains/{id} [get]
func (h *apiHandler) GetDomain(c echo.Context) error {
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.domainUsecase.GetDomain(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// GetDomains godoc
// @Summary Get domains
// @Description Get domain list
// @Tags ドメイン
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param name query string false "ドメイン名"
// @Param can_view query boolean false "閲覧可能か"
// @Param is_japan query boolean false "日本のサイトか"
// @Param is_send query boolean false "mapsで問い合わせページを開いたか"
// @Param owner_id query string false "owner_id"
// @Param status query string false "ステータス"
// @Param industry query string false "業種"
// @Param is_ssl query boolean false "SSL対応可否"
// @Success 200 {array} []domain.Domain
// @Router /domains [get]
func (h *apiHandler) GetDomains(c echo.Context) error {
	var req domain.GetDomainsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.domainUsecase.GetDomains(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// UpdateDomain godoc
// @Summary Update domain
// @Description Update domain information
// @Tags ドメイン
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Param request body domain.UpdateDomainRequest true "更新ドメイン情報"
// @Success 200 {object} domain.Domain
// @Router /domains/{id} [put]
func (h *apiHandler) UpdateDomain(c echo.Context) error {
	var req domain.UpdateDomainRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return handleError(c, err)
	}
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.domainUsecase.UpdateDomain(c.Request().Context(), id, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// DeleteDomain godoc
// @Summary Delete domain
// @Description Delete domain by name
// @Tags ドメイン
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 204
// @Router /domains/{id} [delete]
func (h *apiHandler) DeleteDomain(c echo.Context) error {
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := h.domainUsecase.DeleteDomain(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// AnalyzeDomains godoc
// @Summary サイトの情報を解析する
// @Tags ドメイン
// @Accept json
// @Produce json
// @Success 204
// @Router /domains/analyze [post]
func (h *apiHandler) AnalyzeDomains(c echo.Context) error {
	err := h.gptUsecase.AnalyzeDomains(c.Request().Context())
	if err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// FetchDomains godoc
// @Summary Fetch domains
// @Description Fetch domain information from target
// @Tags ViewDNS
// @Accept json
// @Produce json
// @Param request body domain.PostFetchRequest true "Fetch request"
// @Success 202
// @Router /fetch [post]
func (h *apiHandler) FetchDomains(c echo.Context) error {
	var req domain.PostFetchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return handleError(c, err)
	}
	fmt.Println(req)
	err := h.fetchUsecase.Fetch(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// GetTargets godoc
// @Summary Get targets
// @Description Get target list
// @Tags ターゲット
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} domain.Target
// @Router /targets [get]
func (h *apiHandler) GetTargets(c echo.Context) error {
	var req domain.GetTargetsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.targetUsecase.GetTargets(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateTarget godoc
// @Summary Create target
// @Description Create new target
// @Tags ターゲット
// @Accept json
// @Produce json
// @Param request body domain.CreateTargetRequest true "作成ターゲット情報"
// @Success 201 {object} domain.Target
// @Router /targets [post]
func (h *apiHandler) CreateTarget(c echo.Context) error {
	var req domain.CreateTargetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.targetUsecase.CreateTarget(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusCreated, resp)
}

// UpdateTarget godoc
// @Summary Update target
// @Description Update target information
// @Tags ターゲット
// @Accept json
// @Produce json
// @Param request body domain.UpdateTargetRequest true "更新ターゲット情報"
// @Success 200 {object} domain.Target
// @Router /targets/{id} [put]
func (h *apiHandler) UpdateTarget(c echo.Context) error {
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var req domain.UpdateTargetRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.targetUsecase.UpdateTarget(c.Request().Context(), id, req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// DeleteTarget godoc
// @Summary Delete target
// @Description Delete target by id
// @Tags ターゲット
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 204
// @Router /targets/{id} [delete]
func (h *apiHandler) DeleteTarget(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err = h.targetUsecase.DeleteTarget(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// GetLogs godoc
// @Summary Get logs
// @Description Get log list
// @Tags ログ
// @Accept json
// @Produce json
// @Param name query string false "処理名"
// @Param category query string false "カテゴリー"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} domain.Log
// @Router /logs [get]
func (h *apiHandler) GetLogs(c echo.Context) error {
	var req domain.GetLogsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.logUsecase.GetLogs(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateLog godoc
// @Summary Create log
// @Description Create new log
// @Tags ログ
// @Accept json
// @Produce json
// @Param request body domain.CreateLogRequest true "作成ログ情報"
// @Success 201 {object} domain.Log
// @Router /logs [post]
func (h *apiHandler) CreateLog(c echo.Context) error {
	var req domain.CreateLogRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.logUsecase.CreateLogs(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusCreated, resp)
}

// GetTasks godoc
// @Summary Get tasks
// @Description Get task list
// @Tags タスク
// @Accept json
// @Produce json
// @Success 200 {array} domain.Task
// @Router /tasks [get]
func (h *apiHandler) GetTasks(c echo.Context) error {
	resp, err := h.taskUsecase.GetTasks(c.Request().Context())
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateTask godoc
// @Summary Create task
// @Description Create new task
// @Tags タスク
// @Accept json
// @Produce json
// @Param request body domain.CreateTaskRequest true "作成タスク情報"
// @Success 201 {object} domain.Task
// @Router /tasks [post]
func (h *apiHandler) CreateTask(c echo.Context) error {
	var req domain.CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.taskUsecase.CreateTask(c.Request().Context(), &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusCreated, resp)
}

// UpdateTask godoc
// @Summary Update task
// @Description Update task information
// @Tags タスク
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param request body domain.UpdateTaskRequest true "更新タスク情報"
// @Success 200 {object} domain.Task
// @Router /tasks/{id} [put]
func (h *apiHandler) UpdateTask(c echo.Context) error {
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var req domain.UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.taskUsecase.UpdateTask(c.Request().Context(), id, &req)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

// DeleteTask godoc
// @Summary Delete task
// @Description Delete task by id
// @Tags タスク
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 204
// @Router /tasks/{id} [delete]
func (h *apiHandler) DeleteTask(c echo.Context) error {
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err := h.taskUsecase.DeleteTask(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// ExecuteTasks godoc
// @Summary 全てのタスクを実行します
// @Description
// @Tags タスク
// @Accept json
// @Produce json
// @Success 201
// @Router /tasks/execute [post]
func (h *apiHandler) ExecuteTasks(c echo.Context) error {
	err := h.taskUsecase.ExecuteTasks(c.Request().Context())
	if err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// ExecuteTask godoc
// @Summary タスクを実行します
// @Description
// @Tags タスク
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 201
// @Router /tasks/{id}/execute [post]
func (h *apiHandler) ExecuteTask(c echo.Context) error {
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.taskUsecase.ExecuteTask(c.Request().Context(), id)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, resp)
}

func handleError(c echo.Context, err error) error {
	// ログ出力
	slog.Error("Handler error",
		"path", c.Request().URL.Path,
		"method", c.Request().Method,
		"error", err.Error(),
	)

	// エラータイプに応じてステータスコードを決定
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Error:   "not_found",
			Message: "The requested resource was not found",
		})

	case errors.Is(err, domain.ErrAlreadyExists):
		return c.JSON(http.StatusConflict, domain.ErrorResponse{
			Error:   "already_exists",
			Message: "The resource already exists",
		})

	case errors.Is(err, domain.ErrConflict):
		return c.JSON(http.StatusConflict, domain.ErrorResponse{
			Error:   "conflict",
			Message: "Resource conflict occurred",
		})

	case errors.Is(err, domain.ErrValidation):
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})

	case errors.Is(err, domain.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "invalid_input",
			Message: err.Error(),
		})

	case errors.Is(err, domain.ErrExternalAPI):
		return c.JSON(http.StatusBadGateway, domain.ErrorResponse{
			Error:   "external_api_error",
			Message: "External service is unavailable",
		})

	case errors.Is(err, domain.ErrTimeout):
		return c.JSON(http.StatusGatewayTimeout, domain.ErrorResponse{
			Error:   "timeout",
			Message: "Request timed out",
		})

	case errors.Is(err, domain.ErrDatabase):
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "database_error",
			Message: "Database operation failed",
		})

	case errors.Is(err, domain.ErrTransaction):
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "transaction_error",
			Message: "Transaction operation failed",
		})

	case errors.Is(err, domain.ErrUnauthorized):
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})

	case errors.Is(err, domain.ErrForbidden):
		return c.JSON(http.StatusForbidden, domain.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied",
		})

	default:
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "internal_error",
			Message: "An unexpected error occurred",
		})
	}
}
