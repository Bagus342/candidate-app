package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"candidate-app/internal/product"
)

func TestApplicationListsProducts(t *testing.T) {
	store := &memoryStore{products: []product.Product{{ID: 1, Name: "Notebook", Description: "A5", Price: 1250}}}
	handler, err := NewHandler(product.NewService(store), "../../web/templates")
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), "Notebook") || !strings.Contains(response.Body.String(), "$12.50") {
		t.Fatalf("response does not contain rendered product: %s", response.Body.String())
	}
}

func TestApplicationCreatesAndDeletesProduct(t *testing.T) {
	store := &memoryStore{}
	handler, err := NewHandler(product.NewService(store), "../../web/templates")
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	createRequest := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader("name=Pen&description=Blue&price=250"))
	createRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusSeeOther || len(store.products) != 1 {
		t.Fatalf("create response = %d, products = %d", createResponse.Code, len(store.products))
	}

	deleteRequest := httptest.NewRequest(http.MethodPost, "/products/1", strings.NewReader("_method=DELETE"))
	deleteRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	deleteResponse := httptest.NewRecorder()
	handler.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusSeeOther || len(store.products) != 0 {
		t.Fatalf("delete response = %d, products = %d", deleteResponse.Code, len(store.products))
	}
}

type memoryStore struct {
	products []product.Product
}

func (m *memoryStore) List(context.Context) ([]product.Product, error) { return m.products, nil }
func (m *memoryStore) Get(_ context.Context, id int64) (product.Product, error) {
	for _, item := range m.products {
		if item.ID == id { return item, nil }
	}
	return product.Product{}, product.ErrNotFound
}
func (m *memoryStore) Create(_ context.Context, item product.Product) (product.Product, error) {
	item.ID = int64(len(m.products) + 1)
	m.products = append(m.products, item)
	return item, nil
}
func (m *memoryStore) Update(_ context.Context, item product.Product) (product.Product, error) {
	for index := range m.products {
		if m.products[index].ID == item.ID { m.products[index] = item; return item, nil }
	}
	return product.Product{}, product.ErrNotFound
}
func (m *memoryStore) Delete(_ context.Context, id int64) error {
	for index := range m.products {
		if m.products[index].ID == id { m.products = append(m.products[:index], m.products[index+1:]...); return nil }
	}
	return product.ErrNotFound
}
