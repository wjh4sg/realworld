package jwt

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/onexstack/realworld/apiserver/cache"
)

const (
	accessTokenSubject  = "access"
	refreshTokenSubject = "refresh"
)

// Claims defines the JWT payload used by the API.
type Claims struct {
	UserID int64  `json:"user_id"`
	JTI    string `json:"jti"`
	jwt.RegisteredClaims
}

// Manager manages access and refresh tokens.
type Manager struct {
	secret     string
	jwtCache   cache.JWTCache
	tokenTTL   time.Duration
	refreshTTL time.Duration
}

// NewManager creates a new JWT manager.
func NewManager(secret string, jwtCache cache.JWTCache) *Manager {
	return &Manager{
		secret:     secret,
		jwtCache:   jwtCache,
		tokenTTL:   24 * time.Hour,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

// GenerateToken creates an access token.
func (m *Manager) GenerateToken(userID int64) (string, error) {
	return m.generateToken(userID, accessTokenSubject, m.tokenTTL)
}

// GenerateRefreshToken creates a refresh token.
func (m *Manager) GenerateRefreshToken(userID int64) (string, error) {
	return m.generateToken(userID, refreshTokenSubject, m.refreshTTL)
}

// ValidateToken validates an access token.
// Kept for backward compatibility with older call sites.
func (m *Manager) ValidateToken(ctx context.Context, tokenString string) (int64, error) {
	return m.ValidateAccessToken(ctx, tokenString)
}

// ValidateAccessToken validates an access token.
func (m *Manager) ValidateAccessToken(ctx context.Context, tokenString string) (int64, error) {
	return m.validateToken(ctx, tokenString, accessTokenSubject)
}

// ValidateRefreshToken validates a refresh token.
func (m *Manager) ValidateRefreshToken(ctx context.Context, tokenString string) (int64, error) {
	return m.validateToken(ctx, tokenString, refreshTokenSubject)
}

// RevokeToken revokes a valid token.
func (m *Manager) RevokeToken(ctx context.Context, tokenString string) error {
	token, claims, err := m.parseToken(tokenString)
	if err != nil || !token.Valid {
		return errors.New("invalid token")
	}

	if m.jwtCache != nil {
		if err := m.jwtCache.AddToBlacklist(ctx, claims.JTI); err != nil {
			return err
		}

		tokenHash := m.generateTokenHash(tokenString)
		_ = m.jwtCache.InvalidateValidationCache(ctx, tokenHash)
	}

	return nil
}

// RefreshToken exchanges a valid refresh token for a new access token.
func (m *Manager) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	userID, err := m.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	_ = m.RevokeToken(ctx, refreshToken)

	return m.GenerateToken(userID)
}

func (m *Manager) generateToken(userID int64, subject string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		JTI:    uuid.New().String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "realworld-api",
			Subject:   subject,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secret))
}

func (m *Manager) validateToken(ctx context.Context, tokenString, expectedSubject string) (int64, error) {
	tokenHash := m.generateTokenHash(tokenString)

	if m.jwtCache != nil {
		userID, err := m.jwtCache.GetValidationResult(ctx, tokenHash)
		if err == nil && userID > 0 {
			return userID, nil
		}
	}

	token, claims, err := m.parseToken(tokenString)
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	if expectedSubject != "" && claims.Subject != expectedSubject {
		return 0, errors.New("invalid token type")
	}

	if m.jwtCache != nil {
		isBlacklisted, err := m.jwtCache.IsInBlacklist(ctx, claims.JTI)
		if err == nil && isBlacklisted {
			return 0, errors.New("token has been revoked")
		}
	}

	if m.jwtCache != nil {
		_ = m.jwtCache.SetValidationResult(ctx, tokenHash, claims.UserID)
	}

	return claims.UserID, nil
}

func (m *Manager) parseToken(tokenString string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(m.secret), nil
	})
	if err != nil {
		return nil, nil, err
	}

	return token, claims, nil
}

func (m *Manager) generateTokenHash(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}
