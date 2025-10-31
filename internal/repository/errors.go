package repository

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNotFound = errors.New("not found")
)

func HandleDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %v", ErrNotFound, err)
	}

	return err
}
