package utils

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func PostRequest(url string, payload interface{}) (*http.Response, error) {
    jsonData, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    return http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}
