package models

import "time"

type Product struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	CreatedAt time.Time `json:"created_at"`
}

type Customer struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Qty       int     `json:"qty"`
	PriceEach float64 `json:"price_each"`
}

type Order struct {
	ID         int64       `json:"id"`
	CustomerID int64       `json:"customer_id"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items"`
}
