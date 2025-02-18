package common

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

type ctxTransactionKey struct{}

func EnsureSingleTx(
	ctx context.Context, db *bun.DB, opts *sql.TxOptions, run func(ctx context.Context, tx bun.Tx) error,
) error {
	existingTx := ctx.Value(ctxTransactionKey{})
	if existingTx != nil {
		tx, ok := existingTx.(*bun.Tx)
		if ok {
			return run(ctx, *tx)
		}
	}

	return db.RunInTx(ctx, opts, func(ctx context.Context, tx bun.Tx) error {
		ctx = context.WithValue(ctx, ctxTransactionKey{}, &tx)
		return run(ctx, tx)
	})
}
