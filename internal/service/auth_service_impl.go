package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/soaringjerry/dreamhub/internal/entity"
	"github.com/soaringjerry/dreamhub/internal/repository"
	"github.com/soaringjerry/dreamhub/pkg/apperr"
	"github.com/soaringjerry/dreamhub/pkg/config" // Assuming config holds JWT secret and expiration
	"github.com/soaringjerry/dreamhub/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// Ensure authServiceImpl implements AuthService interface.
var _ AuthService = (*authServiceImpl)(nil)

type authServiceImpl struct {
	userRepo      repository.UserRepository
	jwtSecret     []byte
	jwtExpiration time.Duration
}

// jwtCustomClaims defines the structure for JWT claims.
type jwtCustomClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// NewAuthService creates a new instance of authServiceImpl.
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	// TODO: Ensure JWT_SECRET and JWT_EXPIRATION_MINUTES are loaded in config.LoadConfig()
	jwtSecret := []byte(cfg.JWTSecret) // Get secret from config
	if len(jwtSecret) == 0 {
		logger.Warn("JWT_SECRET is not set in config, using default insecure secret. SET THIS IN PRODUCTION!")
		jwtSecret = []byte("default_insecure_secret_please_change")
	}
	jwtExpiration := time.Duration(cfg.JWTExpirationMinutes) * time.Minute // Get expiration from config
	if jwtExpiration == 0 {
		logger.Warn("JWT_EXPIRATION_MINUTES is not set or zero in config, using default 60 minutes.")
		jwtExpiration = 60 * time.Minute
	}

	return &authServiceImpl{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

// Register handles new user registration.
func (s *authServiceImpl) Register(ctx context.Context, payload RegisterPayload) (*entity.SanitizedUser, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.ErrorContext(ctx, "密码哈希失败", "error", err)
		return nil, apperr.New(apperr.CodeInternal, "注册失败，请稍后重试")
	}

	// Create user entity
	user := &entity.User{
		ID:           uuid.New(), // Generate new UUID for the user
		Username:     payload.Username,
		PasswordHash: string(hashedPassword),
		// CreatedAt and UpdatedAt will be set by the repository or DB default
	}

	// Attempt to create user in the repository
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		// Check if it's a known conflict error from the repo
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Code == apperr.CodeConflict {
			return nil, err // Return the specific conflict error
		}
		// Otherwise, return a generic internal error
		logger.ErrorContext(ctx, "创建用户时仓库出错", "username", user.Username, "error", err)
		return nil, apperr.New(apperr.CodeInternal, "注册失败，请稍后重试").Wrap(err)
	}

	logger.InfoContext(ctx, "用户注册成功", "user_id", user.ID, "username", user.Username)
	sanitizedUser := user.Sanitize()
	return &sanitizedUser, nil
}

// Login authenticates a user and returns a JWT.
func (s *authServiceImpl) Login(ctx context.Context, creds LoginCredentials) (*LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetUserByUsername(ctx, creds.Username)
	if err != nil {
		var appErr *apperr.AppError
		if errors.As(err, &appErr) && appErr.Code == apperr.CodeNotFound {
			// User not found, return specific unauthenticated error
			return nil, apperr.New(apperr.CodeUnauthenticated, "用户名或密码错误")
		}
		// Other repository errors
		logger.ErrorContext(ctx, "登录时获取用户失败", "username", creds.Username, "error", err)
		return nil, apperr.New(apperr.CodeInternal, "登录失败，请稍后重试").Wrap(err)
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password))
	if err != nil {
		// If passwords don't match (bcrypt returns specific error)
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			logger.WarnContext(ctx, "密码不匹配", "username", creds.Username)
			return nil, apperr.New(apperr.CodeUnauthenticated, "用户名或密码错误")
		}
		// Other potential errors during comparison
		logger.ErrorContext(ctx, "密码比较失败", "username", creds.Username, "error", err)
		return nil, apperr.New(apperr.CodeInternal, "登录失败，请稍后重试").Wrap(err)
	}

	// Password is correct, generate JWT
	expirationTime := time.Now().Add(s.jwtExpiration)
	claims := &jwtCustomClaims{
		UserID: user.ID.String(), // Use user's UUID as string
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "dreamhub",    // Optional: identify the issuer
			Subject:   user.Username, // Optional: identify the subject
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		logger.ErrorContext(ctx, "JWT 签名失败", "username", creds.Username, "error", err)
		return nil, apperr.New(apperr.CodeInternal, "登录失败，无法生成认证令牌").Wrap(err)
	}

	logger.InfoContext(ctx, "用户登录成功", "user_id", user.ID, "username", user.Username)
	sanitizedUser := user.Sanitize()
	return &LoginResponse{
		Token: tokenString,
		User:  sanitizedUser,
	}, nil
}

// ValidateToken parses and validates a JWT string.
func (s *authServiceImpl) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what we expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			logger.WarnContext(ctx, "格式错误的 JWT", "error", err)
			return "", apperr.New(apperr.CodeUnauthenticated, "无效的认证令牌")
		} else if errors.Is(err, jwt.ErrTokenExpired) {
			logger.WarnContext(ctx, "过期的 JWT", "error", err)
			return "", apperr.New(apperr.CodeUnauthenticated, "认证令牌已过期")
		} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
			logger.WarnContext(ctx, "尚未生效的 JWT", "error", err)
			return "", apperr.New(apperr.CodeUnauthenticated, "认证令牌尚未生效")
		} else {
			logger.WarnContext(ctx, "JWT 解析/验证失败", "error", err)
			return "", apperr.New(apperr.CodeUnauthenticated, "无效的认证令牌").Wrap(err)
		}
	}

	if claims, ok := token.Claims.(*jwtCustomClaims); ok && token.Valid {
		// Optionally, check if user still exists in DB? (Might add latency)
		// _, err := s.userRepo.GetUserByID(ctx, claims.UserID)
		// if err != nil { ... return error ... }

		// Validate UserID format (should be UUID)
		_, parseErr := uuid.Parse(claims.UserID)
		if parseErr != nil {
			logger.ErrorContext(ctx, "JWT 中包含无效的用户 ID 格式", "user_id", claims.UserID, "error", parseErr)
			return "", apperr.New(apperr.CodeUnauthenticated, "认证令牌无效（用户标识错误）")
		}

		logger.DebugContext(ctx, "JWT 验证成功", "user_id", claims.UserID)
		return claims.UserID, nil
	}

	logger.WarnContext(ctx, "无效的 JWT Claims 或令牌无效")
	return "", apperr.New(apperr.CodeUnauthenticated, "无效的认证令牌")
}
