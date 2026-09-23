package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/pasDamola/property-listings-api/internal/domain"
	"github.com/pasDamola/property-listings-api/internal/handler"
	"github.com/pasDamola/property-listings-api/internal/service"
)

// Mock Service
type MockListingService struct {
	mock.Mock
}

func (m *MockListingService) Create(ctx context.Context, req *domain.CreateListingRequest) (*domain.Listing, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Listing), args.Error(1)
}

func (m *MockListingService) GetByID(ctx context.Context, id string) (*domain.Listing, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Listing), args.Error(1)
}

func (m *MockListingService) Update(ctx context.Context, id string, req *domain.CreateListingRequest) (*domain.Listing, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Listing), args.Error(1)
}

func (m *MockListingService) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockListingService) Search(ctx context.Context, filters domain.SearchFilters) (*domain.PaginatedResponse, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PaginatedResponse), args.Error(1)
}

func TestCreateListing_Success(t *testing.T) {
	mockSvc := new(MockListingService)
	h := handler.NewListingHandler(mockSvc)

	reqBody := domain.CreateListingRequest{
		Title: "Test Apartment", Price: 1500, Type: "rent", Bedrooms: 2,
		Latitude: 6.5244, Longitude: 3.3792, AgentID: "123e4567-e89b-12d3-a456-426614174000",
	}
	
	mockSvc.On("Create", mock.Anything, mock.Anything).Return(&domain.Listing{ID: "123", Title: "Test Apartment"}, nil)

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	mockSvc := new(MockListingService)
	h := handler.NewListingHandler(mockSvc)

	mockSvc.On("GetByID", mock.Anything, "123").Return(nil, service.ErrNotFound)

	req, _ := http.NewRequest(http.MethodGet, "/123", nil)
	rr := httptest.NewRecorder()
	
	// Inject chi URL params for testing
	ctx := chi.NewRouteContext()
	ctx.URLParams.Add("id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, ctx))

	h.GetByID(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockSvc.AssertExpectations(t)
}