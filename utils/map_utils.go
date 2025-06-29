package utils

// Helper functions untuk ekstraksi data dari map
func GetString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func GetInt32(data map[string]interface{}, key string) int32 {
	switch val := data[key].(type) {
	case int32:
		return val
	case int:
		return int32(val)
	case int64:
		return int32(val)
	case float32:
		return int32(val)
	case float64:
		return int32(val)
	}
	return 0
}

func GetBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}