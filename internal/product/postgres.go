package product

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) List(ctx context.Context) ([]Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, description, price, created_at, updated_at FROM products ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]Product, 0)
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func (r *PostgresRepository) Get(ctx context.Context, id int64) (Product, error) {
	var product Product
	err := r.pool.QueryRow(ctx, `SELECT id, name, description, price, created_at, updated_at FROM products WHERE id = $1`, id).
		Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CreatedAt, &product.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return product, err
}

func (r *PostgresRepository) Create(ctx context.Context, product Product) (Product, error) {
	err := r.pool.QueryRow(ctx, `INSERT INTO products (name, description, price) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`, product.Name, product.Description, product.Price).
		Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	return product, err
}

func (r *PostgresRepository) Update(ctx context.Context, product Product) (Product, error) {
	err := r.pool.QueryRow(ctx, `UPDATE products SET name = $1, description = $2, price = $3, updated_at = NOW() WHERE id = $4 RETURNING updated_at`, product.Name, product.Description, product.Price, product.ID).
		Scan(&product.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return product, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
