package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/zuxt268/sales/internal/interfaces/dto/request"
	_ "github.com/zuxt268/sales/internal/interfaces/dto/response"
	"github.com/zuxt268/sales/internal/model"
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
	DeployWordpress(c echo.Context) error
	AssortWordpress(c echo.Context) error
	OutputSheet(c echo.Context) error
}

type apiHandler struct {
	fetchUsecase  usecase.FetchUsecase
	domainUsecase usecase.DomainUsecase
	targetUsecase usecase.TargetUsecase
	logUsecase    usecase.LogUsecase
	gptUsecase    usecase.GptUsecase
	taskUsecase   usecase.TaskUsecase
	deployUsecase usecase.DeployUsecase
	sheetUsecase  usecase.SheetUsecase
}

func NewApiHandler(
	fetchUsecase usecase.FetchUsecase,
	domainUsecase usecase.DomainUsecase,
	targetUsecase usecase.TargetUsecase,
	logUsecase usecase.LogUsecase,
	gptUsecase usecase.GptUsecase,
	taskUsecase usecase.TaskUsecase,
	deployUsecase usecase.DeployUsecase,
	sheetUsecase usecase.SheetUsecase,
) ApiHandler {
	return &apiHandler{
		fetchUsecase:  fetchUsecase,
		domainUsecase: domainUsecase,
		targetUsecase: targetUsecase,
		logUsecase:    logUsecase,
		gptUsecase:    gptUsecase,
		taskUsecase:   taskUsecase,
		deployUsecase: deployUsecase,
		sheetUsecase:  sheetUsecase,
	}
}

// GetDomain godoc
// @Summary Get domain
// @Description Get domain
// @Tags ドメイン
// @Accept json
// @Produce json
// @Param id path string true "ID"
// @Success 200 {object} response.Domain
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
// @Success 200 {array} response.Domains
// @Router /domains [get]
func (h *apiHandler) GetDomains(c echo.Context) error {
	var req request.GetDomains
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	fmt.Println(req)
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
// @Param request body request.UpdateDomain true "更新ドメイン情報"
// @Success 200 {object} response.Domain
// @Router /domains/{id} [put]
func (h *apiHandler) UpdateDomain(c echo.Context) error {
	var req request.UpdateDomain
	if err := c.Bind(&req); err != nil {
		slog.Error(err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		slog.Error(err.Error())
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.domainUsecase.UpdateDomain(c.Request().Context(), id, req)
	if err != nil {
		slog.Error(err.Error())
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
// @Param request body model.PostFetchRequest true "Fetch request"
// @Success 202
// @Router /fetch [post]
func (h *apiHandler) FetchDomains(c echo.Context) error {
	var req model.PostFetchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return handleError(c, err)
	}
	go func() {
		h.fetchUsecase.Fetch(context.Background(), req)
	}()
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
// @Success 200 {array} model.Target
// @Router /targets [get]
func (h *apiHandler) GetTargets(c echo.Context) error {
	var req model.GetTargetsRequest
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
// @Param request body model.CreateTargetRequest true "作成ターゲット情報"
// @Success 201 {object} model.Target
// @Router /targets [post]
func (h *apiHandler) CreateTarget(c echo.Context) error {
	var req model.CreateTargetRequest
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
// @Param request body model.UpdateTargetRequest true "更新ターゲット情報"
// @Success 200 {object} model.Target
// @Router /targets/{id} [put]
func (h *apiHandler) UpdateTarget(c echo.Context) error {
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var req model.UpdateTargetRequest
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
// @Success 200 {array} response.Logs
// @Router /logs [get]
func (h *apiHandler) GetLogs(c echo.Context) error {
	var req request.GetLogs
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
// @Param request body request.CreateLog true "作成ログ情報"
// @Success 201 {object} response.Log
// @Router /logs [post]
func (h *apiHandler) CreateLog(c echo.Context) error {
	var req request.CreateLog
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
// @Success 200 {array} model.Task
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
// @Param request body model.CreateTaskRequest true "作成タスク情報"
// @Success 201 {object} model.Task
// @Router /tasks [post]
func (h *apiHandler) CreateTask(c echo.Context) error {
	var req model.CreateTaskRequest
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
// @Param request body model.UpdateTaskRequest true "更新タスク情報"
// @Success 200 {object} model.Task
// @Router /tasks/{id} [put]
func (h *apiHandler) UpdateTask(c echo.Context) error {
	var id int
	if err := echo.PathParamsBinder(c).Int("id", &id).BindError(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	var req model.UpdateTaskRequest
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

// DeployWordpress godoc
// @Summary ワードプレスをデプロイします
// @Description
// @Tags Wordpress
// @Accept json
// @Produce json
// @Param request body request.DeployRequest true "デプロイ情報"
// @Success 202
// @Router /external/api/deploy [post]
func (h *apiHandler) DeployWordpress(c echo.Context) error {
	var req request.DeployRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	go func() {
		h.deployUsecase.Deploy(context.Background(), req)
	}()

	return c.NoContent(http.StatusAccepted)
}

// AssortWordpress godoc
// @Summary ワードプレスを整理し、スプレッドシートに出力します
// @Description
// @Tags Wordpress
// @Accept json
// @Produce json
// @Success 202
// @Router /external/api/assort [post]
func (h *apiHandler) AssortWordpress(c echo.Context) error {
	go func() {
		h.deployUsecase.Assort(context.Background())
	}()
	return c.NoContent(http.StatusAccepted)
}

// OutputSheet godoc
// @Summary スプレッドシートに出力する
// @Description
// @Tags ドメイン
// @Accept json
// @Produce json
// @Success 201
// @Router /domains/output [post]
func (h *apiHandler) OutputSheet(c echo.Context) error {
	err := h.sheetUsecase.RivalSheetOutput(c.Request().Context())
	if err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusOK)
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
	case errors.Is(err, model.ErrNotFound):
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "The requested resource was not found",
		})

	case errors.Is(err, model.ErrAlreadyExists):
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "already_exists",
			Message: "The resource already exists",
		})

	case errors.Is(err, model.ErrConflict):
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "conflict",
			Message: "Resource conflict occurred",
		})

	case errors.Is(err, model.ErrValidation):
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})

	case errors.Is(err, model.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_input",
			Message: err.Error(),
		})

	case errors.Is(err, model.ErrExternalAPI):
		return c.JSON(http.StatusBadGateway, model.ErrorResponse{
			Error:   "external_api_error",
			Message: "External service is unavailable",
		})

	case errors.Is(err, model.ErrTimeout):
		return c.JSON(http.StatusGatewayTimeout, model.ErrorResponse{
			Error:   "timeout",
			Message: "Request timed out",
		})

	case errors.Is(err, model.ErrDatabase):
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "database_error",
			Message: "Database operation failed",
		})

	case errors.Is(err, model.ErrTransaction):
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "transaction_error",
			Message: "Transaction operation failed",
		})

	case errors.Is(err, model.ErrUnauthorized):
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})

	case errors.Is(err, model.ErrForbidden):
		return c.JSON(http.StatusForbidden, model.ErrorResponse{
			Error:   "forbidden",
			Message: "Access denied",
		})

	default:
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
	}
}
