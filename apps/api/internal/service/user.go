package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gatewarden/api/internal/model"
)

// ErrInvalidRole is returned when assigning an unknown role.
var ErrInvalidRole = errors.New("invalid role")

// validRoles are the roles supported by basic RBAC (free tier). Advanced
// role hierarchies are a premium feature.
var validRoles = map[string]bool{"admin": true, "viewer": true}

// UserService manages users and basic role assignment.
type UserService struct {
	db *pgxpool.Pool
}

func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

func (s *UserService) List(ctx context.Context) ([]model.User, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, email, role, created_at, updated_at FROM users ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("user rows: %w", err)
	}
	return users, nil
}

// SetRole assigns admin/viewer to a user. The seed admin cannot be demoted
// (there must always be an admin).
func (s *UserService) SetRole(ctx context.Context, userID, role string) error {
	if !validRoles[role] {
		return ErrInvalidRole
	}
	// Never demote the seed admin.
	var email string
	err := s.db.QueryRow(ctx, `SELECT email FROM users WHERE id = $1`, userID).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get user: %w", err)
	}
	if email == "admin@gatewarden.dev" && role != "admin" {
		return errors.New("cannot demote the seed admin")
	}

	_, err = s.db.Exec(ctx, `UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`, userID, role)
	if err != nil {
		return fmt.Errorf("set role: %w", err)
	}
	return nil
}
