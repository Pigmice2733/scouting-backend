package analysis

import "fmt"

// Schema defines a type for a schema, a mapping of key names to their types.
// Types supported are: "number", and "bool".
type Schema map[string]string

// Data provides a type for data, a map of key names to their values.
type Data map[string]interface{}

// Results provides a type for analysis results, a map of key names to the
// results (ex. averages).
type Results map[string]float64

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Average averages all "number" and "bool" fields. True is considered 1, and false
// is considered zero.
func Average(schema Schema, data ...Data) (Results, error) {
	results := make(Results)

	for k, v := range schema {
		var sum float64

		switch v {
		case "number":
			for _, datum := range data {
				switch value := datum[k].(type) {
				case int:
					sum += float64(value)
				case float64:
					sum += value
				default:
					return results, fmt.Errorf("analysis: data type '%T' of field '%v' does not match schema type '%v'", datum[k], datum[k], v)
				}
			}

		case "bool":
			for _, datum := range data {
				switch value := datum[k].(type) {
				case bool:
					sum += float64(btoi(value))
				default:
					return results, fmt.Errorf("analysis: data type '%T' of field '%v' does not match schema type '%v'", datum[k], datum[k], v)
				}
			}
		}

		results[k] = sum / float64(len(data))
	}
	return results, nil
}
