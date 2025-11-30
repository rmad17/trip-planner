# Google OAuth Configuration Guide

## Overview

This guide explains how to configure Google OAuth login for the Trip Planner application on Digital Ocean.

## Changes Made

### Code Changes

1. **User Model Extended** (`accounts/models.go:28-46`)
   - Added Google OAuth fields: `google_id`, `name`, `first_name`, `last_name`, `avatar_url`, `locale`
   - Added OAuth metadata: `provider`, `access_token`, `refresh_token`, `expires_at`
   - Made `email` field unique to prevent duplicate accounts

2. **Dynamic Callback URL** (`accounts/routers.go:20-33`)
   - Updated to use `GOOGLE_OAUTH_CALLBACK_URL` environment variable
   - Falls back to `http://localhost:8080/auth/google/callback` for local development

3. **Enhanced OAuth Callback** (`accounts/auth.go:166-249`)
   - Creates new user or updates existing user from Google data
   - Stores all Google profile information in database
   - Generates JWT token for authenticated sessions
   - Returns JSON response with token and user data

4. **Database Migration** (`migrations/20251130100619_add_google_oauth_fields.sql`)
   - Adds all new OAuth fields to users table
   - Creates indexes on `google_id` and `provider` for performance
   - Adds unique constraint on `email`

## Digital Ocean Setup

### Step 1: Get Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Enable Google+ API
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client ID"
5. Configure OAuth consent screen if needed
6. For Application type, select "Web application"
7. Add Authorized redirect URIs:
   - For production: `https://your-domain.com/auth/google/callback`
   - For development: `http://localhost:8080/auth/google/callback`
8. Save and copy the **Client ID** and **Client Secret**

### Step 2: Configure Environment Variables

Add these to your `.env` file on the Digital Ocean droplet:

```bash
# Google OAuth Configuration
GOOGLE_OAUTH_CLIENT_ID=your-actual-client-id-here
GOOGLE_OAUTH_CLIENT_SECRET=your-actual-client-secret-here

# IMPORTANT: Callback URL should point to your BACKEND API
# This is where Google redirects after authentication
GOOGLE_OAUTH_CALLBACK_URL=https://api.yourdomain.com/auth/google/callback

# Frontend URL where users will be redirected after successful login
# This is where your React/Vue/etc app is hosted
FRONTEND_URL=https://app.yourdomain.com
```

Make sure to replace:
- `your-actual-client-id-here` with your Google OAuth Client ID
- `your-actual-client-secret-here` with your Google OAuth Client Secret
- `api.yourdomain.com` with your **BACKEND API domain** (not frontend!)
- `app.yourdomain.com` with your **FRONTEND app domain**

**Example configurations:**

**Development:**
```bash
GOOGLE_OAUTH_CALLBACK_URL=http://localhost:8080/auth/google/callback  # Backend
FRONTEND_URL=http://localhost:3000  # Frontend
```

**Production:**
```bash
GOOGLE_OAUTH_CALLBACK_URL=https://api.trip-planner.com/auth/google/callback  # Backend
FRONTEND_URL=https://trip-planner.com  # Frontend
```

**Using Droplet IP (not recommended):**
```bash
GOOGLE_OAUTH_CALLBACK_URL=https://157.245.111.32/auth/google/callback  # Backend IP
FRONTEND_URL=https://trip-planner-fe.ondigitalocean.app  # Frontend domain
```

### Step 3: Update docker-compose.yml Environment

The `docker-compose.yml` has been updated to include the Google OAuth environment variables:

```yaml
environment:
  - GOOGLE_OAUTH_CLIENT_ID=${GOOGLE_OAUTH_CLIENT_ID}
  - GOOGLE_OAUTH_CLIENT_SECRET=${GOOGLE_OAUTH_CLIENT_SECRET}
  - GOOGLE_OAUTH_CALLBACK_URL=${GOOGLE_OAUTH_CALLBACK_URL}
```

### Step 4: Run Database Migration

Apply the migration to add the new fields:

```bash
# Using Atlas (recommended)
atlas schema apply \
  --url "postgres://user:pass@localhost:5432/dbname?sslmode=disable" \
  --to "file://migrations/20251130100619_add_google_oauth_fields.sql"

# Or using GORM AutoMigrate
go run cmd/migrate/main.go
```

### Step 5: Deploy to Digital Ocean

