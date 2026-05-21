package cmd

import "encoding/json"

// unmarshalNamedList handles endpoints that wrap arrays in an object envelope.
func unmarshalNamedList(data json.RawMessage, key string) []map[string]any {
	if list := unmarshalList(data); list != nil {
		return list
	}
	var envelope map[string]any
	if err := json.Unmarshal(data, &envelope); err == nil {
		return mapsFromAnyList(envelope[key])
	}
	return nil
}

func mapsFromAnyList(value any) []map[string]any {
	switch list := value.(type) {
	case []map[string]any:
		return list
	case []any:
		out := make([]map[string]any, 0, len(list))
		for _, item := range list {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

func strFirst(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v := str(m, key); v != "" {
			return v
		}
	}
	return ""
}
