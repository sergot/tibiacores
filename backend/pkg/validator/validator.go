package validator

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator is a custom validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the given struct
func (v *CustomValidator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// New creates a new validator for Echo
func New() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}
