package utils

import (
    "log"
)

func LogInfo(message string) {
    log.Printf("INFO: %s", message)
}

func LogError(err error) {
    log.Printf("ERROR: %v", err)
}
