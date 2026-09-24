package product

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context) ([]Product, error) {
	return s.repository.List(ctx)
}

func (s *Service) Get(ctx context.Context, id int64) (Product, error) {
	return s.repository.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, product Product) (Product, error) {
	if err := product.Validate(); err != nil {
		return Product{}, err
	}
	return s.repository.Create(ctx, product)
}

func (s *Service) Update(ctx context.Context, product Product) (Product, error) {
	if err := product.Validate(); err != nil {
		return Product{}, err
	}
	return s.repository.Update(ctx, product)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repository.Delete(ctx, id)
}
