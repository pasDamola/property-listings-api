package service

import (
	"context"
	"errors"

	"github.com/pasDamola/property-listings-api/internal/domain"
	"github.com/pasDamola/property-listings-api/internal/repository"
)

var (
	ErrNotFound = errors.New("listing not found")
)

type ListingService interface {
	Create(ctx context.Context, req *domain.CreateListingRequest) (*domain.Listing, error)
	GetByID(ctx context.Context, id string) (*domain.Listing, error)
	Update(ctx context.Context, id string, req *domain.CreateListingRequest) (*domain.Listing, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, filters domain.SearchFilters) (*domain.PaginatedResponse, error)
}

type listingService struct {
	repo repository.ListingRepository
}

func NewListingService(repo repository.ListingRepository) ListingService {
	return &listingService{repo: repo}
}

func (s *listingService) Create(ctx context.Context, req *domain.CreateListingRequest) (*domain.Listing, error) {
	return s.repo.Create(ctx, req)
}

func (s *listingService) GetByID(ctx context.Context, id string) (*domain.Listing, error) {
	listing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if listing == nil {
		return nil, ErrNotFound
	}
	return listing, nil
}

func (s *listingService) Update(ctx context.Context, id string, req *domain.CreateListingRequest) (*domain.Listing, error) {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, req)
}

func (s *listingService) Delete(ctx context.Context, id string) error {
	_, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *listingService) Search(ctx context.Context, filters domain.SearchFilters) (*domain.PaginatedResponse, error) {
	if filters.Page <= 0 { filters.Page = 1 }
	if filters.Limit <= 0 || filters.Limit > 100 { filters.Limit = 10 }

	listings, total, err := s.repo.Search(ctx, filters)
	if err != nil {
		return nil, err
	}

	totalPages := (total + filters.Limit - 1) / filters.Limit

	return &domain.PaginatedResponse{
		Data: listings,
		Meta: domain.Meta{
			Total:      total,
			Page:       filters.Page,
			Limit:      filters.Limit,
			TotalPages: totalPages,
		},
	}, nil
}