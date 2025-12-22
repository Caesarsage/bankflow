package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Caesarsage/bankflow/identity-service/internal/kafka"
	"github.com/Caesarsage/bankflow/identity-service/internal/models"
	"github.com/Caesarsage/bankflow/identity-service/internal/repository"
	"github.com/Caesarsage/bankflow/identity-service/pkg/hash"
	"github.com/Caesarsage/bankflow/identity-service/pkg/jwt"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo   *repository.UserRepository
	jwtManager *jwt.JWTManager
	kafkaProducer *kafka.Producer
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repository.UserRepository,
	jwtManager *jwt.JWTManager,
	kafkaProducer *kafka.Producer) *AuthService {

	return &AuthService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		kafkaProducer: kafkaProducer,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil && err != repository.ErrUserNotFound {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		IsVerified:   false,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Publish user.registered event to Kafka
	event := &kafka.Event{
		EventID: uuid.New().String(),
		EventType: "user.registered",
		UserID: user.ID,
		Email: user.Email,
		Data: map[string]interface{}{
			"is_verified": user.IsVerified,
			"created_at": user.CreatedAt,
		},
		Timestamp: time.Now(),
	}

	go func() {
		if err := s.kafkaProducer.PublishEvent(context.Background(), event); err != nil {
			log.Printf("error publishing")
			//TODO: add message queue retry mechanism
		}
	}()

	// TODO: Send verification email

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, ErrAccountLocked
	}

	// Check if account is active
	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	// Verify password
	if !hash.CheckPassword(req.Password, user.PasswordHash) {
		// Increment failed attempts
		_ = s.userRepo.IncrementFailedLoginAttempts(ctx, user.ID)
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		CreatedAt:    time.Now(),
	}

	err = s.userRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	// Update last login
	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	event := &kafka.Event{
		EventID:   uuid.New().String(),
		EventType: "user.logged_in",
		UserID:    user.ID,
		Email:     user.Email,
		Data: map[string]interface{}{
			"ip_address": ipAddress,
			"user_agent": userAgent,
			"login_time": time.Now(),
		},
		Timestamp: time.Now(),
	}

	go func() {
		_ = s.kafkaProducer.PublishEvent(context.Background(), event)
	}()

	user.PasswordHash = ""

	return &models.LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// RefreshToken generates new access token from refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get session
	session, err := s.userRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if account is active
	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	// Generate new tokens
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Update session with new refresh token
	_ = s.userRepo.DeleteSession(ctx, refreshToken)
	newSession := &models.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		IPAddress:    session.IPAddress,
		UserAgent:    session.UserAgent,
		CreatedAt:    time.Now(),
	}
	err = s.userRepo.CreateSession(ctx, newSession)
	if err != nil {
		return nil, err
	}

	// Remove password hash from response
	user.PasswordHash = ""

	return &models.LoginResponse{
		User:         user,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}

// Logout invalidates refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.userRepo.DeleteSession(ctx, refreshToken)
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Remove password hash
	user.PasswordHash = ""
	return user, nil
}

// Reset password
// func (s *AuthService) ResetPassword(ctx context.Context, email string) (string, error) {
// 	user, err := s.userRepo.GetUserByEmail(ctx, email)

// 	if err != nil {
// 		return "", err
// 	}

// 	// Check if account is active
// 	if !user.IsActive {
// 		return "", ErrAccountInactive
// 	}

// }

// Verify email/phone
