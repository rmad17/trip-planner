-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create travel_documents table to store RAG content
CREATE TABLE IF NOT EXISTS travel_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    city_name VARCHAR(255) NOT NULL,
    city_country VARCHAR(255) NOT NULL,
    document_type VARCHAR(100) NOT NULL, -- 'historical', 'logistics', 'food', 'activities', 'nearby_places'
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    embedding vector(1536), -- OpenAI ada-002 or similar dimension
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_travel_docs_city ON travel_documents(city_name);
CREATE INDEX IF NOT EXISTS idx_travel_docs_type ON travel_documents(document_type);
CREATE INDEX IF NOT EXISTS idx_travel_docs_city_type ON travel_documents(city_name, document_type);

-- Create index for vector similarity search using HNSW (Hierarchical Navigable Small World)
-- This enables fast approximate nearest neighbor search
CREATE INDEX IF NOT EXISTS idx_travel_docs_embedding ON travel_documents
USING hnsw (embedding vector_cosine_ops);

-- Alternative: IVFFlat index (comment out HNSW above and uncomment below if preferred)
-- CREATE INDEX IF NOT EXISTS idx_travel_docs_embedding ON travel_documents
-- USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Create GIN index on metadata for JSON queries
CREATE INDEX IF NOT EXISTS idx_travel_docs_metadata ON travel_documents USING gin(metadata);

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
COMMENT ON COLUMN travel_documents.embedding IS 'Vector embedding for semantic search (dimension: 1536)';
COMMENT ON COLUMN travel_documents.document_type IS 'Category: historical, logistics, food, activities, nearby_places';
COMMENT ON COLUMN travel_documents.metadata IS 'Additional structured data (tags, sources, coordinates, etc.)';
