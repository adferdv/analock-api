package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adfer-dev/analock-api/constants"
	"github.com/adfer-dev/analock-api/models"
	"github.com/adfer-dev/analock-api/services"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

// --- Mock Implementations ---
type mockTokenManager struct {
	ValidateTokenFunc func(token string) error
	GetClaimsFunc     func(token string) (jwt.MapClaims, error)
	GenerateTokenFunc func(user models.User, tokenKind models.TokenKind) (string, error)
}

func (m *mockTokenManager) ValidateToken(token string) error {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(token)
	}
	return nil
}

func (m *mockTokenManager) GetClaims(token string) (jwt.MapClaims, error) {
	if m.GetClaimsFunc != nil {
		return m.GetClaimsFunc(token)
	}
	return nil, nil
}

func (m *mockTokenManager) GenerateToken(user models.User, tokenKind models.TokenKind) (string, error) {
	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(user, tokenKind)
	}
	return "", nil
}

type mockTokenService struct {
	GetTokenByIdFunc       func(id uint) (*models.Token, error)
	GetTokenByValueFunc    func(token string) (*models.Token, error)
	GetUserTokenByKindFunc func(userId uint, kind models.TokenKind) (*models.Token, error)
	GetUserTokenPairFunc   func(userId uint) ([2]*models.Token, error)
	SaveTokenFunc          func(tokenBody *models.Token) (*models.Token, error)
	UpdateTokenFunc        func(tokenBody *models.Token) (*models.Token, error)
	DeleteTokenFunc        func(id uint) error
}

func (m *mockTokenService) GetTokenById(id uint) (*models.Token, error) {
	if m.GetTokenByIdFunc != nil {
		return m.GetTokenByIdFunc(id)
	}
	return nil, nil
}

func (m *mockTokenService) GetTokenByValue(token string) (*models.Token, error) {
	if m.GetTokenByValueFunc != nil {
		return m.GetTokenByValueFunc(token)
	}
	return nil, nil
}

func (m *mockTokenService) GetUserTokenByKind(userId uint, kind models.TokenKind) (*models.Token, error) {
	if m.GetUserTokenByKindFunc != nil {
		return m.GetUserTokenByKindFunc(userId, kind)
	}
	return nil, nil
}

func (m *mockTokenService) GetUserTokenPair(userId uint) ([2]*models.Token, error) {
	if m.GetUserTokenPairFunc != nil {
		return m.GetUserTokenPairFunc(userId)
	}
	return [2]*models.Token{}, nil
}

func (m *mockTokenService) SaveToken(tokenBody *models.Token) (*models.Token, error) {
	if m.SaveTokenFunc != nil {
		return m.SaveTokenFunc(tokenBody)
	}
	return nil, nil
}

func (m *mockTokenService) UpdateToken(tokenBody *models.Token) (*models.Token, error) {
	if m.UpdateTokenFunc != nil {
		return m.UpdateTokenFunc(tokenBody)
	}
	return nil, nil
}

func (m *mockTokenService) DeleteToken(id uint) error {
	if m.DeleteTokenFunc != nil {
		return m.DeleteTokenFunc(id)
	}
	return nil
}

type mockUserService struct {
	GetUserByIdFunc    func(id uint) (*models.User, error)
	GetUserByEmailFunc func(email string) (*models.User, error)
	DeleteUserFunc     func(id uint) error
	SaveUserFunc       func(userBody services.UserBody) (*models.User, error)
	UpdateUserFunc     func(userBody services.UserBody) (*models.User, error)
}

func (m *mockUserService) GetUserById(id uint) (*models.User, error) {
	if m.GetUserByIdFunc != nil {
		return m.GetUserByIdFunc(id)
	}
	return nil, nil
}

func (m *mockUserService) GetUserByEmail(email string) (*models.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, nil
}

func (m *mockUserService) DeleteUser(id uint) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(id)
	}
	return nil
}

func (m *mockUserService) SaveUser(userBody services.UserBody) (*models.User, error) {
	if m.SaveUserFunc != nil {
		return m.SaveUserFunc(userBody)
	}
	return nil, nil
}

func (m *mockUserService) UpdateUser(userBody services.UserBody) (*models.User, error) {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(userBody)
	}
	return nil, nil
}

