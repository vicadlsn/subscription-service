package response

import (
	"errors"
	"net/http"
	"subservice/internal/models"

	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidRequestBody = errors.New("invalid request body")
	ErrValidationFailed   = errors.New("validation failed")
	ErrInvalidQueryParams = errors.New("invalid query params")
)

type ErrorResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

var errorMapping = map[error]ErrorResponse{
	ErrInvalidRequestBody: {
		Status:  http.StatusBadRequest,
		Message: "Некорректное тело запроса",
	},
	ErrValidationFailed: {
		Status:  http.StatusUnprocessableEntity,
		Message: "Ошибка валидации",
	},
	ErrInvalidQueryParams: {
		Status:  http.StatusBadRequest,
		Message: "Неверный параметр запроса",
	},
	models.ErrSubscriptionNotFound: {
		Status:  http.StatusNotFound,
		Message: "Подписка не найдена",
	},
}

func MapError(err error) ErrorResponse {
	for mappedErr, mappedResponse := range errorMapping {
		if errors.Is(err, mappedErr) {
			return mappedResponse
		}
	}
	return ErrorResponse{
		Status:  http.StatusInternalServerError,
		Message: "Internal server error",
	}
}

func ExtractValidationDetails(err error) interface{} {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		details := make(map[string]string)
		for _, e := range validationErrors {
			details[e.Field()] = e.Tag()
		}
		return details
	}
	return nil
}
