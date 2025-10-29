package routes

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/adii2ma/dbms-backend/database"
	"github.com/adii2ma/dbms-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type createRequestInput struct {
	Type        string  `json:"type" binding:"required"`
	Description *string `json:"description"`
	UserID      string  `json:"user_id"`
	RoomID      *int    `json:"room_id"`
	RoomNumber  string  `json:"room_number"`
	Block       string  `json:"block"`
}

var errActiveRequestExists = errors.New("active request already exists for this room and type")
var errRoomBlockMismatch = errors.New("room does not belong to provided block")

// CreateRequest handles creation of a new cleaning or maintenance request.
func CreateRequest(c *gin.Context) {
	var input createRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	requestType := models.RequestType(strings.ToLower(strings.TrimSpace(input.Type)))
	if requestType != models.RequestTypeCleaning && requestType != models.RequestTypeMaintenance {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported request type",
		})
		return
	}

	if input.Description != nil {
		trimmed := strings.TrimSpace(*input.Description)
		if trimmed == "" {
			input.Description = nil
		} else {
			input.Description = &trimmed
		}
	}

	ctx := c.Request.Context()

	var userID *uuid.UUID
	roomNumber := strings.TrimSpace(input.RoomNumber)
	block := strings.TrimSpace(input.Block)

	if input.UserID != "" {
		parsed, err := uuid.Parse(input.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user_id",
			})
			return
		}

		var user models.User
		if err := database.DB.NewSelect().Model(&user).Where("id = ?", parsed).Scan(ctx); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "User not found",
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to look up user",
			})
			return
		}

		userID = &parsed

		if roomNumber == "" && user.RoomName != nil {
			cleaned := strings.TrimSpace(*user.RoomName)
			if cleaned != "" {
				roomNumber = cleaned
			}
		}

		if block == "" && user.Block != nil {
			trimmed := strings.TrimSpace(*user.Block)
			if trimmed != "" {
				block = trimmed
			}
		}
	}

	if input.RoomID == nil && (roomNumber == "" || block == "") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room_number and block or room_id is required",
		})
		return
	}

	var createdRequest *models.Request

	err := database.DB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var room models.Room
		var roomID int

		if input.RoomID != nil {
			roomID = *input.RoomID
			if err := tx.NewSelect().Model(&room).Where("id = ?", roomID).Scan(ctx); err != nil {
				return err
			}
			if block != "" && !strings.EqualFold(room.Block, block) {
				return errRoomBlockMismatch
			}
		} else {
			err := tx.NewSelect().
				Model(&room).
				Where("room_number = ?", roomNumber).
				Where("block = ?", block).
				Scan(ctx)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					room = models.Room{RoomNumber: roomNumber, Block: block}
					if _, err := tx.NewInsert().Model(&room).Exec(ctx); err != nil {
						return err
					}
				} else {
					return err
				}
			}
			roomID = room.ID
			block = room.Block
		}

		if userID != nil {
			exists, err := tx.NewSelect().
				Model((*models.RoomMember)(nil)).
				Where("room_id = ?", roomID).
				Where("block = ?", block).
				Where("user_id = ?", *userID).
				Exists(ctx)
			if err != nil {
				return err
			}

			if !exists {
				member := &models.RoomMember{RoomID: roomID, Block: block, UserID: *userID}
				if _, err := tx.NewInsert().Model(member).Exec(ctx); err != nil {
					return err
				}
			}
		}

		exists, err := tx.NewSelect().
			Model((*models.Request)(nil)).
			Where("room_id = ?", roomID).
			Where("type = ?", requestType).
			Where("status = ?", models.RequestStatusActive).
			Exists(ctx)
		if err != nil {
			return err
		}

		if exists {
			return errActiveRequestExists
		}

		request := &models.Request{
			RoomID: roomID,
			Type:   requestType,
			Status: models.RequestStatusActive,
		}

		if input.Description != nil {
			request.Description = input.Description
		}

		if userID != nil {
			request.UserID = userID
		}

		if _, err := tx.NewInsert().Model(request).Exec(ctx); err != nil {
			return err
		}

		if err := tx.NewSelect().Model(request).WherePK().Scan(ctx); err != nil {
			return err
		}

		createdRequest = request
		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Room not found",
			})
		case errors.Is(err, errActiveRequestExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "An active request already exists for this room and type",
			})
		case errors.Is(err, errRoomBlockMismatch):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "room does not belong to provided block",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create request",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Request created successfully",
		"request": createdRequest,
	})
}

// GetActiveRequest resolves the active request for a room/type combination.
func GetActiveRequest(c *gin.Context) {
	typeParam := strings.TrimSpace(c.Query("type"))
	requestType := models.RequestType(strings.ToLower(typeParam))
	if requestType == "" {
		requestType = models.RequestTypeCleaning
	}

	if requestType != models.RequestTypeCleaning && requestType != models.RequestTypeMaintenance {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unsupported request type",
		})
		return
	}

	roomIDParam := strings.TrimSpace(c.Query("room_id"))
	roomNumber := strings.TrimSpace(c.Query("room_number"))
	blockParam := strings.TrimSpace(c.Query("block"))

	ctx := c.Request.Context()
	var roomID int

	if roomIDParam != "" {
		parsedID, err := strconv.Atoi(roomIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid room_id",
			})
			return
		}
		roomID = parsedID
	} else if roomNumber != "" {
		if blockParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "block query parameter is required when using room_number",
			})
			return
		}
		var room models.Room
		if err := database.DB.NewSelect().
			Model(&room).
			Where("room_number = ?", roomNumber).
			Where("block = ?", blockParam).
			Scan(ctx); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusOK, gin.H{"request": nil})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to look up room",
			})
			return
		}
		roomID = room.ID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room_id or room_number query parameter is required",
		})
		return
	}

	request := new(models.Request)
	if err := database.DB.NewSelect().
		Model(request).
		Relation("Room").
		Relation("User").
		Where("room_id = ?", roomID).
		Where("type = ?", requestType).
		Where("status = ?", models.RequestStatusActive).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, gin.H{"request": nil})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve active request",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"request": request})
}
