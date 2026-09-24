package web

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"candidate-app/internal/product"
)

type ProductService interface {
	List(context.Context) ([]product.Product, error)
	Get(context.Context, int64) (product.Product, error)
	Create(context.Context, product.Product) (product.Product, error)
	Update(context.Context, product.Product) (product.Product, error)
	Delete(context.Context, int64) error
}

type Handler struct {
	service   ProductService
	templates map[string]*template.Template
}

type pageData struct {
	Title    string
	Product  product.Product
	Products []product.Product
	Error    string
}

func NewHandler(service ProductService, templateDir string) (*Handler, error) {
	funcs := template.FuncMap{
		"formatPrice": func(cents int64) string {
			return strconv.FormatInt(cents/100, 10) + "." + fmt.Sprintf("%02d", cents%100)
		},
	}
	products, err := template.New("pages").Funcs(funcs).ParseFiles(templateDir+"/layout.html", templateDir+"/products.html")
	if err != nil {
		return nil, err
	}
	productForm, err := template.New("pages").Funcs(funcs).ParseFiles(templateDir+"/layout.html", templateDir+"/product-form.html")
	if err != nil {
		return nil, err
	}
	return &Handler{
		service: service,
		templates: map[string]*template.Template{
			"products":     products,
			"product-form": productForm,
		},
	}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/" && r.Method == http.MethodGet:
		h.list(w, r)
	case r.URL.Path == "/products/new" && r.Method == http.MethodGet:
		h.form(w, r, product.Product{})
	case r.URL.Path == "/products" && r.Method == http.MethodPost:
		h.create(w, r)
	case strings.HasPrefix(r.URL.Path, "/products/"):
		h.productAction(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.List(r.Context())
	if err != nil {
		h.serverError(w, err)
		return
	}
	h.render(w, "products", pageData{Title: "Products", Products: products})
}

func (h *Handler) form(w http.ResponseWriter, r *http.Request, item product.Product) {
	h.render(w, "product-form", pageData{Title: "Product form", Product: item})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	item := productFromForm(r)
	if _, err := h.service.Create(r.Context(), item); err != nil {
		h.render(w, "product-form", pageData{Title: "Product form", Product: item, Error: err.Error()})
		return
	}
	h.redirect(w, r, "/")
}

func (h *Handler) productAction(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, "/products/"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		item, err := h.service.Get(r.Context(), id)
		if errors.Is(err, product.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			h.serverError(w, err)
			return
		}
		h.form(w, r, item)
	case http.MethodPost:
		if r.FormValue("_method") == http.MethodDelete {
			if err := h.service.Delete(r.Context(), id); err != nil {
				h.serverError(w, err)
				return
			}
			h.redirect(w, r, "/")
			return
		}
		item := productFromForm(r)
		item.ID = id
		if _, err := h.service.Update(r.Context(), item); err != nil {
			h.render(w, "product-form", pageData{Title: "Product form", Product: item, Error: err.Error()})
			return
		}
		h.redirect(w, r, "/")
	case http.MethodDelete:
		if err := h.service.Delete(r.Context(), id); err != nil {
			h.serverError(w, err)
			return
		}
		h.redirect(w, r, "/")
	default:
		w.Header().Set("Allow", "GET, POST, DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func productFromForm(r *http.Request) product.Product {
	price, _ := strconv.ParseInt(r.FormValue("price"), 10, 64)
	return product.Product{Name: strings.TrimSpace(r.FormValue("name")), Description: strings.TrimSpace(r.FormValue("description")), Price: price}
}

func (h *Handler) render(w http.ResponseWriter, name string, data pageData) {
	templates, ok := h.templates[name]
	if !ok {
		h.serverError(w, fmt.Errorf("template %q not found", name))
		return
	}
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		h.serverError(w, err)
	}
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request, location string) {
	http.Redirect(w, r, location, http.StatusSeeOther)
}

func (h *Handler) serverError(w http.ResponseWriter, err error) {
	http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
}
