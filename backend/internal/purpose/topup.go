package purpose

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type TopupHandler struct{}

func (TopupHandler) Apply(context.Context, pgx.Tx, Payment) error {
	return nil
}
