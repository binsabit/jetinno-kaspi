package db

import (
	"context"
	"time"
)

type Order struct {
	ID               int64
	OrderNo          string
	VendingMachineNo string
	VendingMachineID int64
	ProductID        int64
	QRType           string
	Amount           float32
	CreatedAt        time.Time
	UpdatedAt        time.Time
	TxnID            uint64
	TxnDate          string
	Status           int
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

func (d *Database) GetOrder(ctx context.Context, vmcNo string, orderNo string) (Order, error) {
	query := `SELECT qr_type, paid, amount, orders.status 
			FROM orders	
			JOIN vending_machines on vending_machines.id = orders.vending_machine_id 
			 WHERE orders.order_no = $1 AND  vending_machines.no= $2`

	var order Order

	err := d.db.QueryRow(ctx, query, orderNo, vmcNo).Scan(&order.QRType, &order.Paid, &order.Amount, &order.Status)

	return order, err
}

func (d *Database) GetOrderByID(ctx context.Context, orderID int64) (Order, error) {
	query := `SELECT orders.id, qr_type, paid, amount, orders.status, orders.no, vending_machines.no 
			FROM orders	
			JOIN vending_machines on vending_machines.id = orders.vending_machine_id 
			 WHERE orders.id = $1`

	var order Order

	err := d.db.QueryRow(ctx, query, orderID).Scan(&order.ID, &order.QRType, &order.Paid, &order.Amount, &order.Status, &order.OrderNo, &order.VendingMachineNo)

	return order, err
}

func (d *Database) UpdateOrder(ctx context.Context, vmcID int64, orderNo string, status int) error {
	query := `UPDATE orders 
              SET status = $1,
                  updated_at = now()
			WHERE order_no = $2 AND vending_machine_id = $3`

	_, err := d.db.Exec(ctx, query, status, orderNo, vmcID)
	return err
}