type mockDiaryEntryService struct {
	GetDiaryEntryByIdFunc       func(id uint) (*models.DiaryEntry, error)
	GetUserEntriesFunc          func(userId uint) ([]*models.DiaryEntry, error)
	GetUserEntriesTimeRangeFunc func(userId uint, startDate int64, endDate int64) ([]*models.DiaryEntry, error)
	SaveDiaryEntryFunc          func(diaryEntryBody *services.SaveDiaryEntryBody, userId uint) (*models.DiaryEntry, error)
	UpdateDiaryEntryFunc        func(diaryEntryId uint, diaryEntryBody *services.UpdateDiaryEntryBody) (*models.DiaryEntry, error)
	DeleteDiaryEntryFunc        func(id uint) error
}

func (m *mockDiaryEntryService) GetDiaryEntryById(id uint) (*models.DiaryEntry, error) {
	if m.GetDiaryEntryByIdFunc != nil {
		return m.GetDiaryEntryByIdFunc(id)
	}
	return nil, nil
}

func (m *mockDiaryEntryService) GetUserEntries(userId uint) ([]*models.DiaryEntry, error) {
	if m.GetUserEntriesFunc != nil {
		return m.GetUserEntriesFunc(userId)
	}
	return nil, nil
}

func (m *mockDiaryEntryService) GetUserEntriesTimeRange(userId uint, startDate int64, endDate int64) ([]*models.DiaryEntry, error) {
	if m.GetUserEntriesTimeRangeFunc != nil {
		return m.GetUserEntriesTimeRangeFunc(userId, startDate, endDate)
	}
	return nil, nil
}

func (m *mockDiaryEntryService) SaveDiaryEntry(diaryEntryBody *services.SaveDiaryEntryBody, userId uint) (*models.DiaryEntry, error) {
	if m.SaveDiaryEntryFunc != nil {
		return m.SaveDiaryEntryFunc(diaryEntryBody, userId)
	}
	return nil, nil
}

func (m *mockDiaryEntryService) UpdateDiaryEntry(diaryEntryId uint, diaryEntryBody *services.UpdateDiaryEntryBody) (*models.DiaryEntry, error) {
	if m.UpdateDiaryEntryFunc != nil {
		return m.UpdateDiaryEntryFunc(diaryEntryId, diaryEntryBody)
	}
	return nil, nil
}

func (m *mockDiaryEntryService) DeleteDiaryEntry(id uint) error {
	if m.DeleteDiaryEntryFunc != nil {
		return m.DeleteDiaryEntryFunc(id)
	}
	return nil
}

