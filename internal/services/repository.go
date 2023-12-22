package services

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type DBTX interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type Transaction struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	TxnID     int64     `json:"txn_id"`
	TxnDate   int64     `json:"txn_date,omitempty"`
	Result    int       `json:"result"`
	Sum       float64   `json:"sum"`
	Comment   string    `json:"comment,omitempty"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func CreateTransaction(ctx context.Context, db DBTX, txn Transaction) (int64, error) {
	query := `INSERT INTO transactions(txn_id,txn_date,result,sum,comment,status) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`

	args := []interface{}{txn.TxnID, txn, txn.TxnDate, txn.Result, txn.Result, txn.Sum, txn.Comment, txn.Status}

	var id int64

	err := db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

type UpdateArgs struct {
	Result *int
	Status *bool
}

func UpdateTransaction(ctx context.Context, db DBTX, upd UpdateArgs) error {
	query := `UPDATE transactions set 
              	status = coalesce($1,status)
              	result = coalesce($2,result);`

	_, err := db.Exec(ctx, query, upd.Status, upd.Result)
	if err != nil {
		return err
	}
	return nil
}

func GetTransactionBYTxnID(ctx context.Context, db DBTX, id int64) (*Transaction, error) {
	txn := &Transaction{}

	query := `SELECT id, txn_id,txn_date, result,sum, comment, status FROM transactions where txn_id = $1`

	err := db.QueryRow(ctx, query, id).
		Scan(&txn.ID,
			&txn.TxnID,
			&txn.TxnDate,
			&txn.Result,
			&txn.Sum,
			&txn.Comment,
			&txn.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return txn, nil
}
