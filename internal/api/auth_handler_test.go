package api

import "testing"

func TestValidateEmail(t *testing.T) {
	t.Run("Given email is empty/whitespace, should return false", func(t *testing.T) {
		email := "   "
		isValid := validateEmail(email)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given email doesn't contain @, should return false", func(t *testing.T) {
		email := "usernameATdomain.com"
		isValid := validateEmail(email)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given email contains more than one @, should return false", func(t *testing.T) {
		email := "user@name@domain.com"
		isValid := validateEmail(email)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given email doesn't contain ., should return false", func(t *testing.T) {
		email := "user@domainDOTcom"
		isValid := validateEmail(email)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given valid email, should return true", func(t *testing.T) {
		email := "user@domain.com"
		isValid := validateEmail(email)

		if !isValid {
			t.Error("Expected isValid = true")
		}
	})
}

func TestValidatePassword(t *testing.T) {
	t.Run("Given password is empty/whitespace, should return false", func(t *testing.T) {
		password := "   "
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password is less than 8 characters, should return false", func(t *testing.T) {
		password := "A1!defg"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password is less than 8 characters (where e.g. emoji 😀 = 1), should return false", func(t *testing.T) {
		password := "A1!def😀"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password doesn't contain a lowercase letter, should return false", func(t *testing.T) {
		password := "A1!BCDEFG"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password doesn't contain an upper letter, should return false", func(t *testing.T) {
		password := "a1!bcdefg"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password doesn't contain a number, should return false", func(t *testing.T) {
		password := "a!!bcdefG"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password doesn't contain a special character, should return false", func(t *testing.T) {
		password := "a12bcdefG"
		isValid := validatePassword(password)

		if isValid {
			t.Error("Expected isValid = false")
		}
	})

	t.Run("Given password meets criteria, should return true", func(t *testing.T) {
		password := "a1!bcdefG"
		isValid := validatePassword(password)

		if !isValid {
			t.Error("Expected isValid = true")
		}
	})

}
