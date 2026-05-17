package handler

import (
	"net/http"

	"github.com/eghbalii/go-company-service/internal/adapter/http/dto"
	"github.com/eghbalii/go-company-service/internal/usecase/company"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CompanyHandler handles HTTP requests for company resources.
type CompanyHandler struct {
	uc company.UseCase
}

// NewCompanyHandler creates a CompanyHandler backed by the given use case.
func NewCompanyHandler(uc company.UseCase) *CompanyHandler {
	return &CompanyHandler{uc: uc}
}

// Create godoc
//
//	@Summary		Create a company
//	@Description	Creates a new company. Requires authentication.
//	@Tags			companies
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.CreateCompanyRequest	true	"Company payload"
//	@Success		201		{object}	dto.SuccessResponse{data=dto.CompanyResponse}
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		409		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/companies [post]
func (h *CompanyHandler) Create(c echo.Context) error {
	var req dto.CreateCompanyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	result, err := h.uc.Create(c.Request().Context(), req.ToCreateInput())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.OK(dto.ToCompanyResponse(result)))
}

// Update godoc
//
//	@Summary		Update a company
//	@Description	Partially updates a company by ID. Requires authentication.
//	@Tags			companies
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Company UUID"
//	@Param			body	body		dto.UpdateCompanyRequest	true	"Fields to update"
//	@Success		200		{object}	dto.SuccessResponse{data=dto.CompanyResponse}
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Failure		404		{object}	dto.ErrorResponse
//	@Failure		422		{object}	dto.ErrorResponse
//	@Router			/companies/{id} [patch]
func (h *CompanyHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid company id")
	}

	var req dto.UpdateCompanyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	result, err := h.uc.Update(c.Request().Context(), id, req.ToUpdateInput())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.OK(dto.ToCompanyResponse(result)))
}

// Delete godoc
//
//	@Summary		Delete a company
//	@Description	Deletes a company by ID. Requires authentication.
//	@Tags			companies
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Company UUID"
//	@Success		204
//	@Failure		401	{object}	dto.ErrorResponse
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/companies/{id} [delete]
func (h *CompanyHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid company id")
	}

	if err := h.uc.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// GetByID godoc
//
//	@Summary		Get a company
//	@Description	Retrieves a single company by ID. Public endpoint.
//	@Tags			companies
//	@Produce		json
//	@Param			id	path		string	true	"Company UUID"
//	@Success		200	{object}	dto.SuccessResponse{data=dto.CompanyResponse}
//	@Failure		404	{object}	dto.ErrorResponse
//	@Router			/companies/{id} [get]
func (h *CompanyHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid company id")
	}

	result, err := h.uc.GetByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.OK(dto.ToCompanyResponse(result)))
}
