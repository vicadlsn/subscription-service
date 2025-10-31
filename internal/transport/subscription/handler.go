package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"subservice/internal/models"
	"subservice/internal/transport/response"
	"subservice/internal/transport/util"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type SubscriptionService interface {
	Create(ctx context.Context, model models.Subscription) (*models.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, model models.Subscription) (*models.Subscription, error)

	Delete(ctx context.Context, id uuid.UUID) error
	ListByFilter(ctx context.Context, filters models.SubscriptionFilter) ([]models.Subscription, error)
	GetTotalCost(ctx context.Context, filters *models.CostCalculationFilter) (int, error)
}

type SubscriptionHandler struct {
	service   SubscriptionService
	lg        *slog.Logger
	validator *validator.Validate
}

func NewSubscriptionHandler(service SubscriptionService, lg *slog.Logger, validator *validator.Validate) *SubscriptionHandler {
	return &SubscriptionHandler{service: service, lg: lg, validator: validator}
}

// @Summary Создать новую подписку
// @Description Создает новую запись о подписке пользователя.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body SubscriptionCreateDTO true "Данные для создания подписки"
// @Success 201 {object} SubscriptionDTO  "Подписка создана"
// @Failure 400 {object} response.ErrorResponse "Неверный формат запроса"
// @Failure 422 {object} response.ErrorResponse "Ошибка валидации"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) error {
	op := "SubscriptionHandler.Create"
	log := h.lg.With(slog.String("op", op))

	var dto SubscriptionCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		log.Debug("failed to decode request body", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrInvalidRequestBody, err)
	}

	if err := h.validator.Struct(dto); err != nil {
		log.Debug("validation failed", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrValidationFailed, err)
	}

	model, err := FromSubscriptionCreateDTO(dto)
	if err != nil {
		log.Debug("invalid start/end date format", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrValidationFailed, err)
	}
	createdModel, err := h.service.Create(r.Context(), model)

	if err != nil {
		log.Error("failed to create subscription",
			slog.Any("error", err),
		)
		return err
	}

	response := ToSubscriptionDTO(*createdModel)
	if err := util.WriteJSON(w, http.StatusCreated, response); err != nil {
		log.Error("Failed to write response", slog.Any("error", err))
		return err
	}

	return nil
}

// @Summary Получить подписку по ID
// @Description Возвращает информаци о подписке по её UUID.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 200 {object} SubscriptionDTO "Подписка получена"
// @Failure 400 {object} response.ErrorResponse "Неверный ID подписки"
// @Failure 404 {object} response.ErrorResponse "Подписка не найдена"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) error {
	op := "SubscriptionHandler.GetByID"
	log := h.lg.With(slog.String("op", op))

	id, err := util.ParsePathUUID(r, "id")
	if err != nil {
		log.Debug("invalid subscription id", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrInvalidQueryParams, err)
	}

	model, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		log.Error("failed to get subscription",
			slog.Any("id", id),
			slog.Any("error", err),
		)
		return err
	}

	response := ToSubscriptionDTO(*model)
	if err := util.WriteJSON(w, http.StatusOK, response); err != nil {
		log.Error("Failed to write response", slog.Any("error", err))
		return err
	}

	return nil
}

// @Summary Список подписок с фильтрацией
// @Description Возвращает список подписок с возможностью фильтртации, пагинации и сортировки
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string false "Фильтр по ID пользователя (UUID)"
// @Param service_name query string false "Фильтр по названию сервиса"
// @Param start_date query string false "Фильтр по дате начала подписки (ММ-ГГГГ)"
// @Param end_date query string false "Фильтр по дате окончания подписки (ММ-ГГГГ)"
// @Param limit query int false "Количество получаемых записей (по умолчанию 50)" minimum(1) maximum(1000)
// @Param offset query int false "Отступ (по умолчанию 0)" minimum(0)
// @Param sort_by query string false "Поле для сортировки" Enums(start_date,end_date,price,created_at)
// @Param sort_order query string false "Порядок сортировки" Enums(asc,desc)
// @Success 200 {array} SubscriptionDTO "Список подписок получен"
// @Failure 400 {object} response.ErrorResponse "Неверные параметры запроса"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions [get]
func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) error {
	op := "SubscriptionHandler.GetAll"
	log := h.lg.With(slog.String("op", op))

	filters, err := SubscriptionFilterFromRequest(r)
	if err != nil {
		log.Debug("Invalid list filters", slog.Any("error", err))
		return response.ErrInvalidQueryParams
	}

	models, err := h.service.ListByFilter(r.Context(), *filters)
	if err != nil {
		log.Error("failed to get subscriptions",
			slog.Any("error", err),
		)
		return err
	}

	responses := make([]SubscriptionDTO, 0, len(models))
	for _, m := range models {
		responses = append(responses, ToSubscriptionDTO(m))
	}

	if err := util.WriteJSON(w, http.StatusOK, responses); err != nil {
		log.Error("Failed to write response", slog.Any("error", err))
		return err
	}

	return nil
}

