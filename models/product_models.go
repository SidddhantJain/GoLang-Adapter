package models

type Product struct {
    ProductID string  `json:"product_id"`
    Name      string  `json:"name"`
    Price     float64 `json:"price"`
    Stock     int     `json:"stock"`
}
