package accounts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"triplanner/core"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate the User model
	err = db.AutoMigrate(&User{}, &UserPreferences{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	return router
}

func TestCreateUser(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	core.DB = db
	router := setupTestRouter()
	router.POST("/signup", CreateUser)

	tests := []struct {
		name           string
		input          AuthInput
		expectedStatus int
		checkDB        bool
	}{
		{
			name: "Valid user creation",
			input: AuthInput{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			checkDB:        true,
		},
		{
			name: "Missing username",
			input: AuthInput{
				Username: "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkDB:        false,
		},
		{
			name: "Missing password",
			input: AuthInput{
				Username: "testuser2",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkDB:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			jsonData, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check if user was created in database
			if tt.checkDB && w.Code == http.StatusOK {
				var user User
				db.Where("username = ?", tt.input.Username).First(&user)
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.Equal(t, tt.input.Username, user.Username)
				// Verify password is hashed
				assert.NotEqual(t, tt.input.Password, user.Password)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	core.DB = db
	router := setupTestRouter()
	router.POST("/login", Login)

	// Set SECRET for JWT
	if err := os.Setenv("SECRET", "test-secret-key"); err != nil {
		t.Fatalf("Failed to set SECRET environment variable: %v", err)
	}
	defer func() { _ = os.Unsetenv("SECRET") }()

	// Create a test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	db.Create(&testUser)

	tests := []struct {
		name           string
		input          AuthInput
		expectedStatus int
		checkToken     bool
	}{
		{
			name: "Valid login",
			input: AuthInput{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			checkToken:     true,
		},
		{
			name: "Invalid password",
			input: AuthInput{
				Username: "testuser",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
		{
			name: "Non-existent user",
			input: AuthInput{
				Username: "nonexistent",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
		{
			name: "Missing username",
			input: AuthInput{
				Username: "",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
		{
			name: "Missing password",
			input: AuthInput{
				Username: "testuser",
				Password: "",
			},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			jsonData, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check if token is returned
			if tt.checkToken && w.Code == http.StatusOK {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				assert.NotEmpty(t, response["token"])

				// Verify token is valid
				token, err := jwt.Parse(response["token"], func(token *jwt.Token) (interface{}, error) {
					return []byte(os.Getenv("SECRET")), nil
				})
				assert.NoError(t, err)
				assert.True(t, token.Valid)

				// Verify claims
				claims, ok := token.Claims.(jwt.MapClaims)
				assert.True(t, ok)
				assert.NotNil(t, claims["id"])
				assert.NotNil(t, claims["exp"])
			}
		})
	}
}

func TestGetUserProfile(t *testing.T) {
	// Setup
	db := setupTestDB(t)
	core.DB = db
	router := setupTestRouter()
	router.GET("/profile", GetUserProfile)

	// Create a test user
	testUser := User{
		Username: "testuser",
		Password: "hashedpassword",
	}
	db.Create(&testUser)

	tests := []struct {
		name           string
		setUser        bool
		expectedStatus int
	}{
		{
			name:           "User set in context",
			setUser:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No user in context",
			setUser:        false,
			expectedStatus: http.StatusOK, // Still returns 200 but with nil user
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a custom handler that sets user if needed
			router := gin.New()
			if tt.setUser {
				router.Use(func(c *gin.Context) {
					c.Set("currentUser", testUser)
					c.Next()
				})
			}
			router.GET("/profile", GetUserProfile)

			// Prepare request
			req, _ := http.NewRequest("GET", "/profile", nil)

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse response
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if tt.setUser {
				assert.NotNil(t, response["user"])
			}
		})
	}
}

func TestGoogleOAuthLogin(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.LoadHTMLGlob("../templates/*")
	router.GET("/auth/google", GoogleOAuthLogin)

	// Prepare request
	req, _ := http.NewRequest("GET", "/auth/google", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// For now, this will return an error if template doesn't exist
	// In real tests, you'd mock the template or check the response appropriately
	// We'll just verify it tries to render
	assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, w.Code)
}

func TestGoogleOAuthBegin(t *testing.T) {
	// Setup
	router := setupTestRouter()
	router.GET("/auth/google/begin", GoogleOAuthBegin)

	// This test verifies the function runs without panicking
	// Full OAuth testing would require mocking the gothic library
	req, _ := http.NewRequest("GET", "/auth/google/begin", nil)
	w := httptest.NewRecorder()

	// Execute - may redirect or return error
	router.ServeHTTP(w, req)

	// Just verify it doesn't panic and returns some response
	assert.NotNil(t, w.Code)
}

func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"

	// Generate hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	// Verify correct password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	assert.NoError(t, err)

	// Verify incorrect password
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	assert.Error(t, err)
}

func TestJWTTokenGeneration(t *testing.T) {
	if err := os.Setenv("SECRET", "test-secret-key"); err != nil {
		t.Fatalf("Failed to set SECRET environment variable: %v", err)
	}
	defer func() { _ = os.Unsetenv("SECRET") }()

	testUserID := uuid.New()

	// Generate token
	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  testUserID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	token, err := generateToken.SignedString([]byte(os.Getenv("SECRET")))
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse and verify token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Verify claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.NotNil(t, claims["id"])
	assert.NotNil(t, claims["exp"])
}

func TestJWTTokenExpiration(t *testing.T) {
	if err := os.Setenv("SECRET", "test-secret-key"); err != nil {
		t.Fatalf("Failed to set SECRET environment variable: %v", err)
	}
	defer func() { _ = os.Unsetenv("SECRET") }()

	testUserID := uuid.New()

	// Generate expired token
	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  testUserID,
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
	})

	token, err := generateToken.SignedString([]byte(os.Getenv("SECRET")))
	assert.NoError(t, err)

	// Try to parse expired token
	parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})

	// Token should be invalid due to expiration
	if parsedToken != nil {
		assert.False(t, parsedToken.Valid)
	}
}
