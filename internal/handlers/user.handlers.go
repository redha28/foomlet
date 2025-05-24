package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redha28/foomlet/internal/config"
	"github.com/redha28/foomlet/internal/middlewares"
	"github.com/redha28/foomlet/internal/models"
	"github.com/redha28/foomlet/internal/repositories"
	"github.com/redha28/foomlet/pkg"
)

type UserHandler struct {
	repo   repositories.UserRepoInterface
	config *config.Config
}

func NewUserHandler(repo repositories.UserRepoInterface) *UserHandler {
	return &UserHandler{
		repo:   repo,
		config: config.GetConfig(),
	}
}

func (u *UserHandler) Login(c *gin.Context) {
	var loginReq models.UserLogin
	response := models.NewResponse(c)

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		if strings.Contains(err.Error(), "Field validation for 'Pin'") {
			response.BadRequest("Pin length must be exactly 6 characters", err.Error())
			return
		}
		response.BadRequest("Invalid input", err.Error())
		return
	}

	// Fetch user from repository
	user, err := u.repo.GetUserByPhone(c, loginReq.Phone)
	if err != nil {
		response.Unauthorized("Invalid credentials", "Phone Number dan PIN tidak cocok")
		return
	}

	// Check password/PIN
	hash := pkg.InitHashConfig()
	hash.UseDefaultConfig()
	isValid, err := hash.CompareHashAndPassword(user.Pin, loginReq.Pin)
	if err != nil {
		response.InternalServerError("Failed to verify PIN", err.Error())
		return
	}
	if !isValid {
		response.Unauthorized("Invalid credentials", "Phone Number dan PIN tidak cocok")
		return
	}

	// Generate JWT tokens using config values
	jwtUtil := pkg.NewJwtUtil(
		u.config.JWT.AccessSecret,
		u.config.JWT.RefreshSecret,
		u.config.JWT.AccessExpiry,
		u.config.JWT.RefreshExpiry,
	)

	accessToken, err := jwtUtil.GenerateAccessToken(user.ID)
	if err != nil {
		response.InternalServerError("Failed to generate token", err.Error())
		return
	}

	refreshToken, err := jwtUtil.GenerateRefreshToken(user.ID)
	if err != nil {
		response.InternalServerError("Failed to generate refresh token", err.Error())
		return
	}

	// Return response
	c.JSON(200, gin.H{
		"status": "SUCCESS",
		"result": models.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

func (u *UserHandler) Register(c *gin.Context) {
	var userReq models.UserRegist
	response := models.NewResponse(c)
	if err := c.ShouldBindJSON(&userReq); err != nil {
		if strings.Contains(err.Error(), "Field validation for 'Pin'") {
			response.BadRequest("Pin length must be exactly 6 characters", err.Error())
			return
		}
		response.BadRequest("Invalid input", err.Error())
		return
	}
	hash := pkg.InitHashConfig()
	hash.UseDefaultConfig()
	hadhedPin, err := hash.GenHashedPassword(userReq.Pin)
	if err != nil {
		response.BadRequest("Failed to hash password", err.Error())
		return
	}
	result, err := u.repo.UseRegister(c, userReq, hadhedPin)
	if err != nil {
		if strings.Contains(err.Error(), "user already exists") {
			response.BadRequest("User already exists", err.Error())
			return
		}
		response.InternalServerError("Failed to register user", err.Error())
		return
	}
	response.Created("User registered successfully", result)
}

func (u *UserHandler) UpdateProfile(c *gin.Context) {
	response := models.NewResponse(c)

	// Get user ID from the context (set by AuthMiddleware)
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		response.Unauthorized("User not authenticated", nil)
		return
	}

	// Parse request body
	var updateReq models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response.BadRequest("Invalid input", err.Error())
		return
	}

	// Validate required fields
	if updateReq.Firstname == "" || updateReq.Lastname == "" {
		response.BadRequest("Nama Firstname atau Lastname tidak boleh kosong", nil)
		return
	}

	// Call repository to update profile
	result, err := u.repo.UpdateUserProfile(c, userID, updateReq)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			response.NotFound("User not found", err.Error())
			return
		}
		response.InternalServerError("Failed to update profile", err.Error())
		return
	}

	// Return success response
	c.JSON(200, gin.H{
		"status": "SUCCESS",
		"result": result,
	})
}

func (u *UserHandler) RefreshToken(c *gin.Context) {
	var refreshReq models.RefreshTokenRequest
	response := models.NewResponse(c)

	if err := c.ShouldBindJSON(&refreshReq); err != nil {
		response.BadRequest("Invalid input", err.Error())
		return
	}

	// Validate refresh token using config values
	jwtUtil := pkg.NewJwtUtil(
		u.config.JWT.AccessSecret,
		u.config.JWT.RefreshSecret,
		u.config.JWT.AccessExpiry,
		u.config.JWT.RefreshExpiry,
	)

	claims, err := jwtUtil.ValidateRefreshToken(refreshReq.RefreshToken)
	if err != nil {
		response.Unauthorized("Invalid refresh token", err.Error())
		return
	}

	// Generate new access token
	accessToken, err := jwtUtil.GenerateAccessToken(claims.UserID)
	if err != nil {
		response.InternalServerError("Failed to generate token", err.Error())
		return
	}

	// Return response
	c.JSON(200, gin.H{
		"status": "SUCCESS",
		"result": models.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshReq.RefreshToken,
		},
	})
}
