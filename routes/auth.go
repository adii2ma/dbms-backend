package routes

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
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

	// Start a transaction so we create user, room and room_member atomically
	tx, err := database.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("[SignUp] failed to begin tx: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer func() {
		// ensure rollback if not committed
		_ = tx.Rollback()
	}()

	// Insert user and get generated ID
	_, err = tx.NewInsert().Model(user).Returning("id").Exec(ctx)
	if err != nil {
		log.Printf("[SignUp] insert user failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	// If room name provided, create/find room and add room_member
	if req.RoomName != nil && *req.RoomName != "" {
		if req.Block == nil || *req.Block == "" {
			log.Printf("[SignUp] room_name provided without block")
			c.JSON(http.StatusBadRequest, gin.H{"error": "block is required when room_name is provided"})
			return
		}

		room := new(models.Room)
		err = tx.NewSelect().
			Model(room).
			Where("room_number = ?", *req.RoomName).
			Where("block = ?", *req.Block).
			Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// create room
				room = &models.Room{
					Block:      *req.Block,
					RoomNumber: *req.RoomName,
				}
				_, err = tx.NewInsert().Model(room).Returning("id").Exec(ctx)
				if err != nil {
					log.Printf("[SignUp] create room failed: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
					return
				}
			} else {
				log.Printf("[SignUp] find room failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
		}

		// create room_member linking user and room
		rm := &models.RoomMember{
			RoomID: room.ID,
			Block:  room.Block,
			UserID: user.ID,
		}
		_, err = tx.NewInsert().Model(rm).Exec(ctx)
		if err != nil {
			log.Printf("[SignUp] create room_member failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add room member"})
			return
		}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		log.Printf("[SignUp] commit failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database commit failed"})
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
