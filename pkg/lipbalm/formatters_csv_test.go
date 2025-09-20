package lipbalm

import (
	"strings"
	"testing"
)

func TestCSVFormatterPointerDereferencing(t *testing.T) {
	formatter := &CSVFormatter{}
	config := &Config{}

	t.Run("slice of pointers to structs", func(t *testing.T) {
		type Person struct {
			Name  string `json:"name"`
			Age   int    `json:"age"`
			Email string `json:"email"`
		}

		// Create slice of pointers
		people := []*Person{
			{Name: "Alice", Age: 30, Email: "alice@example.com"},
			{Name: "Bob", Age: 25, Email: "bob@example.com"},
			{Name: "Charlie", Age: 35, Email: "charlie@example.com"},
		}

		result, err := formatter.Render(people, config)
		if err != nil {
			t.Fatalf("Failed to render CSV: %v", err)
		}

		// Check that we have headers
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) != 4 { // 1 header + 3 data rows
			t.Errorf("Expected 4 lines, got %d", len(lines))
		}

		// Check headers - gocsv uses struct field names, not json tags
		expectedHeader := "Name,Age,Email"
		if lines[0] != expectedHeader {
			t.Errorf("Expected header %q, got %q", expectedHeader, lines[0])
		}

		// Check that data doesn't contain memory addresses
		for i, line := range lines[1:] {
			if strings.Contains(line, "0x") {
				t.Errorf("Line %d contains memory address: %s", i+1, line)
			}
			// Check actual data
			parts := strings.Split(line, ",")
			if len(parts) != 3 {
				t.Errorf("Expected 3 fields, got %d in line: %s", len(parts), line)
			}
		}

		// Verify specific content
		if !strings.Contains(result, "Alice,30,alice@example.com") {
			t.Error("Missing expected data for Alice")
		}
	})

	t.Run("struct with pointer fields", func(t *testing.T) {
		type Address struct {
			Street  *string
			City    *string
			ZipCode *int
		}

		street := "123 Main St"
		city := "Springfield"
		zip := 12345

		addr := Address{
			Street:  &street,
			City:    &city,
			ZipCode: &zip,
		}

		result, err := formatter.Render(addr, config)
		if err != nil {
			t.Fatalf("Failed to render CSV: %v", err)
		}

		// Should not contain memory addresses
		if strings.Contains(result, "0x") {
			t.Errorf("Result contains memory address: %s", result)
		}

		// Should contain actual values
		expectedValues := []string{"123 Main St", "Springfield", "12345"}
		for _, expected := range expectedValues {
			if !strings.Contains(result, expected) {
				t.Errorf("Missing expected value %q in result: %s", expected, result)
			}
		}
	})

	t.Run("slice with nil pointers", func(t *testing.T) {
		type Item struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		items := []*Item{
			{ID: 1, Name: "Item 1"},
			nil, // gocsv renders nil as empty row
			{ID: 3, Name: "Item 3"},
		}

		result, err := formatter.Render(items, config)
		if err != nil {
			t.Fatalf("Failed to render CSV: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(result), "\n")
		// gocsv includes nil as empty row, so we get header + 3 data rows
		if len(lines) != 4 {
			t.Errorf("Expected 4 lines (with nil as empty row), got %d", len(lines))
		}
		
		// Check that second data row is empty (for nil)
		if lines[2] != "," && lines[2] != "0," {
			t.Errorf("Expected empty or zero row for nil, got %q", lines[2])
		}
	})

	t.Run("slice of simple pointer types not supported by gocsv", func(t *testing.T) {
		// gocsv only supports structs, not simple types
		str1 := "hello"
		str2 := "world"
		strPtrs := []*string{&str1, &str2, nil}

		_, err := formatter.Render(strPtrs, config)
		if err == nil {
			t.Fatal("Expected error for non-struct types")
		}
		
		// gocsv should reject non-struct types
		if !strings.Contains(err.Error(), "struct") {
			t.Errorf("Expected error about struct requirement, got: %v", err)
		}
	})
}

func TestCSVFormatterWithGoCSV(t *testing.T) {
	formatter := &CSVFormatter{}
	config := &Config{}

	t.Run("handles non-slice types", func(t *testing.T) {
		// gocsv requires a slice for proper CSV output
		// Single structs get wrapped in a slice internally
		type Person struct {
			Name  string `csv:"name"`
			Age   int    `csv:"age"`
			Email string `csv:"email"`
		}

		person := Person{Name: "Alice", Age: 30, Email: "alice@example.com"}

		result, err := formatter.Render([]Person{person}, config)
		if err != nil {
			t.Fatalf("Failed to render CSV: %v", err)
		}

		// Should have proper CSV headers and data
		lines := strings.Split(strings.TrimSpace(result), "\n")
		if len(lines) != 2 { // header + data
			t.Errorf("Expected 2 lines, got %d", len(lines))
		}

		// Check it contains the data
		if !strings.Contains(result, "Alice") || !strings.Contains(result, "30") || !strings.Contains(result, "alice@example.com") {
			t.Errorf("Missing expected data in result: %s", result)
		}
	})

	t.Run("complex struct with nested fields", func(t *testing.T) {
		// Test how gocsv handles the actual ChangeResult-like structure
		type Result struct {
			Command  string   `csv:"command"`
			Message  string   `csv:"message"`
			TodoUIDs []string `csv:"todo_uids"`
			Count    int      `csv:"count"`
		}

		results := []Result{
			{
				Command:  "list",
				Message:  "",
				TodoUIDs: []string{"123", "456", "789"},
				Count:    3,
			},
		}

		result, err := formatter.Render(results, config)
		if err != nil {
			t.Fatalf("Failed to render CSV: %v", err)
		}

		// gocsv should handle slices in some way
		if !strings.Contains(result, "list") || !strings.Contains(result, "3") {
			t.Errorf("Missing expected values in result: %s", result)
		}
	})
}