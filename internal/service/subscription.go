package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"subservice/internal/models"
	"subservice/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, sub *models.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByFilter(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error)
	SumCostByFilter(ctx context.Context, filter *models.CostCalculationFilter) (int, error)
}

type SubscriptionService struct {
	repo SubscriptionRepository
	lg   *slog.Logger
}

func NewSubscriptionService(repo SubscriptionRepository, lg *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: repo, lg: lg}
}

func (s *SubscriptionService) Create(ctx context.Context, sub models.Subscription) (*models.Subscription, error) {
	if err := s.repo.Create(ctx, &sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	s.lg.Info("flight plan created",
		slog.String("op", "SubscriptionService.Create"),
		slog.Any("id", sub.ID),
	)

	return &sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, models.ErrSubscriptionNotFound
		}
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) ListByFilter(ctx context.Context, filter models.SubscriptionFilter) ([]models.Subscription, error) {
	subs, err := s.repo.ListByFilter(ctx, &filter)
	if err != nil {
		return nil, err
	}

	return subs, nil
}

func (s *SubscriptionService) Update(ctx context.Context, sub models.Subscription) (*models.Subscription, error) {
	if _, err := s.repo.GetByID(ctx, sub.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, models.ErrSubscriptionNotFound
		}
		return nil, err
	}

	if err := s.repo.Update(ctx, &sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	s.lg.Info("subscription updated",
		slog.String("op", "SubscriptionService.Update"),
		slog.Any("id", sub.ID),
	)

	return &sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return models.ErrSubscriptionNotFound
		}
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	s.lg.Info("subscription deleted",
		slog.String("op", "SubscriptionService.Delete"),
		slog.Any("id", id),
	)

	return nil
}

func (s *SubscriptionService) GetTotalCost(ctx context.Context, filter *models.CostCalculationFilter) (int, error) {
	total, err := s.repo.SumCostByFilter(ctx, filter)

	if err != nil {
		return 0, fmt.Errorf("failed to calculate subscriptions cost: %w", err)
	}

	return total, nil
}
