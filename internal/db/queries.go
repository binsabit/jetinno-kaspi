package db

import (
	"context"
	"time"
)

type Order struct {
	ID               int64
	OrderNo          string
	VendingMachineID int64
	ProductID        int64
	QRType           string
	Amount           float32
	CreatedAt        time.Time
	UpdatedAt        time.Time
	TxnID            uint64
	TxnDate          string
	Status           bool
	Comment          string
	TxnSum           float64
	Paid             bool
}

func (d *Database) CreateOrder(ctx context.Context, order Order) error {
	query := `INSERT INTO orders 
				(order_no, vending_machine_id, product_id, qr_type, amount) 
				VALUES				
				($1,$2,$3,$4,$5);`

	params := []interface{}{order.OrderNo, order.VendingMachineID, order.ProductID, order.QRType, order.Amount}

	_, err := d.db.Exec(ctx, query, params...)

	return err
}

func (d *Database) GetOrder(ctx context.Context, vmcNo int64, orderNo string) (Order, error) {
	query := `SELECT qr_type, paid, amount FROM orders WHERE order_no = $1 AND vending_machine_id = $2`

	var order Order

	err := d.db.QueryRow(ctx, query, vmcNo, order).Scan(&order.QRType, &order.Paid, &order.Amount)

	return order, err
}

func (d *Database) UpdateOrder(ctx context.Context, vmcNo int64, orderNo string) error {
	query := `UPDATE orders
              SET status = true,
                  updated_at = now()
			WHERE order_no = $1 AND vending_machine_id = $2`

	_, err := d.db.Exec(ctx, query, orderNo, vmcNo)
	return err
}
