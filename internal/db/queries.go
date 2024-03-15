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

type VendingMachine struct {
	ID       int64
	No       int64
	Password string
}

func (d *Database) GetVmdIDByNo(ctx context.Context, vmcNo string) (int64, int, error) {
	query := `SELECT id,health FROM vending_machines where no = $1`

	var id int64
	var status int
	err := d.db.QueryRow(ctx, query, vmcNo).Scan(&id, &status)

	return id, status, err
}

func (d *Database) GetVcmByNo(ctx context.Context, vcmNo string) (VendingMachine, error) {
	query := `SELECT id,no,password FROM vending_machines where no = $1`

	var v VendingMachine
	err := d.db.QueryRow(ctx, query, vcmNo).Scan(&v.ID, &v.Password)
	return v, err
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
	query := `SELECT orders.id, qr_type, paid, amount, orders.status, orders.order_no, vending_machines.no, txn_id,txn_sum 
			FROM orders	
			JOIN vending_machines on vending_machines.id = orders.vending_machine_id 
			 WHERE orders.id = $1`

	var order Order

	err := d.db.QueryRow(ctx, query, orderID).Scan(&order.ID, &order.QRType, &order.Paid, &order.Amount, &order.Status, &order.OrderNo, &order.VendingMachineNo, &order.TxnID, &order.TxnSum)

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

func (d *Database) GetMachineStatus(ctx context.Context, no string) (int, error) {
	query := `select health from vending_machines where no=$1`

	var status int

	err := d.db.QueryRow(ctx, query, no).Scan(&status)
	if err != nil {
		return 0, err
	}

	return status, nil
}

func (d *Database) UpdateMachineStatus(ctx context.Context, no string, status int) error {
	query := `update vending_machines set health = $1 where no = $2`

	_, err := d.db.Exec(ctx, query, status, no)
	return err
}

func (d *Database) CreateError(ctx context.Context, id int64, code, description string) error {
	query := `insert into machine_errors (vmc_id, code, description) values($1,$2,$3)`

	_, err := d.db.Exec(ctx, query, id, code, description)
	return err

}
