# OAuth2 Authentication Setup Guide

This guide explains how to set up OAuth2 authentication with GitHub and Google for both the Go backend and Next.js frontend.

## Prerequisites

- GitHub account for GitHub OAuth
- Google account for Google OAuth
- Application deployed or running locally

## Backend Configuration

### Environment Variables

Add these environment variables to your `.env` file or deployment configuration:

```bash
# GitHub OAuth
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=http://localhost:8080/api/v1/auth/oauth/github/callback

# Google OAuth
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/oauth/google/callback

# OAuth Security
OAUTH_STATE_SECRET=your-random-32-byte-secret-here

# Frontend URL (for redirects after OAuth)
FRONTEND_URL=http://localhost:3000
```

For production, update the URLs accordingly:
- `GITHUB_REDIRECT_URL=https://yourdomain.com/api/v1/auth/oauth/github/callback`
- `GOOGLE_REDIRECT_URL=https://yourdomain.com/api/v1/auth/oauth/google/callback`
- `FRONTEND_URL=https://yourdomain.com`

### Database Migration

Run the OAuth migration to update your database schema:

```bash
# Apply the migration
psql -U your_user -d your_database -f migrations/002_oauth_providers.sql
```

This migration:
- Adds OAuth fields to the users table
- Creates oauth_accounts table for linking multiple providers
- Creates oauth_states table for CSRF protection
- Allows nullable passwords for OAuth-only users

## Frontend Configuration

### Environment Variables

Add to your frontend `.env.local`:

```bash
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1

# For production
# NEXT_PUBLIC_API_URL=https://yourdomain.com/api/v1
```

## Setting up OAuth Applications

### GitHub OAuth Setup

1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Click "New OAuth App"
3. Fill in the application details:
   - **Application name**: Your App Name
   - **Homepage URL**: `http://localhost:3000` (or your production URL)
   - **Authorization callback URL**: `http://localhost:8080/api/v1/auth/oauth/github/callback`
4. Click "Register application"
5. Copy the **Client ID**
6. Generate a new client secret and copy it
7. Add these to your backend environment variables

### Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API or Google Identity API
4. Go to "Credentials" → "Create Credentials" → "OAuth client ID"
5. Configure OAuth consent screen if prompted:
   - Choose "External" for public apps
   - Fill in required fields
   - Add scopes: `userinfo.email` and `userinfo.profile`
6. Create OAuth 2.0 Client ID:
   - **Application type**: Web application
   - **Name**: Your App Name
   - **Authorized JavaScript origins**: 
     - `http://localhost:3000`
     - `http://localhost:8080`
   - **Authorized redirect URIs**: 
     - `http://localhost:8080/api/v1/auth/oauth/google/callback`
7. Copy the **Client ID** and **Client Secret**
8. Add these to your backend environment variables

## Usage

### Sign In with OAuth

Once configured, users can:
1. Click "Sign in with GitHub" or "Sign in with Google" on the login page
2. Authorize the application on the provider's page
3. Get redirected back and automatically signed in

### Features

- **Automatic account linking**: If a user with the same email exists, OAuth account is linked
- **Multiple providers**: Users can link both GitHub and Google to the same account
- **Profile data**: Avatar and username are imported from OAuth providers
- **Secure**: CSRF protection with state parameter
- **Token management**: OAuth tokens are securely stored for future API calls

### API Endpoints

#### Public Endpoints
- `GET /api/v1/auth/oauth/providers` - List enabled OAuth providers
- `GET /api/v1/auth/oauth/:provider/authorize` - Initiate OAuth flow
- `GET /api/v1/auth/oauth/:provider/callback` - OAuth callback handler

#### Protected Endpoints (require authentication)
- `GET /api/v1/auth/oauth/linked` - Get user's linked OAuth accounts
- `POST /api/v1/auth/oauth/:provider/link` - Link OAuth account to existing user
- `DELETE /api/v1/auth/oauth/:provider/unlink` - Unlink OAuth account

## Security Considerations

1. **State Parameter**: Always validated to prevent CSRF attacks
2. **HTTPS Required**: Use HTTPS in production for all OAuth flows
3. **Token Storage**: OAuth tokens are encrypted and stored securely
4. **Scope Limitation**: Only request necessary scopes (email and profile)
5. **Account Unlinking**: Users can't unlink their only authentication method

## Troubleshooting

### Common Issues

1. **"Provider not enabled" error**
   - Ensure CLIENT_ID and CLIENT_SECRET are set for the provider
   - Check that environment variables are loaded correctly

2. **Redirect URI mismatch**
   - Ensure the redirect URI in OAuth app settings matches exactly
   - Include protocol (http/https) and trailing slashes if any

3. **State validation failed**
   - Check that OAUTH_STATE_SECRET is set
   - Ensure cookies are enabled in the browser
   - State expires after 10 minutes

4. **User creation failed**
   - Check database connection
   - Ensure migrations have been run
   - Check for unique constraint violations (username/email)

### Testing OAuth Locally

For local development:
1. Use `localhost` (not `127.0.0.1`) for consistency
2. Both GitHub and Google support localhost redirect URIs
3. Ensure your backend and frontend are running on the expected ports

## Production Deployment

When deploying to production:

1. Update all URLs in environment variables
2. Use strong, randomly generated OAUTH_STATE_SECRET
3. Enable HTTPS for all endpoints
4. Set secure cookie flags in production
5. Consider implementing rate limiting on OAuth endpoints
6. Monitor for suspicious OAuth activity
7. Regularly rotate OAuth client secrets

## Additional Features

The implementation supports:
- PKCE (Proof Key for Code Exchange) for enhanced security
- Account merging (linking OAuth to existing email/password accounts)
- Multiple OAuth providers per user
- OAuth token refresh (where supported by provider)
- Profile synchronization from OAuth providers