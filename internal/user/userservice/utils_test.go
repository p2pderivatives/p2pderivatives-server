package userservice

import (
	"testing"
)

func TestUser_VerifyNewPassword(t *testing.T) {
	testCases := []struct {
		name         string
		testPassword string
		isOk         bool
	}{
		{name: "OK        :", testPassword: "P@ssw0rdAlice", isOk: true},
		{name: "No Number :", testPassword: "P@sswordAlice", isOk: false},
		{name: "No Upper  :", testPassword: "p@ssw0rdalice", isOk: false},
		{name: "No Lower  :", testPassword: "P@SSW0RDALICE", isOk: false},
		{name: "No Special:", testPassword: "Passw0rdAlice", isOk: false},
		{name: "Too short :", testPassword: "P@ssw0r", isOk: false},                           // 7
		{name: "Too long  :", testPassword: "P@ssw0rd9012345678901234567890123", isOk: false}, // 33
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := VerifyNewPassword(tc.testPassword)
			if tc.isOk != actual {
				t.Errorf("Fail %s isOk=%v", tc.testPassword, tc.isOk)
			}
		})
	}
}
