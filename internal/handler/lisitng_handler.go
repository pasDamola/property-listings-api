package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pasDamola/property-listings-api/internal/domain"
	"github.com/pasDamola/property-listings-api/internal/service"
	"github.com/pasDamola/property-listings-api/internal/validator"
)

type ListingHandler struct {
	svc service.ListingService
}

func NewListingHandler(svc service.ListingService) *ListingHandler {
	return &ListingHandler{svc: svc}
}

func (h *ListingHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/search", h.Search)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *ListingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if err := validator.Validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "Validation failed",
			"fields": validator.FormatValidationErrors(err),
		})
		return
	}

	listing, err := h.svc.Create(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create listing")
		return
	}
	respondJSON(w, http.StatusCreated, listing)
}

func (h *ListingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	listing, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Listing not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to retrieve listing")
		return
	}
	respondJSON(w, http.StatusOK, listing)
}

func (h *ListingHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req domain.CreateListingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if err := validator.Validate.Struct(req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "Validation failed",
			"fields": validator.FormatValidationErrors(err),
		})
		return
	}

	listing, err := h.svc.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Listing not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update listing")
		return
	}
	respondJSON(w, http.StatusOK, listing)
}

func (h *ListingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "Listing not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete listing")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ListingHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	
	filters := domain.SearchFilters{
		Type:     domain.ListingType(q.Get("type")),
		Page:     parseInt(q.Get("page"), 1),
		Limit:    parseInt(q.Get("limit"), 10),
		Bedrooms: parseInt(q.Get("bedrooms"), 0),
		MinPrice: parseFloat(q.Get("minPrice"), 0),
		MaxPrice: parseFloat(q.Get("maxPrice"), 0),
		Lat:      parseFloat(q.Get("lat"), 0),
		Lng:      parseFloat(q.Get("lng"), 0),
		Radius:   parseFloat(q.Get("radius"), 10), // we set default 10km
	}

	res, err := h.svc.Search(r.Context(), filters)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to search listings")
		return
	}
	respondJSON(w, http.StatusOK, res)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func parseInt(s string, fallback int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return fallback
}

func parseFloat(s string, fallback float64) float64 {
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return fallback
}