// Test checkAuth function
func TestCheckAuth(t *testing.T) {
	// Temporarily replace global variables with mock implementations
	originalTokenManager := tokenManager
	originalTokenService := tokenService
	originalUserService := userService
	defer func() {
		tokenManager = originalTokenManager
		tokenService = originalTokenService
		userService = originalUserService
	}()

	tests := []testCaseCheckAuth{
		{
			name:        "No Authorization header",
			authHeader:  "",
			expectedErr: errors.New("authorization token must be provided, starting with Bearer"),
		},
		{
			name:        "Invalid Authorization header format",
			authHeader:  "InvalidToken",
			expectedErr: errors.New("authorization token must be provided, starting with Bearer"),
		},
		{
			name:                 "Expired token",
			authHeader:           "Bearer expired.token",
			mockValidateTokenErr: &jwt.ValidationError{Errors: jwt.ValidationErrorExpired},
			expectedErr:          errors.New("token expired. Please, get a new one at /auth/refresh-token"),
		},
		{
			name:                 "Invalid token (generic)",
			authHeader:           "Bearer invalid.token",
			mockValidateTokenErr: errors.New("some invalid token error"),
			expectedErr:          errors.New("token not valid"),
		},
		{
			name:                   "Token not found in database (revoked)",
			authHeader:             "Bearer valid.token",
			mockValidateTokenErr:   nil,
			mockGetTokenByValue:    nil,
			mockGetTokenByValueErr: errors.New("token not found"),
			expectedErr:            errors.New("token revoked"),
		},
		{
			name:                   "Valid token, admin user, non-user-accessible POST",
			authHeader:             "Bearer admin.token",
			mockValidateTokenErr:   nil,
			mockGetTokenByValue:    &models.Token{},
			mockGetTokenByValueErr: nil,
			mockGetClaims:          jwt.MapClaims{"sub": float64(1)},
			mockGetClaimsErr:       nil,
			mockGetUserByEmail:     &models.User{Role: models.Admin},
			mockGetUserByEmailErr:  nil,
			reqMethod:              http.MethodPost,
			reqURLPath:             "/api/v1/users",
			expectedErr:            nil,
		},
		{
			name:                   "Valid token, non-admin user, non-user-accessible POST",
			authHeader:             "Bearer user.token",
			mockValidateTokenErr:   nil,
			mockGetTokenByValue:    &models.Token{},
			mockGetTokenByValueErr: nil,
			mockGetClaims:          jwt.MapClaims{"sub": float64(1)},
			mockGetClaimsErr:       nil,
			mockGetUserByEmail:     &models.User{Role: models.Standard},
			mockGetUserByEmailErr:  nil,
			reqMethod:              http.MethodPost,
			reqURLPath:             "/api/v1/users",
			expectedErr:            nil,
		},
		{
			name:                   "Valid token, non-admin user, user-accessible POST (diaryEntries)",
			authHeader:             "Bearer user.token",
			mockValidateTokenErr:   nil,
			mockGetTokenByValue:    &models.Token{},
			mockGetTokenByValueErr: nil,
			mockGetClaims:          jwt.MapClaims{"sub": float64(1)},
			mockGetClaimsErr:       nil,
			mockGetUserByEmail:     &models.User{Role: models.Standard},
			mockGetUserByEmailErr:  nil,
			reqMethod:              http.MethodPost,
			reqURLPath:             "/api/v1/diaryEntries",
			expectedErr:            nil,
		},
		{
			name:                   "Valid token, non-admin user, user-accessible GET (activityRegistrations)",
			authHeader:             "Bearer user.token",
			mockValidateTokenErr:   nil,
			mockGetTokenByValue:    &models.Token{},
			mockGetTokenByValueErr: nil,
			mockGetClaims:          jwt.MapClaims{"sub": float64(1)},
			mockGetClaimsErr:       nil,
			mockGetUserByEmail:     &models.User{Role: models.Standard},
			mockGetUserByEmailErr:  nil,
			reqMethod:              http.MethodGet,
			reqURLPath:             "/api/v1/activityRegistrations/books/user/123",
			expectedErr:            nil,
		},
		{
			name:                   "Valid token, non-admin user, user-accessible PUT (diaryEntries)",
			authHeader:             "Bearer user.token",
			mockValidateTokenErr:   nil,
			mockGetTokenByValue:    &models.Token{},
			mockGetTokenByValueErr: nil,
			mockGetClaims:          jwt.MapClaims{"sub": float64(1)},
			mockGetClaimsErr:       nil,
			mockGetUserByEmail:     &models.User{Role: models.Standard},
			mockGetUserByEmailErr:  nil,
			reqMethod:              http.MethodPut,
			reqURLPath:             "/api/v1/diaryEntries/123",
			expectedErr:            nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			tokenManager = &mockTokenManager{
				ValidateTokenFunc: testCase.mockValidateTokenErrFunc(),
				GetClaimsFunc:     func(token string) (jwt.MapClaims, error) { return testCase.mockGetClaims, testCase.mockGetClaimsErr },
			}
			tokenService = &mockTokenService{
				GetTokenByValueFunc: func(token string) (*models.Token, error) {
					return testCase.mockGetTokenByValue, testCase.mockGetTokenByValueErr
				},
				GetUserTokenByKindFunc: func(userId uint, kind models.TokenKind) (*models.Token, error) { return &models.Token{}, nil },
				GetTokenByIdFunc:       func(id uint) (*models.Token, error) { return &models.Token{}, nil },
				GetUserTokenPairFunc:   func(userId uint) ([2]*models.Token, error) { return [2]*models.Token{}, nil },
				SaveTokenFunc:          func(tokenBody *models.Token) (*models.Token, error) { return &models.Token{}, nil },
				UpdateTokenFunc:        func(tokenBody *models.Token) (*models.Token, error) { return &models.Token{}, nil },
				DeleteTokenFunc:        func(id uint) error { return nil },
			}
			userService = &mockUserService{
				GetUserByEmailFunc: func(email string) (*models.User, error) {
					return testCase.mockGetUserByEmail, testCase.mockGetUserByEmailErr
				},
				GetUserByIdFunc: func(id uint) (*models.User, error) { return &models.User{}, nil },
				DeleteUserFunc:  func(id uint) error { return nil },
				SaveUserFunc:    func(userBody services.UserBody) (*models.User, error) { return &models.User{}, nil },
				UpdateUserFunc:  func(userBody services.UserBody) (*models.User, error) { return &models.User{}, nil },
			}

			req, _ := http.NewRequest(testCase.reqMethod, testCase.reqURLPath, nil)
			if testCase.authHeader != "" {
				req.Header.Set("Authorization", testCase.authHeader)
			}

			err := checkAuth(req)

			if (err == nil && testCase.expectedErr != nil) || (err != nil && testCase.expectedErr == nil) || (err != nil && err.Error() != testCase.expectedErr.Error()) {
				t.Errorf("checkAuth() error = %v, wantErr %v", err, testCase.expectedErr)
			}
		})
	}
}

