package cmd

import "encoding/json"

func rawPayload(data json.RawMessage) any {
	if m := unmarshalMap(data); len(m) > 0 {
		return m
	}
	return unmarshalList(data)
}

func listFromResponse(data json.RawMessage, key string) []map[string]any {
	if list := unmarshalList(data); len(list) > 0 {
		return list
	}
	return listFromValue(unmarshalMap(data)[key])
}

func listFromValue(value any) []map[string]any {
	switch v := value.(type) {
	case []map[string]any:
		return v
	case []any:
		out := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

func stringListFromValue(value any) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, str(map[string]any{"value": item}, "value"))
		}
		return out
	default:
		return nil
	}
}
