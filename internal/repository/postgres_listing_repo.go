package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pasDamola/property-listings-api/internal/domain"
)

type ListingRepository interface {
	Create(ctx context.Context, req *domain.CreateListingRequest) (*domain.Listing, error)
	GetByID(ctx context.Context, id string) (*domain.Listing, error)
	Update(ctx context.Context, id string, req *domain.CreateListingRequest) (*domain.Listing, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, filters domain.SearchFilters) ([]domain.Listing, int, error)
}

type postgresListingRepo struct {
	db *pgxpool.Pool
}

func NewPostgresListingRepo(db *pgxpool.Pool) ListingRepository {
	return &postgresListingRepo{db: db}
}

func (r *postgresListingRepo) Create(ctx context.Context, req *domain.CreateListingRequest) (*domain.Listing, error) {
	query := `
		INSERT INTO listings (title, price, type, bedrooms, geom, agent_id)
		VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography, $7)
		RETURNING id, title, price, type, bedrooms, ST_Y(geom::geometry) as latitude, ST_X(geom::geometry) as longitude, agent_id, created_at, updated_at
	`
	var l domain.Listing
	err := r.db.QueryRow(ctx, query, req.Title, req.Price, req.Type, req.Bedrooms, req.Longitude, req.Latitude, req.AgentID).
		Scan(&l.ID, &l.Title, &l.Price, &l.Type, &l.Bedrooms, &l.Latitude, &l.Longitude, &l.AgentID, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		fmt.Printf("DB ERROR: %v\n", err) 
		return nil, fmt.Errorf("failed to create listing: %w", err)
	}
	return &l, nil
}

func (r *postgresListingRepo) GetByID(ctx context.Context, id string) (*domain.Listing, error) {
	query := `
		SELECT id, title, price, type, bedrooms, ST_Y(geom::geometry), ST_X(geom::geometry), agent_id, created_at, updated_at
		FROM listings WHERE id = $1
	`
	var l domain.Listing
	err := r.db.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.Title, &l.Price, &l.Type, &l.Bedrooms, &l.Latitude, &l.Longitude, &l.AgentID, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil 
		}
		return nil, fmt.Errorf("failed to get listing: %w", err)
	}
	return &l, nil
}

func (r *postgresListingRepo) Update(ctx context.Context, id string, req *domain.CreateListingRequest) (*domain.Listing, error) {
	query := `
		UPDATE listings 
		SET title=$1, price=$2, type=$3, bedrooms=$4, geom=ST_SetSRID(ST_MakePoint($5, $6), 4326)::geography, updated_at=NOW()
		WHERE id=$7
		RETURNING id, title, price, type, bedrooms, ST_Y(geom::geometry), ST_X(geom::geometry), agent_id, created_at, updated_at
	`
	var l domain.Listing
	err := r.db.QueryRow(ctx, query, req.Title, req.Price, req.Type, req.Bedrooms, req.Longitude, req.Latitude, id).
		Scan(&l.ID, &l.Title, &l.Price, &l.Type, &l.Bedrooms, &l.Latitude, &l.Longitude, &l.AgentID, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update listing: %w", err)
	}
	return &l, nil
}

func (r *postgresListingRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM listings WHERE id = $1", id)
	return err
}

func (r *postgresListingRepo) Search(ctx context.Context, filters domain.SearchFilters) ([]domain.Listing, int, error) {
	// Build dynamic query safely using pgx argument placeholders
	var queryBuilder strings.Builder
	var countBuilder strings.Builder
	var args []interface{}
	argID := 1

	baseQuery := `
		FROM listings 
		WHERE 1=1
	`
	
	// Append filters
	if filters.Type != "" {
		baseQuery += fmt.Sprintf(" AND type = $%d", argID)
		args = append(args, filters.Type)
		argID++
	}
	if filters.MinPrice > 0 {
		baseQuery += fmt.Sprintf(" AND price >= $%d", argID)
		args = append(args, filters.MinPrice)
		argID++
	}
	if filters.MaxPrice > 0 {
		baseQuery += fmt.Sprintf(" AND price <= $%d", argID)
		args = append(args, filters.MaxPrice)
		argID++
	}
	if filters.Bedrooms > 0 {
		baseQuery += fmt.Sprintf(" AND bedrooms = $%d", argID)
		args = append(args, filters.Bedrooms)
		argID++
	}
	
	// Geospatial filter (radius in meters)
	if filters.Lat != 0 && filters.Lng != 0 && filters.Radius > 0 {
		baseQuery += fmt.Sprintf(" AND ST_DWithin(geom, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::geography, $%d)", argID, argID+1, argID+2)
		args = append(args, filters.Lng, filters.Lat, filters.Radius*1000) // Convert km to meters
		argID += 3
	}

	// Count Query
	countBuilder.WriteString("SELECT COUNT(*) ")
	countBuilder.WriteString(baseQuery)
	
	var total int
	err := r.db.QueryRow(ctx, countBuilder.String(), args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count listings: %w", err)
	}

	// Data Query
	queryBuilder.WriteString("SELECT id, title, price, type, bedrooms, ST_Y(geom::geometry), ST_X(geom::geometry), agent_id, created_at, updated_at ")
	queryBuilder.WriteString(baseQuery)
	
	// Add sorting and pagination
	queryBuilder.WriteString(" ORDER BY created_at DESC")
	offset := (filters.Page - 1) * filters.Limit
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1))
	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search listings: %w", err)
	}
	defer rows.Close()

	var listings []domain.Listing
	for rows.Next() {
		var l domain.Listing
		if err := rows.Scan(&l.ID, &l.Title, &l.Price, &l.Type, &l.Bedrooms, &l.Latitude, &l.Longitude, &l.AgentID, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan listing: %w", err)
		}
		listings = append(listings, l)
	}

	return listings, total, nil
}