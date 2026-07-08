package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cybertortuga/aitriage/internal/models"
	"github.com/cybertortuga/aitriage/internal/server/middleware"
	"github.com/cybertortuga/aitriage/internal/server/repositories"
	"github.com/cybertortuga/aitriage/internal/server/utils"
)

type ProductHandler struct {
	repo *repositories.ProductRepository
}

func NewProductHandler(repo *repositories.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

// --- Product Types ---

func (h *ProductHandler) HandleListProductTypes(w http.ResponseWriter, r *http.Request) {
	pts, err := h.repo.ListProductTypes(r.Context())
	if err != nil {
		utils.JSONError(w, "Failed to list product types", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(pts)
}

func (h *ProductHandler) HandleCreateProductType(w http.ResponseWriter, r *http.Request) {
	var req models.ProductType
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		utils.JSONError(w, "Name is required", http.StatusBadRequest)
		return
	}

	id, err := h.repo.CreateProductType(r.Context(), &req)
	if err != nil {
		utils.JSONError(w, "Failed to create product type", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "id": id})
}

func (h *ProductHandler) HandleUpdateProductType(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var req models.ProductType
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}
	req.ID = id

	err := h.repo.UpdateProductType(r.Context(), &req)
	if err != nil {
		utils.JSONError(w, "Failed to update product type", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (h *ProductHandler) HandleDeleteProductType(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	err := h.repo.DeleteProductType(r.Context(), id)
	if err != nil {
		utils.JSONError(w, "Failed to delete product type", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// --- Products ---

func (h *ProductHandler) HandleListProducts(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.ExtractClaims(r)
	if err != nil {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	products, err := h.repo.List(r.Context(), claims.UserID, claims.GlobalRole)
	if err != nil {
		utils.JSONError(w, "Failed to list products", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	claims, err := middleware.ExtractClaims(r)
	if err != nil {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.Product
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	req.CreatedBy = &claims.UserID

	id, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		utils.JSONError(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	// Make the creator an owner of the product
	_ = h.repo.AddMember(r.Context(), id, claims.UserID, "owner")

	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "id": id})
}

func (h *ProductHandler) HandleGetProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	product, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		utils.JSONError(w, "Product not found", http.StatusNotFound)
		return
	}
	_ = json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) HandleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var req models.Product
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}
	req.ID = id

	err := h.repo.Update(r.Context(), &req)
	if err != nil {
		utils.JSONError(w, "Failed to update product", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (h *ProductHandler) HandleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	err := h.repo.Delete(r.Context(), id)
	if err != nil {
		utils.JSONError(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (h *ProductHandler) HandleUpdateProductSLA(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	var req struct {
		SLACritical int `json:"sla_critical"`
		SLAHigh     int `json:"sla_high"`
		SLAMedium   int `json:"sla_medium"`
		SLALow      int `json:"sla_low"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	p, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		utils.JSONError(w, "Product not found", http.StatusNotFound)
		return
	}

	p.SLACritical = req.SLACritical
	p.SLAHigh = req.SLAHigh
	p.SLAMedium = req.SLAMedium
	p.SLALow = req.SLALow

	err = h.repo.Update(r.Context(), p)
	if err != nil {
		utils.JSONError(w, "Failed to update product SLA", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (h *ProductHandler) HandleGetProductMembers(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	members, err := h.repo.GetMembers(r.Context(), id)
	if err != nil {
		utils.JSONError(w, "Failed to get members", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(members)
}

func (h *ProductHandler) HandleAddProductMember(w http.ResponseWriter, r *http.Request) {
	productIDStr := r.URL.Query().Get("id")
	productID, _ := strconv.ParseInt(productIDStr, 10, 64)

	var req struct {
		UserID int64  `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid data", http.StatusBadRequest)
		return
	}

	err := h.repo.AddMember(r.Context(), productID, req.UserID, req.Role)
	if err != nil {
		utils.JSONError(w, "Failed to add member", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (h *ProductHandler) HandleRemoveProductMember(w http.ResponseWriter, r *http.Request) {
	productIDStr := r.URL.Query().Get("id")
	productID, _ := strconv.ParseInt(productIDStr, 10, 64)

	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	err := h.repo.RemoveMember(r.Context(), productID, userID)
	if err != nil {
		utils.JSONError(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
