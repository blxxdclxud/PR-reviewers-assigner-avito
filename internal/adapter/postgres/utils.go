package postgres

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

func isUniqueViolationError(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			return true
		}
	}
	return false
}