// CheckAuth test case struct
type testCaseCheckAuth struct {
	name                   string
	authHeader             string
	mockValidateTokenErr   error
	mockGetTokenByValue    *models.Token
	mockGetTokenByValueErr error
	mockGetClaims          jwt.MapClaims
	mockGetClaimsErr       error
	mockGetUserByEmail     *models.User
	mockGetUserByEmailErr  error
	reqMethod              string
	reqURLPath             string
	expectedErr            error
}

// Helper function to return a specific *jwt.ValidationError or generic error
func (tt testCaseCheckAuth) mockValidateTokenErrFunc() func(string) error {
	if tt.mockValidateTokenErr == nil {
		return func(string) error { return nil }
	}
	if vErr, ok := tt.mockValidateTokenErr.(*jwt.ValidationError); ok {
		return func(string) error { return vErr }
	}
	return func(string) error { return tt.mockValidateTokenErr }
}

// CheckUserOwnershipMiddleware test case struct
type testCaseCheckUserOwnershipMiddleware struct {
	name                     string
	reqMethod                string
	reqURLPath               string
	reqID                    string
	authHeader               string
	mockGetClaims            jwt.MapClaims
	mockGetClaimsErr         error
	mockGetUserById          *models.User
	mockGetUserByIdErr       error
	mockGetDiaryEntryById    *models.DiaryEntry
	mockGetDiaryEntryByIdErr error
	expectedErr              error
}

