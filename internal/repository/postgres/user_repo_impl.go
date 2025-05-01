package postgres

import (
	"context"
	"errors" // Import errors package
	"time"   // Import time package

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn" // Import for pgconn.PgError
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/apperr" // Import apperr for specific errors
	"github.com/soaringjerry/dreamhub/pkg/logger"
)

// Ensure postgresUserRepository implements UserRepository interface.
var _ repository.UserRepository = (*postgresUserRepository)(nil)

type postgresUserRepository struct {
	db *pgxpool.Pool
}

// NewPostgresUserRepository creates a new instance of postgresUserRepository.
func NewPostgresUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &postgresUserRepository{db: db}
}

// CreateUser inserts a new user record into the database.
func (r *postgresUserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	// Ensure ID is generated if not provided (though default in DB handles this)
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := user.CreatedAt // Use provided time or set new one
	if now.IsZero() {
		now = time.Now()
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.Exec(ctx, query, user.ID, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		// Check for unique constraint violation (duplicate username)
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 is unique_violation
			logger.WarnContext(ctx, "尝试创建已存在的用户", "username", user.Username, "error", err)
			// Return a specific application error using Wrap function
			return apperr.Wrap(err, apperr.CodeConflict, "用户名已存在")
		}
		logger.ErrorContext(ctx, "创建用户时数据库出错", "username", user.Username, "error", err)
		// Use Wrap for internal errors as well
		return apperr.Wrap(err, apperr.CodeInternal, "创建用户失败")
	}
	logger.InfoContext(ctx, "用户创建成功", "user_id", user.ID, "username", user.Username)
	return nil
}

// GetUserByUsername retrieves a user by their username.
func (r *postgresUserRepository) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	var user entity.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WarnContext(ctx, "尝试获取不存在的用户（按用户名）", "username", username)
			// Use Wrap for not found errors
			return nil, apperr.Wrap(err, apperr.CodeNotFound, "用户未找到")
		}
		logger.ErrorContext(ctx, "按用户名获取用户时数据库出错", "username", username, "error", err)
		// Use Wrap for internal errors
		return nil, apperr.Wrap(err, apperr.CodeInternal, "获取用户信息失败")
	}
	return &user, nil
}

// GetUserByID retrieves a user by their ID.
func (r *postgresUserRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	// Validate if the provided ID is a valid UUID
	userID, err := uuid.Parse(id)
	if err != nil {
		logger.WarnContext(ctx, "尝试使用无效的 UUID 获取用户", "invalid_id", id)
		// Use CodeInvalidArgument instead of CodeBadRequest and use Wrap
		return nil, apperr.Wrap(err, apperr.CodeInvalidArgument, "无效的用户 ID 格式")
	}

	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user entity.User
	err = r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.WarnContext(ctx, "尝试获取不存在的用户（按 ID）", "user_id", id)
			// Use Wrap for not found errors
			return nil, apperr.Wrap(err, apperr.CodeNotFound, "用户未找到")
		}
		logger.ErrorContext(ctx, "按 ID 获取用户时数据库出错", "user_id", id, "error", err)
		// Use Wrap for internal errors
		return nil, apperr.Wrap(err, apperr.CodeInternal, "获取用户信息失败")
	}
	return &user, nil
}
