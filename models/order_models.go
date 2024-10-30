package models

type Order struct {
    OrderID     string  `json:"order_id"`
    ProductID   string  `json:"product_id"`
    Quantity    int     `json:"quantity"`
    Price       float64 `json:"price"`
    Status      string  `json:"status"`
}
