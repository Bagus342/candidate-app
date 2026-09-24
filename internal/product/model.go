package product

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrNotFound = errors.New("product not found")

type Product struct {
	ID          int64
	Name        string
	Description string
	Price       int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Repository interface {
	List(context.Context) ([]Product, error)
	Get(context.Context, int64) (Product, error)
	Create(context.Context, Product) (Product, error)
	Update(context.Context, Product) (Product, error)
	Delete(context.Context, int64) error
}

func (p Product) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return errors.New("name is required")
	}
	if len(p.Name) > 120 {
		return errors.New("name must be 120 characters or fewer")
	}
	if p.Price < 0 {
		return errors.New("price cannot be negative")
	}
	return nil
}
