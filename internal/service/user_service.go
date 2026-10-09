package service

import (
	"errors"

	"ewallet/internal/model"
	"ewallet/internal/repository"
	"ewallet/internal/security"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db      *gorm.DB
	users   *repository.UserRepository
	wallets *repository.WalletRepository
	jwt     *security.JWTManager
}

func NewUserService(
	db *gorm.DB,
	users *repository.UserRepository,
	wallets *repository.WalletRepository,
	jwt *security.JWTManager,
) *UserService {
	return &UserService{db: db, users: users, wallets: wallets, jwt: jwt}
}

func (s *UserService) Register(req model.UserRegisterRequest) (*model.User, error) {
	exists, err := s.users.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailTaken
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.users.Create(tx, user); err != nil {
			return err
		}
		wallet := &model.Wallet{UserID: user.ID, Balance: 0}
		return s.wallets.Create(tx, wallet)
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(req model.UserLoginRequest) (string, error) {
	user, err := s.users.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !security.CheckPassword(user.PasswordHash, req.Password) {
		return "", ErrInvalidCredentials
	}

	return s.jwt.Generate(user.ID)
}

func (s *UserService) FindByID(id uuid.UUID) (*model.User, error) {
	return s.users.FindByID(id)
}
