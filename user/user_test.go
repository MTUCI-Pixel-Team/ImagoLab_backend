package user

import (
	"RestAPI/core"
	"errors"
	"testing"
	"time"
)

// /*
// Test mobile.go
// */
// func TestSendSms(t *testing.T) {
// 	testCases := []struct {
// 		message string
// 		Mobile  string
// 		err     error
// 	}{
// 		{
// 			message: "Hello",
// 			Mobile:  "79936977511",
// 			err:     nil,
// 		},
// 		{
// 			message: "",
// 			Mobile:  "79936977511",
// 			err:     errors.New("Message is empty"),
// 		},
// 		{
// 			message: "Hello",
// 			Mobile:  "",
// 			err:     errors.New("Mobile number is empty"),
// 		},
// 	}

// 	err := godotenv.Load("../.env")
// 	if err != nil {
// 		t.Errorf("Error env load %v", err)
// 	}
// 	core.MTS_API_KEY = os.Getenv("MTS_API_KEY")
// 	core.MTS_API_NUMBER = os.Getenv("MTS_API_NUMBER")
// 	if core.MTS_API_KEY == "" || core.MTS_API_NUMBER == "" {
// 		t.Errorf("Error env load %v", err)
// 	}

// 	for _, tc := range testCases {
// 		u := new(User)
// 		u.Mobile = tc.Mobile
// 		err := u.SendSMS(tc.message)
// 		if err != nil {
// 			if tc.err == nil {
// 				t.Errorf("Unexpected error: %v", err)
// 			} else if err.Error() != tc.err.Error() {
// 				t.Errorf("Expected error: %v, got: %v", tc.err, err)
// 			}
// 		} else if tc.err != nil {
// 			t.Errorf("Expected error: %v, but got none", tc.err)
// 		}
// 	}
// }

/*
Test jwtOuth.go
*/
func initSecrets() {
	core.JWT_ACCESS_SECRET_KEY = "testAccessSecretKey"
	core.JWT_REFRESH_SECRET_KEY = "testRefreshSecretKey"
	core.JWT_ACCESS_EXPIRATION_TIME = time.Minute * 5
	core.JWT_REFRESH_EXPIRATION_TIME = time.Minute * 10
}

