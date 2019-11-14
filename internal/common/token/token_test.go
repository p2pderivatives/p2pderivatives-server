package token

import (
	"testing"
	"time"

	"github.com/bouk/monkey"
	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	wantedNow := time.Date(2019, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name             string
		userID           string
		expectedTokenStr string
		expectedExp      int64
		expectedErr      string
	}{
		{
			name:             "Success Case",
			userID:           "1",
			expectedTokenStr: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDYzNDU4MDAsImp0aSI6IjEifQ.QfDVtZbYnXaaQ3_vBJow-s9KT5OKAIT7O3dc9hR_yoc",
			expectedExp:      1800,
			expectedErr:      "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			patch := monkey.Patch(time.Now, func() time.Time { return wantedNow })
			defer patch.Unpatch()

			Init(&Config{
				Secret:     "k^Cc#*mdnS9$nTOY6S1#1i7^e*o1ijSl",
				Exp:        time.Minute * 30,
				RefreshExp: time.Hour * 24 * 30,
			})
			tokenStr, exp, err := GenerateAccessToken(test.userID)
			if err != nil || test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				return
			}
			assert.Equal(t, test.expectedTokenStr, tokenStr)
			assert.Equal(t, test.expectedExp, exp)
		})
	}
}

func TestVerifyToken(t *testing.T) {
	tests := []struct {
		name        string
		wantedNow   time.Time
		tokenStr    string
		userID      string
		expectedErr string
	}{
		{
			name:        "Success Case",
			wantedNow:   time.Date(2019, 1, 1, 12, 0, 0, 0, time.UTC),
			tokenStr:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDYzNDU4MDAsImp0aSI6IjEifQ.QfDVtZbYnXaaQ3_vBJow-s9KT5OKAIT7O3dc9hR_yoc",
			userID:      "1",
			expectedErr: "",
		},
		{
			name:        "Failed Case 1 token invalid",
			wantedNow:   time.Date(2019, 1, 1, 12, 0, 0, 0, time.UTC),
			tokenStr:    "12345",
			userID:      "",
			expectedErr: "12345 is invalid",
		},
		{
			name:        "Failed Case 2 token expired",
			wantedNow:   time.Date(2019, 1, 1, 13, 0, 0, 0, time.UTC),
			tokenStr:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDYzNDU4MDAsImp0aSI6InRlc3QxIn0._HUTrOKAtYzLLUrMzhpA7TOrkl1NEp_M5YoRDDZsDmg",
			userID:      "",
			expectedErr: ErrTokenExpired.Error(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			patch := monkey.Patch(time.Now, func() time.Time { return test.wantedNow })
			defer patch.Unpatch()

			Init(&Config{
				Secret:     "k^Cc#*mdnS9$nTOY6S1#1i7^e*o1ijSl",
				Exp:        time.Minute * 30,
				RefreshExp: time.Hour * 24 * 30,
			})
			result, err := VerifyToken(test.tokenStr)
			if err != nil || test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				return
			}
			assert.Equal(t, test.userID, result)
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	wantedNow := time.Date(2019, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name             string
		input            string
		expectedTokenStr string
		expectedErr      string
	}{
		{
			name:             "Success Case",
			input:            "uuid",
			expectedTokenStr: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDg5MzYwMDAsImp0aSI6InV1aWQifQ.0muLv3oOrCU1Rj8IJvsYqcWd0bE-UWVnC9y8afxRJ0Q",
			expectedErr:      "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			patch := monkey.Patch(time.Now, func() time.Time { return wantedNow })
			defer patch.Unpatch()

			Init(&Config{
				Secret:     "k^Cc#*mdnS9$nTOY6S1#1i7^e*o1ijSl",
				Exp:        time.Minute * 30,
				RefreshExp: time.Hour * 24 * 30,
			})
			tokenStr, err := GenerateRefreshToken(test.input)
			if err != nil || test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
				return
			}
			assert.Equal(t, test.expectedTokenStr, tokenStr)
		})
	}
}
