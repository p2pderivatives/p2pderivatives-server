package userservice

import (
	"unicode"
)

// VerifyNewPassword checks that the provided password satisfies the security
// requirements
func VerifyNewPassword(password string) bool {
	var number, upper, lower, special bool
	letters := 0
	for _, c := range password {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLower(c):
			lower = true
		case isSpecialCharacter(c):
			special = true
		case unicode.IsLetter(c):
		}
		letters++
	}
	validLength := (letters >= 8 && letters <= 32)

	//Only containing all cases is allowed
	return (number && upper && lower && special && validLength)
}

func isSpecialCharacter(c rune) bool {
	specialChar := " !\"#$%&'()*+,-./:;<=>?@[]^_`{|}~"
	for _, s := range specialChar {
		if s == c {
			return true
		}
	}
	return false
}
