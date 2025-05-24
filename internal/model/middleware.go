// internal/model/middleware.go
package model

import (
    "encoding/json"
    "time"
    
    "tibber_loader/internal/utils"
)

// TimeConverterMiddleware converteert timestamps in structs
type TimeConverterMiddleware struct{}

// MarshalJSON converteert timestamps naar ISO 8601
func (tcm *TimeConverterMiddleware) MarshalJSON() ([]byte, error) {
    // Converteer alle time.Time velden naar ISO 8601
    return json.Marshal(tcm)
}

// UnmarshalJSON converteert ISO 8601 naar time.Time
func (tcm *TimeConverterMiddleware) UnmarshalJSON(data []byte) error {
    // Converteer alle ISO 8601 strings naar time.Time
    return json.Unmarshal(data, tcm)
}