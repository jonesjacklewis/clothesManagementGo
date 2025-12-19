package domain

import "testing"

func TestIsValid(t *testing.T) {
	// Happy Path
	t.Run("Given Clothing has positive Price, and not empty ClothingType, Description, Brand, Store, and Size - Then should return nil", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  "Red Loosefit Jumper",
			Brand:        "A&B",
			Store:        "Totally Real Store",
			Size:         "Medium",
		}
		got := item.Validate()

		if got != nil {
			t.Errorf("Expected no error, but got %v", got)
		}
	})

	// Negative Price

	t.Run("Given Clothing has negative Price, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(-2000),
			ClothingType: "Jumper",
			Description:  "Red Loosefit Jumper",
			Brand:        "A&B",
			Store:        "Totally Real Store",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Price must be greater than or equal to 0"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	// Empty/Whitespace ClothingType

	t.Run("Given Clothing has empty ClothingType, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "",
			Description:  "Red Loosefit Jumper",
			Brand:        "A&B",
			Store:        "Totally Real Store",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Type must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	t.Run("Given Clothing has whitespace ClothingType, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "   ",
			Description:  "Red Loosefit Jumper",
			Brand:        "A&B",
			Store:        "Totally Real Store",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Type must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	// Empty/Whitespace Description

	t.Run("Given Clothing has empty Description, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  "",
			Brand:        "A&B",
			Store:        "Totally Real Store",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Description must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	t.Run("Given Clothing has whitespace Description, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  " ",
			Brand:        "A&B",
			Store:        "Totally Real Store",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Description must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	// Empty/Whitespace Brand

	t.Run("Given Clothing has empty Brand, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  "Red Loosefit Jumper",
			Brand:        "",
			Store:        "Totally Real Store",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Brand must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	t.Run("Given Clothing has whitespace Brand, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  "Red Loosefit Jumper",
			Brand:        " ",
			Store:        "Totally Real Store",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Brand must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	// Empty/Whitespace Store

	t.Run("Given Clothing has empty Store, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  "Red Loosefit Jumper",
			Brand:        "A&B",
			Store:        "",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Store must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	t.Run("Given Clothing has whitespace Store, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  "Red Loosefit Jumper",
			Brand:        "A&B",
			Store:        " ",
			Size:         "Medium",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Store must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	// Empty/Whitespace Size

	t.Run("Given Clothing has empty Size, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  "Red Loosefit Jumper",
			Brand:        "A&B",
			Store:        "Totally Real Store",
			Size:         "",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Size must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})

	t.Run("Given Clothing has whitespace Size, should return an appropriate error", func(t *testing.T) {
		var item Clothing = Clothing{
			Price:        Pence(2000),
			ClothingType: "Jumper",
			Description:  "Red Loosefit Jumper",
			Brand:        "A&B",
			Store:        "Totally Real Store",
			Size:         " ",
		}

		got := item.Validate()

		if got == nil {
			t.Error("Expected error, but got nil")
		}

		text := got.Error()
		expectedText := "Clothing Size must not be empty"

		if expectedText != text {
			t.Errorf("Expected '%s' but got '%s'", expectedText, text)
		}
	})
}
