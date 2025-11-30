# Google OAuth URL Configuration - Quick Guide

## TL;DR

**GOOGLE_OAUTH_CALLBACK_URL** = Your **BACKEND** domain/IP
**FRONTEND_URL** = Your **FRONTEND** app domain

---

## The Flow

```
┌─────────────┐
│  Frontend   │ User clicks "Login with Google"
│   (React)   │────────────────────────────────┐
└─────────────┘                                │
                                               ▼
                                      ┌─────────────────┐
                                      │    Backend      │
                                      │  /auth/google/  │
                                      │     begin       │
                                      └────────┬────────┘
                                               │ Redirects user
                                               ▼
                                      ┌─────────────────┐
                                      │  Google OAuth   │
                                      │   Login Page    │
                                      └────────┬────────┘
                                               │ User logs in
                                               ▼
                                      ┌─────────────────┐
                Google redirects to → │    Backend      │
        GOOGLE_OAUTH_CALLBACK_URL     │  /auth/google/  │
                (BACKEND!)            │    callback     │
                                      └────────┬────────┘
                                               │ Generates JWT
                                               ▼
┌─────────────┐                       ┌─────────────────┐
│  Frontend   │ ←──────────────────── │  Redirect with  │
│ /auth/      │    FRONTEND_URL       │  token in URL   │
│  callback   │                       └─────────────────┘
└─────────────┘
     │
     │ Stores token
     ▼
  Dashboard
```

---

## Environment Variable Examples

### Development (Local)

```bash
# Backend runs on localhost:8080
GOOGLE_OAUTH_CALLBACK_URL=http://localhost:8080/auth/google/callback

# Frontend runs on localhost:3000
FRONTEND_URL=http://localhost:3000
```

### Production (Separate Domains)

```bash
# Backend on subdomain
GOOGLE_OAUTH_CALLBACK_URL=https://api.trip-planner.com/auth/google/callback

# Frontend on main domain
FRONTEND_URL=https://trip-planner.com
```

### Production (Using Digital Ocean URLs)

Based on your current setup:

```bash
# Backend on Droplet (behind Caddy)
GOOGLE_OAUTH_CALLBACK_URL=https://your-caddy-domain.com/auth/google/callback

# Frontend on Digital Ocean App Platform
FRONTEND_URL=https://trip-planner-fe-mlmam.ondigitalocean.app
```

### Production (Using Droplet IP - Not Recommended)

```bash
# Backend using droplet IP
GOOGLE_OAUTH_CALLBACK_URL=https://157.245.111.32/auth/google/callback

# Frontend on App Platform
FRONTEND_URL=https://trip-planner-fe-mlmam.ondigitalocean.app
```

**⚠️ Important:** For production, use domain names instead of IPs. SSL certificates work better with domains.

---

## Google Cloud Console Configuration

When setting up OAuth credentials in Google Cloud Console:

**Authorized redirect URIs** should include:

```
# Development
http://localhost:8080/auth/google/callback

# Production (use your actual backend domain)
https://api.trip-planner.com/auth/google/callback
```

**DO NOT add your frontend URL here!** Google redirects to the backend, not the frontend.

---

## What Goes Where?

| Variable | Points To | Example | Why? |
|----------|-----------|---------|------|
| `GOOGLE_OAUTH_CALLBACK_URL` | **Backend API** | `https://api.example.com/auth/google/callback` | Google needs to send auth code to your API |
| `FRONTEND_URL` | **Frontend App** | `https://app.example.com` | Where to redirect users after login |

---

## Common Mistakes ❌

### ❌ Wrong: Callback pointing to frontend
```bash
GOOGLE_OAUTH_CALLBACK_URL=https://trip-planner-fe.ondigitalocean.app/auth/callback  # WRONG!
```

### ✅ Correct: Callback pointing to backend
```bash
GOOGLE_OAUTH_CALLBACK_URL=https://your-droplet-domain.com/auth/google/callback  # RIGHT!
```

---

## Testing

### Test Locally

1. Set up `.env`:
   ```bash
   GOOGLE_OAUTH_CALLBACK_URL=http://localhost:8080/auth/google/callback
   FRONTEND_URL=http://localhost:3000
   ```

2. In Google Console, add: `http://localhost:8080/auth/google/callback`

3. Start backend: `go run app.go`

4. Visit in browser: `http://localhost:8080/auth/google/begin`

5. After login, you should be redirected to: `http://localhost:3000/auth/callback?token=...`

---

## Your Specific Setup

Based on your `docker-compose.yml`, you have these frontends in ALLOWED_ORIGINS:

```
https://trip-planner-fe.blr1.digitaloceanspaces.com
https://trip-planner-fe.blr1.cdn.digitaloceanspaces.com
https://trip-planner-fe-mlmam.ondigitalocean.app
```

So your production config should be:

```bash
# Backend (your Caddy domain or droplet)
GOOGLE_OAUTH_CALLBACK_URL=https://YOUR_BACKEND_DOMAIN/auth/google/callback

# Frontend (one of your frontend URLs)
FRONTEND_URL=https://trip-planner-fe-mlmam.ondigitalocean.app
```

---

## Quick Checklist

- [ ] `GOOGLE_OAUTH_CALLBACK_URL` points to **BACKEND**
- [ ] `FRONTEND_URL` points to **FRONTEND**
- [ ] Google Console redirect URI matches `GOOGLE_OAUTH_CALLBACK_URL`
- [ ] Both URLs use `https://` in production
- [ ] Frontend has route `/auth/callback` to handle token
- [ ] CORS allows frontend domain (already in your `ALLOWED_ORIGINS`)

---

## Need Help?

See full documentation: `GOOGLE_OAUTH_SETUP.md`
