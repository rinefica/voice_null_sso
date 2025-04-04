package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rinefica/voice_null_sso/internal/domain/model"
	"github.com/rinefica/voice_null_sso/internal/lib/sl"
	"log/slog"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	ErrAppNotFound       = errors.New("app not found")
)

type UserSaver interface {
	SaveUser(
		ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (user *model.User, err error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (app *model.App, err error)
}

type Storage struct {
	log  *slog.Logger
	pool *pgxpool.Pool
}

// Инициализация хранилища.
func NewStorage(
	log *slog.Logger,
	storagePath string,
) (*Storage, error) {
	const tag = "storage.CreateStorage"
	logTag := log.With(slog.String("tag", tag))

	poolConfig, err := pgxpool.ParseConfig(storagePath)
	if err != nil {
		logTag.Info("Unable to parse DATABASE_URL:", sl.Err(err))
		return nil, err
	}

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		logTag.Info("Unable to create connection pool:", sl.Err(err))
		return nil, err
	}

	return &Storage{
		log:  log,
		pool: db,
	}, nil
}

// Закрытие хранилища.
func (s *Storage) Close() {
	s.pool.Close()
}

func (s *Storage) SaveUser(
	ctx context.Context,
	email string,
	passHash []byte,
) (uid int64, err error) {
	const tag = "storage.SaveUser"
	log := s.log.With(slog.String("tag", tag))

	row := s.pool.QueryRow(ctx, saveUserQuery, email, passHash)
	var userID int
	err = row.Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			log.Info("user already exists")
			return 0, ErrUserAlreadyExists
		}
		log.Info("Can't save url in table " + err.Error())
		return 0, err
	}

	return int64(userID), nil
}

func (s *Storage) User(
	ctx context.Context,
	email string,
) (user *model.User, err error) {
	const tag = "storage.User"
	log := s.log.With(slog.String("tag", tag))

	row := s.pool.QueryRow(ctx, getUserQuery, email)
	u := model.User{}
	err = row.Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				log.Info("user already exists")
				return nil, ErrUserAlreadyExists
			default:
				log.Info("error " + pgErr.Error())
				return nil, ErrUserNotFound
			}
		}
		log.Info("Can't find user in table " + err.Error())
		return nil, fmt.Errorf("unable to scan row: %w", err)
	}

	return &u, nil
}

func (s *Storage) App(
	ctx context.Context,
	appID int,
) (app *model.App, err error) {
	const tag = "storage.App"
	log := s.log.With(slog.String("tag", tag))
	log.Info("Getting app from database")

	return &model.App{
		ID:     appID,
		Name:   "sso",
		Secret: "somesecret",
	}, nil
}

const (
	saveUserQuery = `INSERT INTO users VALUES (DEFAULT, $1, $2) returning (id);`
	getUserQuery  = `SELECT id, email, pass_hash FROM users WHERE email = $1`
)
