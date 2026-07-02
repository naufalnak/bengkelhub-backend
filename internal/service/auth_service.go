package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/naufalnak/bengkelhub-backend/config"
	"github.com/naufalnak/bengkelhub-backend/internal/domain"
	"github.com/naufalnak/bengkelhub-backend/internal/repository"
	"github.com/naufalnak/bengkelhub-backend/pkg/jwt"
	"github.com/naufalnak/bengkelhub-backend/pkg/tasks"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(req *domain.LoginRequest) (*domain.AuthResponse, error)
	Me(userID uuid.UUID) (*domain.UserResponse, error)
	VerifyEmail(token string) error
	ResendVerification(userID uuid.UUID) error
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo}
}

// generateVerifyToken menghasilkan token random 32-byte hex (64 karakter)
func generateVerifyToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *authService) Register(req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Cek email sudah terdaftar
	_, err := s.userRepo.FindByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email already registered")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Default role
	role := req.Role
	if role == "" {
		role = domain.RoleCustomer
	}

	// Generate verify token
	token, err := generateVerifyToken()
	if err != nil {
		return nil, err
	}
	exp := time.Now().Add(24 * time.Hour)

	user := &domain.User{
		Name:           req.Name,
		Email:          req.Email,
		Password:       string(hashed),
		Phone:          req.Phone,
		Role:           role,
		EmailVerified:  false,
		VerifyToken:    token,
		VerifyTokenExp: &exp,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Kirim email verifikasi async (non-blocking, gak ganggu response Register)
	verifyLink := fmt.Sprintf("%s/verify-email?token=%s",
		config.Cfg.AppBaseURL, token)
	_ = tasks.EnqueueVerificationEmail(tasks.VerificationEmailPayload{
		ToEmail:    user.Email,
		ToName:     user.Name,
		VerifyLink: verifyLink,
	})

	jwtToken, err := jwt.Generate(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: jwtToken,
		User:  user.ToResponse(),
	}, nil
}

func (s *authService) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := jwt.Generate(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

func (s *authService) Me(userID uuid.UUID) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	resp := user.ToResponse()
	return &resp, nil
}

func (s *authService) VerifyEmail(token string) error {
	user, err := s.userRepo.FindByVerifyToken(token)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("invalid or expired verification link")
	}
	if err != nil {
		return err
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	if user.VerifyTokenExp != nil && time.Now().After(*user.VerifyTokenExp) {
		return errors.New("verification link has expired, please request a new one")
	}

	user.EmailVerified = true
	user.VerifyToken = ""
	user.VerifyTokenExp = nil

	return s.userRepo.Update(user)
}

func (s *authService) ResendVerification(userID uuid.UUID) error {
	user, err := s.userRepo.FindByID(userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user not found")
	}
	if err != nil {
		return err
	}

	if user.EmailVerified {
		return errors.New("email already verified")
	}

	// Batasi resend: minimal 2 menit setelah request terakhir
	if user.VerifyTokenExp != nil {
		minNextRequest := user.VerifyTokenExp.Add(-22 * time.Hour) // 24h exp - 22h = 2 menit window
		if time.Now().Before(minNextRequest) {
			return errors.New("please wait at least 2 minutes before requesting a new verification email")
		}
	}

	token, err := generateVerifyToken()
	if err != nil {
		return err
	}
	exp := time.Now().Add(24 * time.Hour)

	user.VerifyToken = token
	user.VerifyTokenExp = &exp

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	verifyLink := fmt.Sprintf("%s/verify-email?token=%s",
		config.Cfg.AppBaseURL, token)
	return tasks.EnqueueVerificationEmail(tasks.VerificationEmailPayload{
		ToEmail:    user.Email,
		ToName:     user.Name,
		VerifyLink: verifyLink,
	})
}
