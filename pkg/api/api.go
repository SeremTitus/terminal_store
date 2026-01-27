package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"terminal_store/pkg/models"
)

type API struct {
	db *sql.DB
}

type createProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type updateStockRequest struct {
	Stock int `json:"stock"`
}

type createCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type createOrderItem struct {
	ProductID int64 `json:"product_id"`
	Qty       int   `json:"qty"`
}

type createOrderRequest struct {
	CustomerID int64             `json:"customer_id"`
	Items      []createOrderItem `json:"items"`
}

func Register(r *gin.Engine, db *sql.DB) {
	api := &API{db: db}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.GET("/products", api.listProducts)
	r.POST("/products", api.createProduct)
	r.PATCH("/products/:id/stock", api.updateStock)

	r.GET("/customers", api.listCustomers)
	r.POST("/customers", api.createCustomer)

	r.GET("/orders", api.listOrders)
	r.POST("/orders", api.createOrder)
}

func (a *API) listProducts(c *gin.Context) {
	rows, err := a.db.Query("SELECT id, name, price, stock, created_at FROM products ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	products := make([]models.Product, 0)
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, p)
	}
	c.JSON(http.StatusOK, products)
}

func (a *API) createProduct(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.Name == "" || req.Price < 0 || req.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product fields"})
		return
	}

	var p models.Product
	p.Name = req.Name
	p.Price = req.Price
	p.Stock = req.Stock
	err := a.db.QueryRow(
		"INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id, created_at",
		p.Name, p.Price, p.Stock,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (a *API) updateStock(c *gin.Context) {
	idText := c.Param("id")
	id, err := strconv.ParseInt(idText, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	var req updateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.Stock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stock must be >= 0"})
		return
	}

	var p models.Product
	err = a.db.QueryRow(
		"UPDATE products SET stock=$1 WHERE id=$2 RETURNING id, name, price, stock, created_at",
		req.Stock, id,
	).Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (a *API) listCustomers(c *gin.Context) {
	rows, err := a.db.Query("SELECT id, name, phone, created_at FROM customers ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	customers := make([]models.Customer, 0)
	for rows.Next() {
		var cu models.Customer
		if err := rows.Scan(&cu.ID, &cu.Name, &cu.Phone, &cu.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		customers = append(customers, cu)
	}
	c.JSON(http.StatusOK, customers)
}

func (a *API) createCustomer(c *gin.Context) {
	var req createCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	var cu models.Customer
	cu.Name = req.Name
	cu.Phone = req.Phone
	err := a.db.QueryRow(
		"INSERT INTO customers (name, phone) VALUES ($1, $2) RETURNING id, created_at",
		cu.Name, cu.Phone,
	).Scan(&cu.ID, &cu.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, cu)
}

func (a *API) listOrders(c *gin.Context) {
	rows, err := a.db.Query("SELECT id, customer_id, created_at FROM orders ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	orders := make([]models.Order, 0)
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items, err := a.orderItems(o.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		o.Items = items
		orders = append(orders, o)
	}
	c.JSON(http.StatusOK, orders)
}

func (a *API) orderItems(orderID int64) ([]models.OrderItem, error) {
	rows, err := a.db.Query("SELECT id, order_id, product_id, qty, price_each FROM order_items WHERE order_id=$1 ORDER BY id", orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.OrderItem, 0)
	for rows.Next() {
		var it models.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.Qty, &it.PriceEach); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, nil
}

func (a *API) createOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.CustomerID <= 0 || len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id and items required"})
		return
	}
	for _, it := range req.Items {
		if it.ProductID <= 0 || it.Qty <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item in order"})
			return
		}
	}

	tx, err := a.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRow("SELECT true FROM customers WHERE id=$1", req.CustomerID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	order.CustomerID = req.CustomerID
	if err := tx.QueryRow("INSERT INTO orders (customer_id) VALUES ($1) RETURNING id, created_at", req.CustomerID).Scan(&order.ID, &order.CreatedAt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	order.Items = make([]models.OrderItem, 0, len(req.Items))
	for _, it := range req.Items {
		var price float64
		var stock int
		err := tx.QueryRow("SELECT price, stock FROM products WHERE id=$1 FOR UPDATE", it.ProductID).Scan(&price, &stock)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "product not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if stock < it.Qty {
			c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient stock"})
			return
		}
		if _, err := tx.Exec("UPDATE products SET stock=stock-$1 WHERE id=$2", it.Qty, it.ProductID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var item models.OrderItem
		item.OrderID = order.ID
		item.ProductID = it.ProductID
		item.Qty = it.Qty
		item.PriceEach = price
		if err := tx.QueryRow(
			"INSERT INTO order_items (order_id, product_id, qty, price_each) VALUES ($1, $2, $3, $4) RETURNING id",
			order.ID, it.ProductID, it.Qty, price,
		).Scan(&item.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		order.Items = append(order.Items, item)
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}
