package postgres

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Tx struct {
	Db pgx.Tx
}

type TxReq func(tx Tx) error

type TxRunner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func ExecTx(ctx context.Context, runner TxRunner, req TxReq) error {
	pgxTx, err := runner.Begin(ctx)
	if err != nil {
		return err
	}

	tx := Tx{
		Db: pgxTx,
	}

	err = req(tx)
	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	return tx.Commit(context.Background())
}

func (p Tx) Stats() *pgxpool.Stat {
	return nil
}

func (p Tx) Begin(ctx context.Context) (pgx.Tx, error) {
	return p.Db.Begin(ctx)
}

func (p Tx) Rollback(ctx context.Context) {
	_ = p.Db.Rollback(ctx)
}

func (p Tx) Commit(ctx context.Context) error {
	return p.Db.Commit(ctx)
}

func (p Tx) Query(query string, args ...any) (pgx.Rows, error) {
	return p.Db.Query(context.Background(), query, args[:]...)
}

func (p Tx) Get(dest interface{}, query string, args ...interface{}) error {
	rows, err := p.Db.Query(context.Background(), query, args[:]...)
	if err != nil {
		return err
	}
	return pgxscan.DefaultAPI.ScanOne(dest, rows)
}

func (p Tx) Select(dest interface{}, query string, args ...interface{}) error {
	rows, err := p.Db.Query(context.Background(), query, args[:]...)
	if err != nil {
		return err
	}
	return pgxscan.DefaultAPI.ScanAll(dest, rows)
}

func (p Tx) Exec(query string, args ...any) (pgconn.CommandTag, error) {
	return p.Db.Exec(context.Background(), query, args[:]...)
}

func (p Tx) QueryRow(query string, args ...interface{}) pgx.Row {
	return p.Db.QueryRow(context.Background(), query, args[:]...)
}
