# DigitalOcean Spaces File Upload Setup

This guide explains how to configure and use DigitalOcean Spaces for file uploads in the Trip Planner application.

## Overview

The application now supports multiple storage providers for document uploads:
- **Local Storage** (default) - Files stored on local filesystem
- **DigitalOcean Spaces** - S3-compatible object storage

## Configuration

### Environment Variables

Add the following variables to your `.env` file:

```env
# Storage Provider Configuration
# Options: local, digitalocean
STORAGE_PROVIDER=digitalocean

# DigitalOcean Spaces Configuration
DO_SPACES_ACCESS_KEY=your_access_key_here
DO_SPACES_SECRET_KEY=your_secret_key_here
DO_SPACES_REGION=nyc3
DO_SPACES_BUCKET=your-bucket-name
DO_SPACES_ENDPOINT=nyc3.digitaloceanspaces.com
```

### Setting up DigitalOcean Spaces

1. **Create a Space** in your DigitalOcean account:
   - Go to the [DigitalOcean Spaces console](https://cloud.digitalocean.com/spaces)
   - Click "Create a Space"
   - Choose a region (e.g., `nyc3`, `sfo3`, `sgp1`)
   - Name your Space (this will be your bucket name)
   - Choose File Listing permissions (recommended: Private or Public for CDN)

2. **Generate API Keys**:
   - Go to API → Spaces access keys
   - Click "Generate New Key"
   - Copy the Access Key and Secret Key
   - Add them to your `.env` file

3. **Configure CORS (Optional)**:
   If you need to access files from a web browser, configure CORS settings:
   - Go to your Space settings
   - Add CORS configuration for allowed origins

## Usage

### Switching Storage Providers

To switch between storage providers, simply change the `STORAGE_PROVIDER` environment variable:

```env
# Use local storage
STORAGE_PROVIDER=local

# Use DigitalOcean Spaces
STORAGE_PROVIDER=digitalocean
```

### Default Behavior

If `STORAGE_PROVIDER` is not set, the application defaults to **local** storage.

### File Organization

Files are stored with the following structured path format:
```
{username}/{trip_name}/{trip_id}/{category}/{file_name}
```

Example:
```
johndoe/paris-vacation/123e4567-e89b-12d3-a456-426614174000/tickets/flight-to-paris.pdf
```

**Components:**
- `username`: Sanitized username of the uploader
- `trip_name`: Sanitized name of the trip
- `trip_id`: UUID of the trip
- `category`: Document category (tickets, invoices, etc.)
- `file_name`: Sanitized file name with extension

**Benefits:**
- Easy navigation through folders
- Clear organization by user and trip
- Simple to debug and manage
- Supports backup and migration by user/trip

For detailed information, see `FILE_PATH_STRUCTURE.md`

## API Endpoints

The document upload API remains the same regardless of storage provider:

### Upload Document
```bash
POST /api/v1/trip/{trip_id}/documents
Content-Type: multipart/form-data

Parameters:
- file: The file to upload
- name: Display name
- category: Document category
- description: Optional description
- notes: Optional notes
- tags: Comma-separated tags
- is_public: Boolean (true/false)
- expires_at: Optional expiration date (RFC3339 format)
```

### Download Document
```bash
GET /api/v1/documents/{document_id}/download
```

### Delete Document
```bash
DELETE /api/v1/documents/{document_id}
```

## Storage Provider Features

### Local Storage
- Files stored in `./uploads/documents/`
- Fast for development and testing
- No external dependencies
- Limited scalability

### DigitalOcean Spaces
- S3-compatible object storage
- Scalable and reliable
- CDN integration available
- Public/private file access control
- Files automatically set to `public-read` ACL
- Direct public URLs available

## Public URLs

When using DigitalOcean Spaces, files are accessible via:
```
https://{bucket}.{endpoint}/{path}
```

Example:
```
https://my-bucket.nyc3.digitaloceanspaces.com/johndoe/paris-vacation/123e4567-e89b-12d3-a456-426614174000/tickets/flight-to-paris.pdf
```

The structured path makes it easy to navigate your Space through the DigitalOcean web interface.

## Security Considerations

1. **API Keys**: Keep your DigitalOcean API keys secure. Never commit them to version control.

2. **Access Control**:
   - Files are uploaded with `public-read` ACL by default
   - Modify `storage/digitalocean.go:77` to change ACL settings
   - Use presigned URLs for temporary private access

3. **File Validation**: The application validates:
   - File size (max 50MB)
   - Content type
   - User permissions

## Troubleshooting

### "Storage provider not available" error
- Ensure `STORAGE_PROVIDER` is correctly set in `.env`
- Verify all required environment variables are set
- Check application logs for initialization errors

### "Failed to upload file" error
- Verify DigitalOcean API keys are correct
- Check Space name and region match your configuration
- Ensure the Space has sufficient storage quota
- Verify network connectivity to DigitalOcean

### Files not accessible
- Check Space permissions
- Verify CORS settings if accessing from browser
- Ensure the endpoint URL is correct

## Migration from Local to DigitalOcean Spaces

To migrate existing local files to DigitalOcean Spaces:

1. Update `STORAGE_PROVIDER` to `digitalocean`
2. New uploads will use DigitalOcean Spaces automatically
3. Existing local files remain accessible (the system supports mixed storage)
4. To migrate old files, you'll need to manually upload them to Spaces and update the database records

## Development and Testing

For local development, use local storage:
```env
STORAGE_PROVIDER=local
```

For production, use DigitalOcean Spaces:
```env
STORAGE_PROVIDER=digitalocean
```

## Technical Architecture

The storage system uses a provider abstraction pattern:
- `storage/interfaces.go` - Storage provider interface
- `storage/config.go` - Storage manager and configuration
- `storage/local.go` - Local filesystem implementation
- `storage/digitalocean.go` - DigitalOcean Spaces implementation
- `storage/init.go` - Global initialization from environment variables

This architecture makes it easy to add additional storage providers (AWS S3, Google Cloud Storage, etc.) in the future.
