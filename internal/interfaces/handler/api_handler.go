package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/zuxt268/sales/internal/entity"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/external"
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
	PollingDomains(c echo.Context) error
	BackupGoogleDrive(c echo.Context) error
	AnalyzeDomains(c echo.Context) error
	GetTargets(c echo.Context) error
	CreateTarget(c echo.Context) error
	UpdateTarget(c echo.Context) error
	DeleteTarget(c echo.Context) error
	DeployWordpress(c echo.Context) error
	DeployWordpressOne(c echo.Context) error
	FetchHomstaDomains(c echo.Context) error
	FetchHomstaDomainDetails(c echo.Context) error
	AssortWordpress(c echo.Context) error
	AnalyzeDomain(c echo.Context) error

	CreateHomsta(c echo.Context) error
	GetHomstas(c echo.Context) error
	GetHomsta(c echo.Context) error

	Fetch(c echo.Context) error
	Polling(c echo.Context) error
	Analyze(c echo.Context) error
	Output(c echo.Context) error
}

type apiHandler struct {
	fetchUsecase  usecase.FetchUsecase
	domainUsecase usecase.DomainUsecase
	targetUsecase usecase.TargetUsecase
	gptUsecase    usecase.GptUsecase
	deployUsecase usecase.DeployUsecase
	sheetUsecase  usecase.SheetUsecase
	growthUsecase usecase.GrowthUsecase
	homstaUsecase usecase.HomstaUsecase
	slackAdapter  adapter.SlackAdapter
}

