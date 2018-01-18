package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompliantData(t *testing.T) {
	testCases := []struct {
		schema    Schema
		data      Data
		compliant bool
	}{
		{
			map[string]string{"field": "number"},
			map[string]interface{}{"field": interface{}(true)},
			false,
		},
		{
			map[string]string{"field": "bool"},
			map[string]interface{}{
				"field": interface{}(false),
				"asdf":  interface{}(4),
			},
			true,
		},
		{
			map[string]string{"field": "bool", "asdf": "number"},
			map[string]interface{}{"field": interface{}(false)},
			true,
		},
		{
			map[string]string{
				"field":  "number",
				"field2": "bool",
				"field3": "number",
			},
			map[string]interface{}{
				"field":  interface{}(3.14159),
				"field2": interface{}(true),
				"field3": interface{}(3),
			},
			true,
		},
	}

	for _, tt := range testCases {
		assert.Equal(t, tt.compliant, CompliantData(tt.schema, tt.data))
	}
}

func TestAverage(t *testing.T) {
	testCases := []struct {
		schema  Schema
		data    []Data
		err     error
		results Results
	}{
		{
			map[string]string{"field": "number", "field2": "bool"},
			[]Data{
				map[string]interface{}{
					"field":  interface{}(4),
					"field2": interface{}(false),
				},
				map[string]interface{}{
					"field":  interface{}(5),
					"field2": interface{}(true),
				},
				map[string]interface{}{
					"field":  interface{}(4),
					"field2": interface{}(false),
				},
				map[string]interface{}{
					"field":  interface{}(5),
					"field2": interface{}(true),
				},
			},
			nil,
			map[string]float64{"field": 4.5, "field2": 0.5},
		},
		{
			map[string]string{"field2": "asdf"},
			[]Data{},
			ErrUnsupportedSchemaType,
			map[string]float64{},
		},
	}

	for _, tt := range testCases {
		results, err := Average(tt.schema, tt.data...)

		if !assert.Equal(t, tt.err, err) {
			t.FailNow()
		}

		assert.Equal(t, tt.results, results)
	}
}