// Test CheckUserOwnershipMiddleware
func TestCheckUserOwnershipMiddleware(t *testing.T) {
	// Temporarily replace global variables with mock implementations
	originalTokenManager := tokenManager
	originalUserService := userService
	originalDiaryEntryService := diaryEntryService
	defer func() {
		tokenManager = originalTokenManager
		userService = originalUserService
		diaryEntryService = originalDiaryEntryService
	}()

	// Init test cases
	tests := []testCaseCheckUserOwnershipMiddleware{
		{
			name:             "Invalid token claims",
			reqMethod:        http.MethodGet,
			reqURLPath:       "/api/v1/diaryEntries/123",
			reqID:            "123",
			authHeader:       "Bearer bad.token",
			mockGetClaimsErr: errors.New("invalid claims"),
			expectedErr:      errors.New("invalid claims"),
		},
		{
			name:                     "GET diary entry - user does not own",
			reqMethod:                http.MethodGet,
			reqURLPath:               "/api/v1/diaryEntries/123",
			reqID:                    "123",
			authHeader:               "Bearer valid.token",
			mockGetClaims:            jwt.MapClaims{"sub": float64(123)},
			mockGetDiaryEntryById:    &models.DiaryEntry{Registration: models.ActivityRegistration{UserRefer: 456}},
			mockGetDiaryEntryByIdErr: nil,
			mockGetUserById:          &models.User{Email: "user@example.com"},
			mockGetUserByIdErr:       nil,
			expectedErr:              errors.New(constants.ErrorUnauthorizedOperation),
		},
		{
			name:                     "GET diary entry - user owns",
			reqMethod:                http.MethodGet,
			reqURLPath:               "/api/v1/diaryEntries/123",
			reqID:                    "123",
			authHeader:               "Bearer valid.token",
			mockGetClaims:            jwt.MapClaims{"sub": float64(123)},
			mockGetDiaryEntryById:    &models.DiaryEntry{Registration: models.ActivityRegistration{UserRefer: 123}},
			mockGetDiaryEntryByIdErr: nil,
			mockGetUserById:          &models.User{Email: "user@example.com"},
			mockGetUserByIdErr:       nil,
			expectedErr:              nil,
		},
		{
			name:                     "PUT diary entry - user does not own",
			reqMethod:                http.MethodPut,
			reqURLPath:               "/api/v1/diaryEntries/123",
			reqID:                    "123",
			authHeader:               "Bearer valid.token",
			mockGetClaims:            jwt.MapClaims{"sub": float64(123)},
			mockGetDiaryEntryById:    &models.DiaryEntry{Registration: models.ActivityRegistration{UserRefer: 456}},
			mockGetDiaryEntryByIdErr: nil,
			mockGetUserById:          &models.User{Email: "user@example.com"},
			mockGetUserByIdErr:       nil,
			expectedErr:              errors.New(constants.ErrorUnauthorizedOperation),
		},
		{
			name:                     "PUT diary entry - user owns",
			reqMethod:                http.MethodPut,
			reqURLPath:               "/api/v1/diaryEntries/123",
			reqID:                    "123",
			authHeader:               "Bearer valid.token",
			mockGetClaims:            jwt.MapClaims{"sub": float64(123)},
			mockGetDiaryEntryById:    &models.DiaryEntry{Registration: models.ActivityRegistration{UserRefer: 123}},
			mockGetDiaryEntryByIdErr: nil,
			mockGetUserById:          &models.User{Email: "user@example.com"},
			mockGetUserByIdErr:       nil,
			expectedErr:              nil,
		},
		{
			name:                     "GET book registration - user does not own",
			reqMethod:                http.MethodGet,
			reqURLPath:               "/api/v1/activityRegistrations/books/123",
			reqID:                    "123",
			authHeader:               "Bearer valid.token",
			mockGetClaims:            jwt.MapClaims{"sub": float64(123)},
			mockGetDiaryEntryById:    &models.DiaryEntry{Registration: models.ActivityRegistration{UserRefer: 456}},
			mockGetDiaryEntryByIdErr: nil,
			mockGetUserById:          &models.User{Email: "user@example.com"},
			mockGetUserByIdErr:       nil,
			expectedErr:              errors.New(constants.ErrorUnauthorizedOperation),
		},
		{
			name:                     "GET game registration - user owns",
			reqMethod:                http.MethodGet,
			reqURLPath:               "/api/v1/activityRegistrations/games/123",
			reqID:                    "123",
			authHeader:               "Bearer valid.token",
			mockGetClaims:            jwt.MapClaims{"sub": float64(123)},
			mockGetDiaryEntryById:    &models.DiaryEntry{Registration: models.ActivityRegistration{UserRefer: 123}},
			mockGetDiaryEntryByIdErr: nil,
			mockGetUserById:          &models.User{Id: 123},
			mockGetUserByIdErr:       nil,
			expectedErr:              nil,
		},
		{
			name:          "Non-GET/PUT method (e.g., POST) - should pass through",
			reqMethod:     http.MethodPost,
			reqURLPath:    "/api/v1/diaryEntries/123",
			reqID:         "123",
			authHeader:    "Bearer valid.token",
			mockGetClaims: jwt.MapClaims{"sub": float64(1)},
			expectedErr:   nil,
		},
		{
			name:          "Non-protected endpoint - should pass through",
			reqMethod:     http.MethodGet,
			reqURLPath:    "/api/v1/users/123",
			reqID:         "123",
			authHeader:    "Bearer valid.token",
			mockGetClaims: jwt.MapClaims{"sub": float64(1)},
			expectedErr:   nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// Set up mock implementations
			tokenManager = &mockTokenManager{
				ValidateTokenFunc: func(token string) error { return nil },
				GetClaimsFunc:     func(token string) (jwt.MapClaims, error) { return testCase.mockGetClaims, testCase.mockGetClaimsErr },
			}
			userService = &mockUserService{
				GetUserByIdFunc:    func(id uint) (*models.User, error) { return testCase.mockGetUserById, testCase.mockGetUserByIdErr },
				GetUserByEmailFunc: func(email string) (*models.User, error) { return &models.User{Email: email}, nil },
			}
			diaryEntryService = &mockDiaryEntryService{
				GetDiaryEntryByIdFunc: func(id uint) (*models.DiaryEntry, error) {
					return testCase.mockGetDiaryEntryById, testCase.mockGetDiaryEntryByIdErr
				},
			}

			// Create request with path variables
			req := httptest.NewRequest(testCase.reqMethod, testCase.reqURLPath, nil)
			req = mux.SetURLVars(req, map[string]string{"id": testCase.reqID})
			if testCase.authHeader != "" {
				req.Header.Set("Authorization", testCase.authHeader)
			}

			// Call the middleware directly
			err := checkUserOwnershipMiddleware(req)

			// Check error
			if (err == nil && testCase.expectedErr != nil) || (err != nil && testCase.expectedErr == nil) || (err != nil && err.Error() != testCase.expectedErr.Error()) {
				t.Errorf("checkUserOwnershipMiddleware() error = %v, wantErr %v", err, testCase.expectedErr)
			}
		})
	}
}
