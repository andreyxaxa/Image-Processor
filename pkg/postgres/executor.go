package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type txKey struct{}

type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (p *Postgres) GetExecutor(ctx context.Context) Executor {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return p.Pool
}

// 1) Begin Tx;
// 2) Updates ctx -> context.WithValue(Tx) && func call;
// 3) err = Tx.Rollback, ok = Tx.Commit.
func (p *Postgres) WithinTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("Postgres - WithinTransaction - p.Pool.Begin: %w", err)
	}

	err = f(context.WithValue(ctx, txKey{}, tx))
	if err != nil {
		_ = tx.Rollback(ctx)

		return fmt.Errorf("Postgres - WithinTransaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("Postgres - WithinTransaction - tx.Commit: %w", err)
	}

	return nil
}