1. Push changes to your repository
2. SSH into your Digital Ocean droplet
3. Pull the latest changes
4. Update your `.env` file with the Google OAuth credentials
5. Rebuild and restart containers:

```bash
docker-compose down
docker-compose build
docker-compose up -d
```

## API Endpoints

### Google OAuth Flow

1. **Start OAuth Flow**
   ```
   GET /auth/google/begin
   ```
   Redirects user to Google login page

2. **OAuth Callback** (handled by Google, then returns to your app)
   ```
   GET /auth/google/callback
   ```

   Response:
   ```json
   {
     "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
     "user": {
       "id": "uuid-here",
       "username": "user@example.com",
       "email": "user@example.com",
       "name": "John Doe",
       "avatar_url": "https://lh3.googleusercontent.com/..."
     }
   }
   ```

### Frontend Integration

**The OAuth Flow:**

1. User clicks "Login with Google" button in your frontend
2. Frontend redirects to: `https://api.yourdomain.com/auth/google/begin`
3. Backend redirects user to Google login page
4. User authenticates with Google
5. Google redirects to: `https://api.yourdomain.com/auth/google/callback`
6. Backend processes callback and redirects to: `https://app.yourdomain.com/auth/callback?token=JWT_TOKEN`
7. Frontend receives token and stores it

**Frontend Implementation:**

**Step 1: Login Button**
```jsx
// In your React/Vue component
function GoogleLoginButton() {
  const handleGoogleLogin = () => {
    // Redirect to backend OAuth endpoint
    window.location.href = 'https://api.yourdomain.com/auth/google/begin';
  };

  return <button onClick={handleGoogleLogin}>Login with Google</button>;
}
```

**Step 2: Handle Callback**
```jsx
// Create a route: /auth/callback
// This receives the token from backend redirect

import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';

function AuthCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  useEffect(() => {
    const token = searchParams.get('token');

    if (token) {
      // Store token in localStorage
      localStorage.setItem('authToken', token);

      // Optionally fetch user profile
      // fetchUserProfile(token);

      // Redirect to dashboard or home
      navigate('/dashboard');
    } else {
      // Handle error
      navigate('/login?error=auth_failed');
    }
  }, [searchParams, navigate]);

  return <div>Completing login...</div>;
}
```

**Step 3: Use Token in API Calls**
```jsx
// API client setup
const apiClient = axios.create({
  baseURL: 'https://api.yourdomain.com/api/v1',
});

// Add token to all requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('authToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Use in components
async function fetchUserProfile() {
  const response = await apiClient.get('/user/profile');
  return response.data;
}
```

## Security Considerations

1. **HTTPS Required**: Google OAuth requires HTTPS in production
2. **Token Storage**: Access tokens and refresh tokens are stored encrypted in database
3. **Password Field**: For Google OAuth users, password field can be left empty
4. **Email Uniqueness**: Email is unique - users can't create duplicate accounts

## Testing

### Local Testing

1. Update your `.env` file with Google OAuth credentials
2. Set callback URL to `http://localhost:8080/auth/google/callback`
3. Run the application: `go run main.go`
4. Navigate to `http://localhost:8080/auth/google/begin`

### Production Testing

1. Ensure HTTPS is configured (Caddy handles this automatically)
2. Update Google OAuth console with production callback URL
3. Test the complete flow from your production domain

## Troubleshooting

### "redirect_uri_mismatch" Error
- Ensure the callback URL in Google Console exactly matches `GOOGLE_OAUTH_CALLBACK_URL`
- Check for trailing slashes or protocol mismatches (http vs https)

### "Invalid credentials" Error
- Verify `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` are correct
- Check that environment variables are properly loaded

### User Not Created
- Check database logs for constraint violations
- Verify email uniqueness constraints
- Ensure database migration was applied

## Additional Fields Available

The following data from Google is now stored in your User model:

- `google_id`: Unique Google user identifier
- `email`: User's email address
- `name`: Full name from Google profile
- `first_name`: First name
- `last_name`: Last name
- `avatar_url`: Profile picture URL
- `locale`: User's locale/language preference
- `provider`: Set to "google" for OAuth users
- `access_token`: Google access token (encrypted)
- `refresh_token`: Google refresh token (encrypted)
- `expires_at`: Token expiration timestamp

These fields can be used to personalize the user experience and access Google APIs on behalf of the user.
