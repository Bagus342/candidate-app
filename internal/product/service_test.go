package product

import (
	"context"
	"testing"
)

func TestServiceCreateRejectsInvalidProduct(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	if _, err := service.Create(context.Background(), Product{Price: 100}); err == nil {
		t.Fatal("expected validation error")
	}
	if repository.created {
		t.Fatal("repository should not be called for invalid input")
	}
}

func TestServiceCreateDelegatesValidProduct(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	created, err := service.Create(context.Background(), Product{Name: "Notebook", Price: 1250})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	if !repository.created || created.Name != "Notebook" {
		t.Fatalf("unexpected create result: %+v", created)
	}
}

type fakeRepository struct {
	created bool
}

func (f *fakeRepository) List(context.Context) ([]Product, error) { return nil, nil }
func (f *fakeRepository) Get(context.Context, int64) (Product, error) { return Product{}, ErrNotFound }
func (f *fakeRepository) Create(_ context.Context, product Product) (Product, error) {
	f.created = true
	product.ID = 1
	return product, nil
}
func (f *fakeRepository) Update(context.Context, Product) (Product, error) { return Product{}, nil }
func (f *fakeRepository) Delete(context.Context, int64) error { return nil }
