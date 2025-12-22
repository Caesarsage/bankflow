package account

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

// GenerateAccountNumber generates a unique account number
// Format: XX-XXXX-XXXX(10 digits)
func GenerateAccountNumber() (string, error) {
	// Generate 1- random digits
	digits := make([]string, 10)
	for i := 0; i < 10; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digits[i] = n.String()
	}

	// Format as XXXX-XXXX-XXXX-XXXX
	accountNumber := fmt.Sprintf("%s-%s-%s",
		strings.Join(digits[0:2], ""),
		strings.Join(digits[2:6], ""),
		strings.Join(digits[6:10], ""))

	return accountNumber, nil
}

// ValidateAccountNumber validates an account number format
func ValidateAccountNumber(accountNumber string) bool {
	// Remove hyphens
	clean := strings.ReplaceAll(accountNumber, "-", "")

	// Must be 16 digits
	if len(clean) != 10 {
		return false
	}

	// Must be all digits
	for _, char := range clean {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
