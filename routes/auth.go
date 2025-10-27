package routes

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"

	"github.com/adii2ma/dbms-backend/database"
	"github.com/adii2ma/dbms-backend/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// SignUpRequest represents the signup request body
type SignUpRequest struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=6"`
	Block    *string `json:"block"`
	RoomName *string `json:"room_name"`
	Phone    *string `json:"phone"`
}

// SignInRequest represents the signin request body
type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SignUpResponse represents the signup response
type SignUpResponse struct {
	Message string       `json:"message"`
	User    *models.User `json:"user"`
}

// SignInResponse represents the signin response
type SignInResponse struct {
	Message string       `json:"message"`
	User    *models.User `json:"user"`
	Success bool         `json:"success"`
}

// SignUp handles user registration
func SignUp(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[SignUp] failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	log.Printf("[SignUp] raw request body: %s", string(bodyBytes))
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req SignUpRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SignUp] validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Check if user already exists
	ctx := context.Background()
	exists, err := database.DB.NewSelect().
		Model((*models.User)(nil)).
		Where("email = ?", req.Email).
		Exists(ctx)

	if err != nil {
		log.Printf("[SignUp] database exists check failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}

	if exists {
		log.Printf("[SignUp] duplicate email attempted: %s", req.Email)
		c.JSON(http.StatusConflict, gin.H{
			"error": "User with this email already exists",
		})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[SignUp] password hash failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process password",
		})
		return
	}

	// Create new user
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Block:    req.Block,
		RoomName: req.RoomName,
		Phone:    req.Phone,
	}

	// Insert user into database
	_, err = database.DB.NewInsert().
		Model(user).
		Exec(ctx)

	if err != nil {
		log.Printf("[SignUp] insert failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[SignUp] user created successfully: %s", req.Email)
	// Return success response
	c.JSON(http.StatusCreated, SignUpResponse{
		Message: "User registered successfully",
		User:    user,
	})
}

// SignIn handles user login
func SignIn(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[SignIn] failed to read request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	log.Printf("[SignIn] raw request body: %s", string(bodyBytes))
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req SignInRequest

	// Validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SignIn] validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Find user by email
	ctx := context.Background()
	user := new(models.User)
	err = database.DB.NewSelect().
		Model(user).
		Where("email = ?", req.Email).
		Scan(ctx)

	if err != nil {
		log.Printf("[SignIn] user lookup failed for %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid email or password",
			"success": false,
		})
		return
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Printf("[SignIn] password mismatch for %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid email or password",
			"success": false,
		})
		return
	}

	log.Printf("[SignIn] login successful for %s", req.Email)
	// Return success response
	c.JSON(http.StatusOK, SignInResponse{
		Message: "Login successful",
		User:    user,
		Success: true,
	})
}
