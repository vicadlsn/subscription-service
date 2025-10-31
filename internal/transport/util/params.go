package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ParsePathUUID(r *http.Request, name string) (uuid.UUID, error) {
	paramStr := chi.URLParam(r, name)
	param, err := uuid.Parse(paramStr)
	if err != nil {
		return uuid.UUID{}, err
	}
	return param, nil
}

func ParseUUID(q url.Values, param string) (*uuid.UUID, error) {
	if str := q.Get(param); str != "" {
		id, err := uuid.Parse(str)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %w", param, err)
		}
		return &id, nil
	}
	return nil, nil
}

func ParseString(q url.Values, param string) *string {
	if str := q.Get(param); str != "" {
		return &str
	}
	return nil
}

func ParseMonthDate(q url.Values, param string) (*time.Time, error) {
	if str := q.Get(param); str != "" {
        t, err := time.Parse("01-2006", str)
		if err != nil {
			return nil, fmt.Errorf("invalid %s format: %w", param, err)
		}
		return &t, nil
	}
	return nil, nil
}

func ParseRequiredMonthDate(q url.Values, param string) (time.Time, error) {
	str := q.Get(param)
	if str == "" {
		return time.Time{}, fmt.Errorf("%s is required", param)
	}
    t, err := time.Parse("01-2006", str)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid %s format: %w", param, err)
	}
	return t, nil
}

func ParseInt(q url.Values, param string, defaultValue, min, max int) (int, error) {
	str := q.Get(param)
	if str == "" {
		return defaultValue, nil
	}

	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", param, err)
	}

	if val < min {
		return 0, fmt.Errorf("%s must be gte %d", param, min)
	}

	if val > max {
		return 0, fmt.Errorf("%s must be lte %d", param, max)
	}

	return val, nil
}
