package services

import (
    "adapter-project/api"
    "adapter-project/models"
)

type ProductService struct {
    api *api.ProductAPI
}

func NewProductService(baseURL string) *ProductService {
    return &ProductService{
        api: api.NewProductAPI(baseURL),
    }
}

func (s *ProductService) GetProduct(productID string) (models.Product, error) {
    productData, err := s.api.GetProduct(productID)
    if err != nil {
        return models.Product{}, err
    }
    // Convert productData to models.Product
    return models.Product{
        ProductID: productID,
        Name:      productData["name"].(string),
    }, nil
}
