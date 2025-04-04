package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/rinefica/voice_null_sso/internal/lib/jwt"
	"github.com/rinefica/voice_null_sso/internal/lib/sl"
	"github.com/rinefica/voice_null_sso/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, email string, password string) (token string, err error)
	Register(ctx context.Context, email string, password string) (userID int64, err error)
}
type AuthServiceImpl struct {
	log          *slog.Logger
	userSaver    storage.UserSaver
	userProvider storage.UserProvider
	appProvider  storage.AppProvider
	tokenTTL     time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func New(
	log *slog.Logger,
	userSaver storage.UserSaver,
	userProvider storage.UserProvider,
	appProvider storage.AppProvider,
	tokenTTL time.Duration,
) *AuthServiceImpl {
	return &AuthServiceImpl{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

func (a *AuthServiceImpl) Login(
	ctx context.Context,
	email string,
	password string,
) (token string, err error) {
	const tag = "auth.login"
	log := a.log.With(slog.String("tag", tag))

	log.Info("start login")

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Debug("user not found", sl.Err(err))
			return "", fmt.Errorf("%s : %w", tag, ErrInvalidCredentials)
		}
		log.Error("failed get user", sl.Err(err))
		return "", fmt.Errorf("%s : %w", tag, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Debug("invalid credentials", sl.Err(err))
		return "", fmt.Errorf("%s : %w", tag, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, 1)
	if err != nil {
		log.Error("failed get app", sl.Err(err))
		return "", fmt.Errorf("%s : %w", tag, err)
	}

	log.Info("success login")

	token, err = jwt.CreateToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed create token", sl.Err(err))
		return "", fmt.Errorf("%s : %w", tag, err)
	}
	return token, nil
}

func (a *AuthServiceImpl) Register(
	ctx context.Context,
	email string,
	password string,
) (userID int64, err error) {
	const tag = "auth.register"
	log := a.log.With(slog.String("tag", tag))

	log.Info("Register user %s", email)

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", err.Error())
		return 0, fmt.Errorf("%s: %w", tag, err)
	}

	uid, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		log.Error("failed to save user", err.Error())
		return 0, fmt.Errorf("%s: %w", tag, err)
	}
	log.Debug("saved user %s", email)
	return int64(uid), nil
}
