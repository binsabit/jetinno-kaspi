package services

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type DBTX interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction Transaction) (*Transaction, error)
	GetTransactionBYTxnID(ctx context.Context, id int64) (*Transaction, error)
}

type repository struct {
	db DBTX
}

func NewRepository(db DBTX) TransactionRepository {
	return &repository{db: db}
}

type Transaction struct {
	ID        int64     `json:"id"`
	TxnID     int64     `json:"txn_id"`
	TxnDate   int       `json:"txn_date,omitempty"`
	Result    int       `json:"result"`
	Sum       float64   `json:"sum"`
	Comment   string    `json:"comment,omitempty"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (r repository) CreateTransaction(ctx context.Context, transaction Transaction) (*Transaction, error) {
	return &Transaction{}, nil
}

func (r repository) GetTransactionBYTxnID(ctx context.Context, id int64) (*Transaction, error) {
	return &Transaction{}, nil
}
