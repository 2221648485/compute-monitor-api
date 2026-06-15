package k8s

import "encoding/json"

func mustJSON(value interface{}) string {
	bytes, _ := json.Marshal(value)
	return string(bytes)
}

func parseStringSlice(raw string) []string {
	var result []string
	_ = json.Unmarshal([]byte(raw), &result)
	return result
}

func parseStringMap(raw string) map[string]string {
	result := map[string]string{}
	_ = json.Unmarshal([]byte(raw), &result)
	return result
}
