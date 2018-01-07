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

// CompliantData returns whether the given data complies to the schema. If the data has a field that is in the
// schema but does not match the type specified in the schema, then it is considered invalid. If there
// is a field in the data that is missing, (is present in the schema but not in the data), the data is still
// considered valid.
func CompliantData(schema Schema, data Data) bool {
	for k, v := range schema {
		dv, ok := data[k]
		if !ok {
			continue // incomplete report, still ok just skip type checking
		}

		if v == "number" {
			_, okfloat := dv.(float64)
			_, okint := dv.(int)
			if !okfloat && !okint { // field is not a "number" (float64 or int), not ok
				return false
			}
		} else if v == "bool" {
			if _, ok := dv.(bool); !ok { // field is not a "bool", not ok
				return false
			}
		}
	}

	return true
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
				val, ok := datum[k]
				if !ok {
					continue
				}

				switch value := val.(type) {
				case int:
					sum += float64(value)
				case float64:
					sum += value
				}
			}

		case "bool":
			for _, datum := range data {
				val, ok := datum[k]
				if !ok {
					continue
				}

				switch value := val.(type) {
				case bool:
					sum += float64(btoi(value))
				}
			}

		default:
			return results, fmt.Errorf("analysis: unsupported schema type")
		}

		results[k] = sum / float64(len(data))
	}

	return results, nil
}
