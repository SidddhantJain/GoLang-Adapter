package services

import (
    "adapter-project/api"
    "adapter-project/models"
)

func LoginService(apiToken, apiSecret string) (*models.LoginResponse, error) {
    return api.Login(apiToken, apiSecret)
}
