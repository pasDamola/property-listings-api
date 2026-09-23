CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS listings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    price DECIMAL(12, 2) NOT NULL,
    type VARCHAR(50) NOT NULL,
    bedrooms INT NOT NULL,
    geom GEOGRAPHY(POINT, 4326) NOT NULL, -- Stores lat/lng efficiently
    agent_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for standard filtering
CREATE INDEX idx_listings_type_price_bedrooms ON listings(type, price, bedrooms);
-- Geospatial index (crucial for radius searches)
CREATE INDEX idx_listings_geom ON listings USING GIST (geom);