-- Migration to update vector dimensions from 1536 (OpenAI) to 768 (Ollama nomic-embed-text)
-- This migration is necessary when switching from OpenAI/Claude to Ollama as the default provider
-- Run this migration if you're using Ollama for embeddings

-- Drop existing table and recreate with 768 dimensions
-- WARNING: This will delete all existing travel documents
DROP TABLE IF EXISTS travel_documents CASCADE;

-- Drop the view that depends on the table
DROP VIEW IF EXISTS v_city_travel_info;

-- Enable pgvector extension (if not already enabled)
CREATE EXTENSION IF NOT EXISTS vector;

-- Create travel_documents table with 768-dimension vectors (Ollama default)
CREATE TABLE travel_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    city_name VARCHAR(255) NOT NULL,
    city_country VARCHAR(255) NOT NULL,
    document_type VARCHAR(100) NOT NULL, -- 'historical', 'logistics', 'food', 'activities', 'nearby_places'
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    embedding vector(768), -- Ollama nomic-embed-text dimension (was 1536 for OpenAI)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_travel_docs_city ON travel_documents(city_name);
CREATE INDEX idx_travel_docs_type ON travel_documents(document_type);
CREATE INDEX idx_travel_docs_city_type ON travel_documents(city_name, document_type);

-- Create index for vector similarity search using HNSW (Hierarchical Navigable Small World)
-- This enables fast approximate nearest neighbor search
CREATE INDEX idx_travel_docs_embedding ON travel_documents
USING hnsw (embedding vector_cosine_ops);

-- Alternative: IVFFlat index (comment out HNSW above and uncomment below if preferred)
-- CREATE INDEX idx_travel_docs_embedding ON travel_documents
-- USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Create GIN index on metadata for JSON queries
CREATE INDEX idx_travel_docs_metadata ON travel_documents USING gin(metadata);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_travel_documents_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update updated_at
CREATE TRIGGER trigger_travel_documents_updated_at
    BEFORE UPDATE ON travel_documents
    FOR EACH ROW
    EXECUTE FUNCTION update_travel_documents_updated_at();

-- Create view for easy content retrieval by city
CREATE OR REPLACE VIEW v_city_travel_info AS
SELECT
    city_name,
    city_country,
    document_type,
    array_agg(
        json_build_object(
            'id', id,
            'title', title,
            'content', content,
            'metadata', metadata,
            'created_at', created_at
        ) ORDER BY created_at DESC
    ) as documents
FROM travel_documents
GROUP BY city_name, city_country, document_type;

-- Add comments for documentation
COMMENT ON TABLE travel_documents IS 'Stores travel information for RAG-based trip planning assistance';
COMMENT ON COLUMN travel_documents.embedding IS 'Vector embedding for semantic search using Ollama (dimension: 768)';
COMMENT ON COLUMN travel_documents.document_type IS 'Category: historical, logistics, food, activities, nearby_places';
COMMENT ON COLUMN travel_documents.metadata IS 'Additional structured data (tags, sources, coordinates, etc.)';

-- Migration info
COMMENT ON TABLE travel_documents IS 'Travel documents table with 768-dim embeddings for Ollama. Updated: 2026-01-04';
