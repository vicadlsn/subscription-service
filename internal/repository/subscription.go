package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"subservice/internal/models"
	"subservice/pkg/postgres"
)

type SubscriptionRepository struct {
	db postgres.PgxClient
}

func NewSubscriptionRepository(db postgres.PgxClient) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	now := time.Now()

	query := `INSERT INTO subscriptions (user_id, service_name, price, start_date, end_date, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING ID, created_at, updated_at`

	err := r.db.QueryRow(ctx, query, sub.UserID, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, now, now).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		return HandleDBError(err)
	}

	return nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	query := `UPDATE subscriptions
              SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = CURRENT_TIMESTAMP
		  WHERE id = $5 
		  RETURNING id, user_id, service_name, price, start_date, end_date, 
		          created_at, updated_at, deleted_at`

	err := r.db.QueryRow(ctx, query, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.ID).Scan(&sub.ID, &sub.UserID, &sub.ServiceName, &sub.Price,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt, &sub.DeletedAt)

	if err != nil {
		return HandleDBError(err)
	}

	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE subscriptions SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return HandleDBError(err)
	}

	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query := `SELECT id, user_id, service_name, price, start_date, end_date, created_at, updated_at, deleted_at FROM subscriptions WHERE id = $1 AND deleted_at IS NULL`

	sub := &models.Subscription{}
	err := r.db.QueryRow(ctx, query, id).Scan(&sub.ID, &sub.UserID, &sub.ServiceName, &sub.Price, &sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt, &sub.DeletedAt)

	if err != nil {
		return nil, HandleDBError(err)
	}

	return sub, nil
}

func (r *SubscriptionRepository) ListByFilter(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error) {
	query := `SELECT id, user_id, service_name, price, start_date, end_date, created_at, updated_at, deleted_at FROM subscriptions WHERE deleted_at IS NULL`

	var filterValues []interface{}
	idx := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", idx)
		filterValues = append(filterValues, *filter.UserID)
		idx++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", idx)
		filterValues = append(filterValues, *filter.ServiceName)
		idx++
	}

	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND start_date >= $%d", idx)
		filterValues = append(filterValues, *filter.StartDate)
		idx++
	}

	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND start_date <= $%d", idx)
		filterValues = append(filterValues, *filter.EndDate)
		idx++
	}

	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", filter.SortBy, filter.SortOrder, idx, idx+1)
	filterValues = append(filterValues, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, filterValues...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subs := []models.Subscription{}
	for rows.Next() {
		var s models.Subscription
		if err := rows.Scan(&s.ID, &s.UserID, &s.ServiceName, &s.Price,
			&s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}

	return subs, nil
}

func (r *SubscriptionRepository) SumCostByFilter(ctx context.Context, filter *models.CostCalculationFilter) (int, error) {
	query := `SELECT COALESCE(SUM(price), 0) as total_cost FROM subscriptions
        WHERE deleted_at IS NULL
        AND start_date <= $1
        AND (end_date IS NULL OR end_date >= $2)`

	var filterValues []interface{}
	idx := 3

	filterValues = append(filterValues, filter.EndDate, filter.StartDate)

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", idx)
		filterValues = append(filterValues, *filter.UserID)
		idx++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", idx)
		filterValues = append(filterValues, *filter.ServiceName)
	}

	totalCost := 0
	err := r.db.QueryRow(ctx, query, filterValues...).Scan(&totalCost)

	if err != nil {
		return 0, err
	}

	return totalCost, nil
}