func TestJWTFunctions(t *testing.T) {
	testCases := []struct {
		name        string
		username    string
		email       string
		expectedErr error
		testFunc    func() error
	}{
		{
			name:        "Generate Secret Key",
			email:       "",
			username:    "",
			expectedErr: nil,
			testFunc: func() error {
				_, err := GenerateSecretKey(32)
				return err
			},
		},
		{
			name:        "Generate Access Token",
			email:       "test@example.com",
			username:    "testuser",
			expectedErr: nil,
			testFunc: func() error {
				_, err := GenerateAccessToken("testuser", "test@example.com")
				return err
			},
		},
		{
			name:        "Generate Refresh Token",
			email:       "test@example.com",
			username:    "testuser",
			expectedErr: nil,
			testFunc: func() error {
				_, err := GenerateRefreshToken("testuser", "test@example.com")
				return err
			},
		},
		{
			name:        "Validate Access Token",
			username:    "testuser",
			email:       "test@example.com",
			expectedErr: nil,
			testFunc: func() error {
				token, err := GenerateAccessToken("testuser", "test@example.com")
				if err != nil {
					return err
				}
				_, err = ValidateToken(token)
				return err
			},
		},
		{
			name:        "Invalid Access Token",
			username:    "",
			email:       "",
			expectedErr: errors.New("error expected"),
			testFunc: func() error {
				_, err := ValidateToken("invalidToken")
				return err
			},
		},
		{
			name:        "Refresh Tokens",
			username:    "testuser",
			email:       "test@example.com",
			expectedErr: nil,
			testFunc: func() error {
				refreshToken, err := GenerateRefreshToken("testuser", "test@example.com")
				if err != nil {
					return err
				}
				_, _, err = RefreshTokens(refreshToken)
				return err
			},
		},
		{
			name:        "Expired Access Token",
			username:    "testuser",
			email:       "test@example.com",
			expectedErr: errors.New("error expected"),
			testFunc: func() error {
				core.JWT_ACCESS_EXPIRATION_TIME = time.Millisecond * 100
				token, err := GenerateAccessToken("testuser", "test@example.com")
				if err != nil {
					return err
				}
				time.Sleep(time.Millisecond * 200)
				_, err = ValidateToken(token)
				core.JWT_ACCESS_EXPIRATION_TIME = time.Minute * 5
				return err
			},
		},
		{
			name:        "Generate Access Token with Empty Username",
			username:    "",
			email:       "test@example.com",
			expectedErr: errors.New("username is empty"),
			testFunc: func() error {
				_, err := GenerateAccessToken("", "test@example.com")
				return err
			},
		},
		{
			name:        "Generate Refresh Token with Empty Email",
			username:    "testuser",
			email:       "",
			expectedErr: errors.New("email is empty"),
			testFunc: func() error {
				_, err := GenerateRefreshToken("testuser", "")
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc()
			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSendActivationEmail(t *testing.T) {
	testCases := []struct {
		name           string
		toEmail        string
		activationCode int
		expectedErr    error
	}{
		{
			name:           "Valid Email",
			toEmail:        "albertmonshtain@gmail.com",
			activationCode: generateActivationCode(),
			expectedErr:    nil,
		},
		{
			name:           "Empty Email",
			toEmail:        "",
			activationCode: generateActivationCode(),
			expectedErr:    errors.New("Email is empty"),
		},
		{
			name:           "Invalid Activation Link",
			toEmail:        "albertmonshtain@gmail.com",
			activationCode: 0,
			expectedErr:    errors.New("Invalid activation code"),
		},
	}

	err := core.InitEnv("../.env")
	if err != nil {
		t.Errorf("Error env load %v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SendActivationEmail(tc.toEmail, tc.activationCode)
			if tc.expectedErr != nil {
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCheckDomain(t *testing.T) {
	testCases := []struct {
		name        string
		email       string
		expectedRes bool
	}{
		{
			name:        "Valid Email",
			email:       "test@gmail.com",
			expectedRes: true,
		},
		{
			name:        "Invalid Email",
			email:       "meger52934@aqqor.com",
			expectedRes: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := CheckDomain(tc.email)
			if res != tc.expectedRes {
				t.Errorf("Expected %v, got %v", tc.expectedRes, res)
			}
		})
	}

}

// func TestSendSMS(t *testing.T) {
// 	testCases := []struct {
// 		name           string
// 		phone          string
// 		activationCode int
// 		expectedErr    error
// 	}{
// 		{
// 			name:           "Valid Phone Number",
// 			phone:          "79266131479",
// 			activationCode: generateActivationCode(),
// 			expectedErr:    nil,
// 		},
// 		{
// 			name:           "Empty Phone Number",
// 			phone:          "",
// 			activationCode: generateActivationCode(),
// 			expectedErr:    errors.New("Mobile number is empty"),
// 		},
// 		{
// 			name:           "Invalid Activation Link",
// 			phone:          "79936977511",
// 			activationCode: 0,
// 			expectedErr:    errors.New("Invalid code: 0"),
// 		},
// 	}

// 	err := core.InitEnv("../.env")
// 	if err != nil {
// 		t.Errorf("Error env load %v", err)
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			err := SendSMS(tc.phone, tc.activationCode)
// 			if tc.expectedErr != nil {
// 				if err == nil || err.Error() != tc.expectedErr.Error() {
// 					t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
// 				}
// 			} else {
// 				if err != nil {
// 					t.Errorf("Unexpected error: %v", err)
// 				}
// 			}
// 		})
// 	}
// }
