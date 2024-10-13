package services

import (
    "adapter-project/api"
    "adapter-project/models"
)

type AuthService struct {
    api *api.AuthAPI
}

func NewAuthService(baseURL string) *AuthService {
    return &AuthService{
        api: api.NewAuthAPI(baseURL),
    }
}

func (s *AuthService) Login(apiToken, apiSecret string) error {
    return s.api.Login(apiToken, apiSecret)
}
