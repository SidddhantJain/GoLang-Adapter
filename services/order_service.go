package services

import (
    "adapter-project/api"
    "adapter-project/models"
)

type OrderService struct {
    api *api.OrderAPI
}

func NewOrderService(baseURL string) *OrderService {
    return &OrderService{
        api: api.NewOrderAPI(baseURL),
    }
}


/*
func (s *OrderService) CreateOrder(orderData models.Order) error {
    // no logic for creation of order - to be redirected using route_initializer.go
    return s.api.CreateOrder(map[string]interface{}{
        "product_id": orderData.ProductID,
        "quantity":   orderData.Quantity,
    })
}
*/
