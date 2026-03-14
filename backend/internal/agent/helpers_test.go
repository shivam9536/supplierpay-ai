package agent

import (
	"encoding/json"
	"testing"
)

// ── toFloat ──────────────────────────────────────────────────────────────────

func TestToFloat(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    float64
		wantOK  bool
	}{
		{"float64", float64(3.14), 3.14, true},
		{"float32", float32(2.5), 2.5, true},
		{"int", int(42), 42.0, true},
		{"int64", int64(100), 100.0, true},
		{"json.Number valid", json.Number("99.9"), 99.9, true},
		{"json.Number invalid", json.Number("abc"), 0, false},
		{"nil", nil, 0, false},
		{"string", "123", 0, false},
		{"zero float64", float64(0), 0, true},
		{"negative", float64(-5.5), -5.5, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toFloat(tc.input)
			if ok != tc.wantOK {
				t.Errorf("toFloat(%v) ok = %v, want %v", tc.input, ok, tc.wantOK)
			}
			if ok && got != tc.want {
				t.Errorf("toFloat(%v) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// ── abs64 ────────────────────────────────────────────────────────────────────

func TestAbs64(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0.0, 0.0},
		{-0.001, 0.001},
		{1e9, 1e9},
	}
	for _, tc := range tests {
		got := abs64(tc.input)
		if got != tc.want {
			t.Errorf("abs64(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

// ── normaliseDesc ─────────────────────────────────────────────────────────────

func TestNormaliseDesc(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Cloud Hosting", "cloud hosting"},
		{"  Support Services  ", "support services"},
		{"UPPER CASE", "upper case"},
		{"", ""},
		{"Mixed  Case", "mixed  case"},
	}
	for _, tc := range tests {
		got := normaliseDesc(tc.input)
		if got != tc.want {
			t.Errorf("normaliseDesc(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// ── toLineItemSlice ───────────────────────────────────────────────────────────

func TestToLineItemSlice(t *testing.T) {
	t.Run("[]map[string]interface{} passthrough", func(t *testing.T) {
		input := []map[string]interface{}{
			{"description": "item1", "quantity": 1.0},
		}
		got, ok := toLineItemSlice(input)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 item, got %d", len(got))
		}
		if got[0]["description"] != "item1" {
			t.Errorf("unexpected description: %v", got[0]["description"])
		}
	})

	t.Run("[]interface{} with map elements", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{"description": "item2", "unit_price": 100.0},
			map[string]interface{}{"description": "item3", "unit_price": 200.0},
		}
		got, ok := toLineItemSlice(input)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 items, got %d", len(got))
		}
	})

	t.Run("[]interface{} with non-map element returns partial", func(t *testing.T) {
		input := []interface{}{
			map[string]interface{}{"description": "item1"},
			"not a map",
		}
		got, ok := toLineItemSlice(input)
		// ok=false because not all elements converted
		if ok {
			t.Error("expected ok=false when not all elements are maps")
		}
		if len(got) != 1 {
			t.Errorf("expected 1 converted item, got %d", len(got))
		}
	})

	t.Run("nil input", func(t *testing.T) {
		got, ok := toLineItemSlice(nil)
		if ok || got != nil {
			t.Errorf("expected nil,false; got %v,%v", got, ok)
		}
	})

	t.Run("string input", func(t *testing.T) {
		got, ok := toLineItemSlice("not a slice")
		if ok || got != nil {
			t.Errorf("expected nil,false for string input")
		}
	})

	t.Run("empty []interface{}", func(t *testing.T) {
		input := []interface{}{}
		got, ok := toLineItemSlice(input)
		if !ok {
			t.Error("expected ok=true for empty slice")
		}
		if len(got) != 0 {
			t.Errorf("expected 0 items, got %d", len(got))
		}
	})
}