// @Summary Обновить подписку
// @Description Обновляет существующую запись о подписке по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Param subscription body SubscriptionUpdateDTO true "Данные для обновления подписки"
// @Success 200 {object} SubscriptionDTO "Подписка обновлена"
// @Failure 400 {object} response.ErrorResponse "Неверное тело запроса или неверный ID"
// @Failure 404 {object} response.ErrorResponse "Подписка не найдена"
// @Failure 422 {object} response.ErrorResponse "Ошибка валидации"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [put]
func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) error {
	op := "SubscriptionHandler.Update"

	id, err := util.ParsePathUUID(r, "id")
	if err != nil {
		h.lg.Debug("invalid subscription id", slog.String("op", op), slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrInvalidQueryParams, err)
	}

	log := h.lg.With(slog.String("op", op), slog.Any("id", id))

	var dto SubscriptionUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		log.Debug("failed to decode request body", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrInvalidRequestBody, err)
	}

	if err := h.validator.Struct(dto); err != nil {
		log.Debug("validation failed", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrValidationFailed, err)
	}

	model, err := FromSubscriptionUpdateDTO(dto, id)
	if err != nil {
		log.Debug("invalid start/end date format", slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrValidationFailed, err)
	}
	updatedModel, err := h.service.Update(r.Context(), model)
	if err != nil {
		log.Error("failed to update subscription",
			slog.Any("error", err),
		)
		return err
	}

	response := ToSubscriptionDTO(*updatedModel)
	if err := util.WriteJSON(w, http.StatusOK, response); err != nil {
		log.Error("Failed to write response", slog.Any("error", err))
		return err
	}

	return nil
}

// @Summary Удалить подписку
// @Description Выполняет мягкое удаление записи о подписке по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки (UUID)"
// @Success 204 "Подписка удалена"
// @Failure 400 {object} response.ErrorResponse "Неверный ID подписки"
// @Failure 404 {object} response.ErrorResponse "Подписка не найдена"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	op := "SubscriptionHandler.Delete"

	id, err := util.ParsePathUUID(r, "id")
	if err != nil {
		h.lg.Debug("invalid subscription id", slog.String("op", op), slog.String("error", err.Error()))
		return fmt.Errorf("%w: %v", response.ErrInvalidQueryParams, err)
	}

	log := h.lg.With(slog.String("op", op), slog.Any("id", id))

	err = h.service.Delete(r.Context(), id)
	if err != nil {
		log.Error("failed to delete subscription",
			slog.Any("error", err),
		)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	return nil
}

// @Summary Получить общую стоимость подписок
// @Description Рассчитывает общую стоимость всех подписок за указанный период с возможностью фильтрации по пользователю и названию сервиса
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param start_date query string true "Дата начала периода (ММ-ГГГГ)"
// @Param end_date query string true "Дата окончания периода (ММ-ГГГГ)"
// @Param user_id query string false "Фильтр по ID пользователя (UUID)"
// @Param service_name query string false "Фильтр по названию сервиса"
// @Success 200 {object} map[string]int "Общая стоимость подписок"
// @Failure 400 {object} response.ErrorResponse "Неверные параметры запроса"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /subscriptions/total [get]
func (h *SubscriptionHandler) GetTotalCost(w http.ResponseWriter, r *http.Request) error {
	op := "SubscriptionHandler.GetTotalCost"
	log := h.lg.With(slog.String("op", op))

	filter, err := SummaryFilterFromRequest(r)
	if err != nil {
		log.Debug("Invalid summary params", slog.Any("error", err))
		return response.ErrInvalidQueryParams
	}

	sum, err := h.service.GetTotalCost(r.Context(), filter)
	if err != nil {
		log.Error("failed to calculate total sum",
			slog.Any("error", err),
		)
		return err
	}

	response := map[string]int{"total": sum}
	if err := util.WriteJSON(w, http.StatusOK, response); err != nil {
		log.Error("Failed to write response", slog.Any("error", err))
		return err
	}

	return nil
}
