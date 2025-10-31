package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

type Subscription struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ServiceName string
	Price       int
	StartDate   time.Time
	EndDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type CostCalculationFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	StartDate   time.Time
	EndDate     time.Time
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName *string
	StartDate   *time.Time
	EndDate     *time.Time
	Limit       int
	Offset      int
	SortBy      string
	SortOrder   string
}

var allowedSortBy = map[string]struct{}{
	"start_date": {},
	"end_date":   {},
	"price":      {},
}

func IsSortByAllowed(sortBy string) bool {
	if _, ok := allowedSortBy[sortBy]; !ok {
		return false
	}
	return true
}
