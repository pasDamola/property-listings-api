package domain

import (
	"time"
)

type ListingType string

const (
	TypeRent     ListingType = "rent"
	TypeSale     ListingType = "sale"
	TypeShortlet ListingType = "shortlet"
)

type Listing struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Price     float64     `json:"price"`
	Type      ListingType `json:"type"`
	Bedrooms  int         `json:"bedrooms"`
	Latitude  float64     `json:"latitude"`
	Longitude float64     `json:"longitude"`
	AgentID   string      `json:"agent_id"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type CreateListingRequest struct {
	Title     string      `json:"title" validate:"required,min=3,max=100"`
	Price     float64     `json:"price" validate:"required,gt=0"`
	Type      ListingType `json:"type" validate:"required,oneof=rent sale shortlet"`
	Bedrooms  int         `json:"bedrooms" validate:"required,gte=0"`
	Latitude  float64     `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64     `json:"longitude" validate:"required,min=-180,max=180"`
	AgentID   string      `json:"agent_id" validate:"required,uuid"`
}

type SearchFilters struct {
	Type     ListingType
	MinPrice float64
	MaxPrice float64
	Bedrooms int
	Lat      float64
	Lng      float64
	Radius   float64
	Page     int
	Limit    int
}

type PaginatedResponse struct {
	Data []Listing `json:"data"`
	Meta Meta      `json:"meta"`
}

type Meta struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"total_pages"`
}