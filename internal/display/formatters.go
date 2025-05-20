package display

import (
	"fmt"
	"time"

	"ws/model"
)

// GetHomeDisplayName returns a human-readable name for a home
func GetHomeDisplayName(home model.Home) string {
	if home.AppNickname != "" {
		return home.AppNickname
	} else if home.Address.Address1 != "" {
		return fmt.Sprintf("%s, %s", home.Address.Address1, home.Address.City)
	} else {
		return home.Id
	}
}

// TrimString trims a string to a maximum length
func TrimString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// GetStatusEmoji returns an emoji representing active/inactive status
func GetStatusEmoji(isActive bool) string {
	if isActive {
		return "🟢"
	}
	return "🔴"
}

// GetPriceLevelEmoji returns an emoji based on price level
func GetPriceLevelEmoji(level string) string {
	switch level {
	case "VERY_CHEAP":
		return "🟢"
	case "CHEAP":
		return "🟩"
	case "NORMAL":
		return "⬜"
	case "EXPENSIVE":
		return "🟧"
	case "VERY_EXPENSIVE":
		return "🔴"
	default:
		return "⚪"
	}
}

// FormatDateTime formats a time.Time for display
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}

// FormatTimeOnly formats time.Time for display as hours:minutes only
func FormatTimeOnly(t time.Time) string {
	return t.Format("15:04")
}

// GetFloat64 safely extracts a float64 from a map
func GetFloat64(data map[string]interface{}, key string) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	return 0
}
