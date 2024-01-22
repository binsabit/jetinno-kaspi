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

func (d *Database) GetVmdIDByNo(ctx context.Context, vmcNo string) (int64, error) {
	query := `SELECT id FROM vending_machines where no = $1`

	var id int64

	err := d.db.QueryRow(ctx, query, vmcNo).Scan(&id)

	return id, err
}

func (d *Database) CreateOrder(ctx context.Context, order Order) (int64, error) {
	query := `INSERT INTO orders 
				(order_no, vending_machine_id, product_id, qr_type, amount, created_at, updated_at) 
				VALUES				
				($1,$2,$3,$4,$5,$6,$7)
				RETURNING id`
	now := time.Now()
	params := []interface{}{order.OrderNo, order.VendingMachineID, order.ProductID, order.QRType, order.Amount, now, now}
	var id int64
	err := d.db.QueryRow(ctx, query, params...).Scan(&id)

	return id, err
}

func (d *Database) GetOrder(ctx context.Context, vmcNo int64, orderNo string) (Order, error) {
	query := `SELECT qr_type, paid, amount 
			JOIN vending_machines on vending_machines.id = orders.vending_machine_id 
			FROM orders WHERE orders.order_no = $1 AND  vending_machines.no= $2`

	var order Order

	err := d.db.QueryRow(ctx, query, orderNo, vmcNo).Scan(&order.QRType, &order.Paid, &order.Amount)

	return order, err
}

func (d *Database) UpdateOrder(ctx context.Context, vmcNo int64, orderNo string) error {
	query := `UPDATE orders o
              SET o.status = true,
                  o.updated_at = now()
            	INNER JOIN vending_machines vm
              on vm.id = o.vending_machine_id
			WHERE o.order_no = $1 AND vm.no = $2`

	_, err := d.db.Exec(ctx, query, orderNo, vmcNo)
	return err
}
