package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type Alteon struct {
	ID                 int64
	Name               string
	BaseURL            string
	Username           string
	Password           string
	InsecureSkipVerify bool
	CACert             string
	Enabled            bool
}

type AlteonsRepo struct {
	db *sql.DB
}

func NewAlteonsRepo(s *Storage) *AlteonsRepo {
	return &AlteonsRepo{db: s.db}
}

func (r *AlteonsRepo) List(ctx context.Context) ([]Alteon, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT id, name, base_url, username, password, insecure_skip_verify, COALESCE(ca_cert, ''), enabled
        FROM alteons
        ORDER BY name
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Alteon
	for rows.Next() {
		var a Alteon
		if err := rows.Scan(&a.ID, &a.Name, &a.BaseURL, &a.Username, &a.Password,
			&a.InsecureSkipVerify, &a.CACert, &a.Enabled); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *AlteonsRepo) Create(ctx context.Context, a Alteon) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO alteons (name, base_url, username, password, insecure_skip_verify, ca_cert, enabled)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `, a.Name, a.BaseURL, a.Username, a.Password, a.InsecureSkipVerify, nullIfEmpty(a.CACert), a.Enabled).Scan(&id)
	return id, err
}

func (r *AlteonsRepo) DeleteByName(ctx context.Context, name string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM alteons WHERE name = $1`, name)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("alteon %q no encontrado", name)
	}
	return nil
}

func (r *AlteonsRepo) SetEnabled(ctx context.Context, name string, enabled bool) error {
	res, err := r.db.ExecContext(ctx, `
        UPDATE alteons SET enabled = $1, updated_at = NOW() WHERE name = $2
    `, enabled, name)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("alteon %q no encontrado", name)
	}
	return nil
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