func NewApiHandler(
	fetchUsecase usecase.FetchUsecase,
	domainUsecase usecase.DomainUsecase,
	targetUsecase usecase.TargetUsecase,
	gptUsecase usecase.GptUsecase,
	deployUsecase usecase.DeployUsecase,
	sheetUsecase usecase.SheetUsecase,
	growthUsecase usecase.GrowthUsecase,
	homstaUsecase usecase.HomstaUsecase,
	slackAdapter adapter.SlackAdapter,
) ApiHandler {
	return &apiHandler{
		fetchUsecase:  fetchUsecase,
		domainUsecase: domainUsecase,
		targetUsecase: targetUsecase,
		gptUsecase:    gptUsecase,
		deployUsecase: deployUsecase,
		sheetUsecase:  sheetUsecase,
		growthUsecase: growthUsecase,
		homstaUsecase: homstaUsecase,
		slackAdapter:  slackAdapter,
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
// @Success 202
// @Router /fetch [post]
func (h *apiHandler) FetchDomains(c echo.Context) error {
	go func() {
		h.fetchUsecase.Fetch(context.Background())
	}()
	return c.NoContent(http.StatusAccepted)
}

// PollingDomains godoc
// @Summary Polling domains
// @Description Polling domain information
// @Tags Domains
// @Accept json
// @Produce json
// @Success 202
// @Router /polling [post]
func (h *apiHandler) PollingDomains(c echo.Context) error {
	go func() {
		h.fetchUsecase.Polling(context.Background())
	}()
	return c.NoContent(http.StatusAccepted)
}

// BackupGoogleDrive godoc
// @Summary Backup domains
// @Description Polling domain information
// @Tags Domains
// @Accept json
// @Produce json
// @Success 202
// @Router /backup [post]
func (h *apiHandler) BackupGoogleDrive(c echo.Context) error {
	go func() {
		_ = h.sheetUsecase.BackupDomainsDirectly(context.Background())
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

// DeployWordpress godoc
// @Summary ワードプレスをデプロイします
// @Description
// @Tags Wordpress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeployRequest true "デプロイ情報"
// @Success 202
// @Router /external/deploy [post]
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

// FetchHomstaDomains godoc
// @Summary ストラテジードライブサーバーにあるドメインフォルダ一覧を取得します。
// @Description
// @Tags Wordpress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200
// @Router /external/fetch/domains [post]
func (h *apiHandler) FetchHomstaDomains(c echo.Context) error {
	domains, err := h.deployUsecase.FetchDomains(context.Background())
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, domains)
}

// FetchHomstaDomainDetails godoc
// @Summary ストラテジードライブサーバーにあるドメインの詳細情報を取得します
// @Description
// @Tags Wordpress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200
// @Router /external/fetch/domains/detail [post]
func (h *apiHandler) FetchHomstaDomainDetails(c echo.Context) error {
	go func() {
		err := h.deployUsecase.FetchDomainDetails(context.Background())
		if err != nil {
			fmt.Println("error fetching domain details", err.Error())
		}
	}()
	return c.NoContent(http.StatusOK)
}

// DeployWordpressOne godoc
// @Summary ワードプレスを一件デプロイします
// @Description
// @Tags Wordpress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.DeployRequest true "デプロイ情報"
// @Success 202
// @Router /external/deploy/one [post]
func (h *apiHandler) DeployWordpressOne(c echo.Context) error {
	var req request.DeployOneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := h.deployUsecase.DeployOne(c.Request().Context(), req); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// AssortWordpress godoc
// @Summary ワードプレスを整理し、スプレッドシートに出力します
// @Description
// @Tags Wordpress
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 202
// @Router /external/assort [post]
func (h *apiHandler) AssortWordpress(c echo.Context) error {
	go func() {
		h.sheetUsecase.Assort(context.Background())
	}()
	return c.NoContent(http.StatusAccepted)
}

// AnalyzeDomain godoc
// @Summary PubSubのwebhookエンドポイント
// @Tags ドメイン
// @Accept json
// @Produce json
// @Success 204
// @Router /webhook/analyze [post]
func (h *apiHandler) AnalyzeDomain(c echo.Context) error {
	// リクエストBodyを読み取ってログ出力
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		return c.JSON(http.StatusBadRequest, "Failed to read request body")
	}

	slog.Info("Received webhook request",
		"body", string(bodyBytes),
		"content-type", c.Request().Header.Get("Content-Type"),
		"content-length", len(bodyBytes))

	// Bodyを再度読めるように復元
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Pub/Sub Push の構造体
	var push struct {
		Message struct {
			Data []byte `json:"data"`
		} `json:"message"`
	}

	// Pub/Sub JSON を解析
	if err := json.Unmarshal(bodyBytes, &push); err != nil {
		slog.Error("Failed to unmarshal pubsub push", "error", err)
		return c.JSON(http.StatusBadRequest, "invalid pubsub message")
	}

	slog.Info("Decoded PubSub data", "data_json", string(push.Message.Data))

	// Base64 decode 済み JSON を DomainMessage に Unmarshal
	var domainMessage external.DomainMessage
	if err := json.Unmarshal(push.Message.Data, &domainMessage); err != nil {
		slog.Error("Failed to unmarshal domain message", "error", err, "decoded", string(push.Message.Data))
		return c.JSON(http.StatusBadRequest, "invalid domain message")
	}

	slog.Info("Successfully parsed domain message", "message", domainMessage)

	if err := h.gptUsecase.AnalyzeDomain(c.Request().Context(), &domainMessage); err != nil {
		return handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Fetch godoc
// @Summary
// @Tags ドメイン
// @Accept json
// @Produce json
// @Success 204
// @Router /growth/fetch [post]
func (h *apiHandler) Fetch(c echo.Context) error {
	err := h.growthUsecase.Fetch(c.Request().Context())
	if err != nil {
		msg := "[Fetch]\n" + err.Error()
		if err := h.slackAdapter.Send(c.Request().Context(), msg); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return handleError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// Polling godoc
// @Summary
// @Tags ドメイン
// @Accept json
// @Produce json
// @Success 204
// @Router /growth/polling [post]
func (h *apiHandler) Polling(c echo.Context) error {
	err := h.growthUsecase.Polling(context.Background())
	if err != nil {
		msg := "[Polling]\n" + err.Error()
		if err := h.slackAdapter.Send(c.Request().Context(), msg); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return handleError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// Analyze godoc
// @Summary
// @Tags ドメイン
// @Accept json
// @Produce json
// @Success 204
// @Router /growth/analyze [post]
func (h *apiHandler) Analyze(c echo.Context) error {
	// リクエストBodyを読み取ってログ出力
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		slog.Error("Failed to read request body", "error", err)
		return c.JSON(http.StatusBadRequest, "Failed to read request body")
	}

	slog.Info("Received webhook request",
		"body", string(bodyBytes),
		"content-type", c.Request().Header.Get("Content-Type"),
		"content-length", len(bodyBytes))

	// Bodyを再度読めるように復元
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Pub/Sub Push の構造体
	var push struct {
		Message struct {
			Data []byte `json:"data"`
		} `json:"message"`
	}

	// Pub/Sub JSON を解析
	if err := json.Unmarshal(bodyBytes, &push); err != nil {
		slog.Error("Failed to unmarshal pubsub push", "error", err)
		return c.JSON(http.StatusBadRequest, "invalid pubsub message")
	}

	slog.Info("Decoded PubSub data", "data_json", string(push.Message.Data))

	// Base64 decode 済み JSON を DomainMessage に Unmarshal
	var domainMessage external.DomainMessage
	if err := json.Unmarshal(push.Message.Data, &domainMessage); err != nil {
		slog.Error("Failed to unmarshal domain message", "error", err, "decoded", string(push.Message.Data))
		return c.JSON(http.StatusBadRequest, "invalid domain message")
	}

	slog.Info("Successfully parsed domain message", "message", domainMessage)

	if err := h.growthUsecase.Analyze(c.Request().Context(), &domainMessage); err != nil {
		return handleError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Output godoc
// @Summary
// @Tags ドメイン
// @Accept json
// @Produce json
// @Success 204
// @Router /growth/output [post]
func (h *apiHandler) Output(c echo.Context) error {
	err := h.sheetUsecase.BackupDomainsDirectly(context.Background())
	if err != nil {
		msg := "[Output]\n" + err.Error()
		if err := h.slackAdapter.Send(c.Request().Context(), msg); err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return handleError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
}

// CreateHomsta godoc
// @Summary Homstaを作成します
// @Tags Homsta
// @Accept json
// @Produce json
// @Param request body request.Homsta true "Homsta情報"
// @Success 201
// @Router /homstas [post]
func (h *apiHandler) CreateHomsta(c echo.Context) error {
	var req request.Homsta
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := h.homstaUsecase.CreateHomsta(c.Request().Context(), req); err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusCreated)
}

// GetHomstas godoc
// @Summary Homsta一覧を取得します
// @Tags Homsta
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} model.Homsta
// @Router /homstas [get]
func (h *apiHandler) GetHomstas(c echo.Context) error {
	var limit, offset *int
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		limitVal, err := strconv.Atoi(limitStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "invalid limit parameter")
		}
		limit = &limitVal
	}
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		offsetVal, err := strconv.Atoi(offsetStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "invalid offset parameter")
		}
		offset = &offsetVal
	}

	homstas, err := h.homstaUsecase.GetHomstas(c.Request().Context(), limit, offset)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, homstas)
}

// GetHomsta godoc
// @Summary Homstaを取得します
// @Tags Homsta
// @Accept json
// @Produce json
// @Param name path string true "Name"
// @Success 200 {object} model.Homsta
// @Router /homstas/{name} [get]
func (h *apiHandler) GetHomsta(c echo.Context) error {
	name := c.Param("name")
	homsta, err := h.homstaUsecase.GetHomsta(c.Request().Context(), name)
	if err != nil {
		return handleError(c, err)
	}
	return c.JSON(http.StatusOK, homsta)
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
	case errors.Is(err, entity.ErrNotFound):
		return c.JSON(http.StatusNotFound, model.ErrorResponse{
			Error:   "not_found",
			Message: "The requested resource was not found",
		})

	case errors.Is(err, entity.ErrAlreadyExists):
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "already_exists",
			Message: "The resource already exists",
		})

	case errors.Is(err, entity.ErrConflict):
		return c.JSON(http.StatusConflict, model.ErrorResponse{
			Error:   "conflict",
			Message: "Resource conflict occurred",
		})

	case errors.Is(err, entity.ErrValidation):
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})

	case errors.Is(err, entity.ErrInvalidInput):
		return c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_input",
			Message: err.Error(),
		})

	case errors.Is(err, entity.ErrExternalAPI):
		return c.JSON(http.StatusBadGateway, model.ErrorResponse{
			Error:   "external_api_error",
			Message: "External service is unavailable",
		})

	case errors.Is(err, entity.ErrTimeout):
		return c.JSON(http.StatusGatewayTimeout, model.ErrorResponse{
			Error:   "timeout",
			Message: "Request timed out",
		})

	case errors.Is(err, entity.ErrDatabase):
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "database_error",
			Message: "Database operation failed",
		})

	case errors.Is(err, entity.ErrTransaction):
		return c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "transaction_error",
			Message: "Transaction operation failed",
		})

	case errors.Is(err, entity.ErrUnauthorized):
		return c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})

	case errors.Is(err, entity.ErrForbidden):
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
