package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/usecase"

	"github.com/labstack/echo/v4"
)

type ApiHandler interface {
	GetDomains(c echo.Context) error
	UpdateDomain(c echo.Context) error
	DeleteDomain(c echo.Context) error
	FetchDomains(c echo.Context) error
	AnalyzeDomains(c echo.Context) error
}

type apiHandler struct {
	fetchUsecase usecase.FetchUsecase
	pageUsecase  usecase.PageUsecase
	gptUsecase   usecase.GptUsecase
}

func NewApiHandler(
	fetchUsecase usecase.FetchUsecase,
	pageUsecase usecase.PageUsecase,
	gptUsecase usecase.GptUsecase,
) ApiHandler {
	return &apiHandler{
		fetchUsecase: fetchUsecase,
		pageUsecase:  pageUsecase,
		gptUsecase:   gptUsecase,
	}
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
// @Param is_send query boolean false "mapsで問い合わせページを開いたか"
// @Param owner_id query string false "owner_id"
// @Param status query string false "ステータス"
// @Param industry query string false "業種"
// @Param is_ssl query boolean false "SSL対応可否"
// @Success 200 {array} domain.Domain
// @Security Bearer
// @Router /domains [get]
func (h *apiHandler) GetDomains(c echo.Context) error {
	var req domain.GetDomainsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.pageUsecase.GetDomains(c.Request().Context(), req)
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
// @Security Bearer
// @Router /domains/{id} [put]
func (h *apiHandler) UpdateDomain(c echo.Context) error {
	var req domain.UpdateDomainRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return handleError(c, err)
	}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	resp, err := h.pageUsecase.UpdateDomain(c.Request().Context(), id, req)
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
// @Security Bearer
// @Router /domains/{id} [delete]
func (h *apiHandler) DeleteDomain(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	err = h.pageUsecase.DeleteDomain(c.Request().Context(), id)
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
// @Security Bearer
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
// @Security Bearer
// @Router /fetch [post]
func (h *apiHandler) FetchDomains(c echo.Context) error {
	var req domain.PostFetchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return handleError(c, err)
	}
	err := h.fetchUsecase.Fetch(c.Request().Context(), req)
	if err != nil {
		return handleError(c, err)
	}
	return c.NoContent(http.StatusAccepted)
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
