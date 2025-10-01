package handler

import (
	"errors"
	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/usecase"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ApiHandler interface {
	GetDomains(c echo.Context) error
	UpdateDomain(c echo.Context) error
	DeleteDomain(c echo.Context) error
	FetchDomains(c echo.Context) error
}

type apiHandler struct {
	fetchUsecase usecase.FetchUsecase
	pageUsecase  usecase.PageUsecase
}

func NewApiHandler(fetchUsecase usecase.FetchUsecase, pageUsecase usecase.PageUsecase) ApiHandler {
	return &apiHandler{
		fetchUsecase: fetchUsecase,
		pageUsecase:  pageUsecase,
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
// @Param name query string false "Domain name"
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
// @Param request body domain.UpdateDomainRequest true "Update domain request"
// @Success 200 {object} domain.Domain
// @Security Bearer
// @Router /domain [put]
func (h *apiHandler) UpdateDomain(c echo.Context) error {
	var req domain.UpdateDomainRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return handleError(c, err)
	}
	resp, err := h.pageUsecase.UpdateDomains(c.Request().Context(), req)
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
// @Param request body domain.DeleteDomainRequest true "Delete domain request"
// @Success 204
// @Security Bearer
// @Router /domain [delete]
func (h *apiHandler) DeleteDomain(c echo.Context) error {
	var req domain.DeleteDomainRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err := req.Validate(); err != nil {
		return handleError(c, err)
	}
	err := h.pageUsecase.DeleteDomains(c.Request().Context(), req)
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
