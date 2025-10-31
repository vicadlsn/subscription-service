package subscription

import (
	"fmt"
	"net/http"
	"strings"
	"subservice/internal/models"
	"subservice/internal/transport/util"
	"time"

	"github.com/google/uuid"
)

type SubscriptionCreateDTO struct {
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	ServiceName string    `json:"service_name" validate:"required,max=64"`
	Price       int       `json:"price" validate:"required,gt=0"`
	StartDate   string    `json:"start_date" validate:"required"`
	EndDate     *string   `json:"end_date" validate:"omitempty"`
}

type SubscriptionUpdateDTO struct {
	ServiceName string  `json:"service_name" validate:"required,max=64"`
	Price       int     `json:"price" validate:"required,gt=0"`
	StartDate   string  `json:"start_date" validate:"required"`
	EndDate     *string `json:"end_date" validate:"omitempty"`
}

type SubscriptionResponseDTO struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func parseMonthYear(s string) (time.Time, error) {
	t, err := time.Parse("01-2006", s)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}

func FromSubscriptionCreateDTO(dto SubscriptionCreateDTO) (models.Subscription, error) {
	start, err := parseMonthYear(dto.StartDate)
	if err != nil {
		return models.Subscription{}, err
	}
	var end *time.Time
	if dto.EndDate != nil {
		et, err := parseMonthYear(*dto.EndDate)
		if err != nil {
			return models.Subscription{}, err
		}
		end = &et
	}

	if end != nil && start.After(*end) {
		return models.Subscription{}, fmt.Errorf("start_date cannot be after end_date")
	}

	return models.Subscription{
		UserID:      dto.UserID,
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

func FromSubscriptionUpdateDTO(dto SubscriptionUpdateDTO, id uuid.UUID) (models.Subscription, error) {
	start, err := parseMonthYear(dto.StartDate)
	if err != nil {
		return models.Subscription{}, err
	}
	var end *time.Time
	if dto.EndDate != nil {
		et, err := parseMonthYear(*dto.EndDate)
		if err != nil {
			return models.Subscription{}, err
		}
		end = &et
	}

	if end != nil && start.After(*end) {
		return models.Subscription{}, fmt.Errorf("start_date cannot be after end_date")
	}
	
	return models.Subscription{
		ID:          id,
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

func ToSubscriptionResponseDTO(model models.Subscription) SubscriptionResponseDTO {
	dto := SubscriptionResponseDTO{
		ID:          model.ID,
		UserID:      model.UserID,
		ServiceName: model.ServiceName,
		Price:       model.Price,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	dto.StartDate = model.StartDate.Format("01-2006")

	if model.EndDate != nil {
		formattedEndDate := model.EndDate.Format("01-2006")
		dto.EndDate = &formattedEndDate
	}

	return dto
}

func SummaryFilterFromRequest(r *http.Request) (*models.CostCalculationFilter, error) {
	q := r.URL.Query()
	filter := &models.CostCalculationFilter{}

	userID, err := util.ParseUUID(q, "user_id")
	if err != nil {
		return nil, err
	}
	filter.UserID = userID

	filter.ServiceName = util.ParseString(q, "service_name")

	startDate, err := util.ParseRequiredMonthDate(q, "start_date")
	if err != nil {
		return nil, err
	}
	filter.StartDate = startDate

	endDate, err := util.ParseRequiredMonthDate(q, "end_date")
	if err != nil {
		return nil, err
	}
	filter.EndDate = endDate


	return filter, nil
}

func SubscriptionFilterFromRequest(r *http.Request) (*models.SubscriptionFilter, error) {
	q := r.URL.Query()
	filter := &models.SubscriptionFilter{}

	userID, err := util.ParseUUID(q, "user_id")
	if err != nil {
		return nil, err
	}
	filter.UserID = userID

	filter.ServiceName = util.ParseString(q, "service_name")

	startDate, err := util.ParseMonthDate(q, "start_date")
	if err != nil {
		return nil, err
	}
	filter.StartDate = startDate

	endDate, err := util.ParseMonthDate(q, "end_date")
	if err != nil {
		return nil, err
	}
	filter.EndDate = endDate

	limit, err := util.ParseInt(q, "limit", 50, 1, 1000)
	if err != nil {
		return nil, err
	}
	filter.Limit = limit

	offset, err := util.ParseInt(q, "offset", 0, 0, 10000)
	if err != nil {
		return nil, err
	}
	filter.Offset = offset

	if sortBy := q.Get("sort_by"); sortBy != "" {
		if !models.IsSortByAllowed(sortBy) {
			return nil, fmt.Errorf("invalid sort_by")
		}
		filter.SortBy = sortBy
	} else {
		filter.SortBy = "created_at"
	}

	if sortOrder := q.Get("sort_order"); sortOrder != "" {
		sortOrder = strings.ToLower(sortOrder)
		if sortOrder != "asc" && sortOrder != "desc" {
			return nil, fmt.Errorf("invalid sort_order")
		}
		filter.SortOrder = sortOrder
	} else {
		filter.SortOrder = "asc"
	}

	return filter, nil
}
