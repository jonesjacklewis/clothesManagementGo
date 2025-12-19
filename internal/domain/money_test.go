package domain

import "testing"

func TestAsPounds(t *testing.T) {
	t.Run("Given 0 pence, then should return £0.00", func(t *testing.T) {
		pence := Pence(0)
		expected := "£0.00"
		got := pence.AsPounds()

		if got != expected {
			t.Errorf("Expected %s got %s\n", expected, got)
		}
	})

	t.Run("Given a positive pence balance less than 100,000 (1255), should format correctly (£12.55)", func(t *testing.T) {
		pence := Pence(1255)
		expected := "£12.55"
		got := pence.AsPounds()

		if got != expected {
			t.Errorf("Expected %s got %s\n", expected, got)
		}

	})

	t.Run("Given a positive pence balance greater than 100,000 (120055), should format correctly (£1,200.55)", func(t *testing.T) {
		pence := Pence(120055)
		expected := "£1,200.55"
		got := pence.AsPounds()

		if got != expected {
			t.Errorf("Expected %s got %s\n", expected, got)
		}

	})

	t.Run("Given a negative pence balance with magnitude less than 100,000 (e.g. -1255), should format correctly (-£12.55)", func(t *testing.T) {
		pence := Pence(-1255)
		expected := "-£12.55"
		got := pence.AsPounds()

		if got != expected {
			t.Errorf("Expected %s got %s\n", expected, got)
		}

	})

	t.Run("Given a negative pence balance with magnitude  greater than 100,000 (-120055), should format correctly (-£1,200.55)", func(t *testing.T) {
		pence := Pence(-120055)
		expected := "-£1,200.55"
		got := pence.AsPounds()

		if got != expected {
			t.Errorf("Expected %s got %s\n", expected, got)
		}

	})
}
