# Pixtify API

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue?style=flat)

Pixtify is a RESTful API for a wallpaper sharing platform. Built with Go and the Fiber framework, it provides a complete backend solution for managing wallpapers, user accounts, collections, and content moderation.

## Overview

- **Total Endpoints:** 46
- **Framework:** Go Fiber v2
- **Database:** PostgreSQL 16
- **Storage:** Cloudflare R2 (S3-compatible)
- **Authentication:** JWT + OAuth 2.0

---

## Features

### User Management
- User registration and login with email/password
- OAuth authentication via GitHub and Google
- JWT-based session management with access and refresh tokens
- Role-based access control (User, Moderator, Owner)
- Account management (update profile, delete account)

### Wallpaper System
- Upload wallpapers with automatic thumbnail generation
- Download original and thumbnail versions
- Like/unlike functionality
- Featured wallpapers curation
- Search by title and description
- Filter by tags
- Trending wallpapers based on recent activity

### Collections
- Create custom collections
- Add and remove wallpapers from collections
- View collection contents

### Tags
- Tag-based categorization
- Filter wallpapers by tag
- Tag management for moderators

### Content Moderation
- User reporting system
- Report review and resolution
- User ban/unban capabilities
- Moderator and Owner role privileges

### System
- Health check endpoint with system statistics
- Rate limiting for sensitive endpoints
- CORS configuration

---

## Endpoint Summary

| Category | Count | Description |
|----------|-------|-------------|
| Authentication | 10 | Login, register, OAuth, token management |
| Users | 9 | Profile, account management, admin actions |
| Wallpapers | 12 | CRUD, search, trending, likes, featured |
| Collections | 7 | Create, manage, add/remove wallpapers |
| Tags | 3 | List, create, delete |
| Reports | 4 | Create, list, review, resolve |
| Health | 1 | System status and metrics |
| **Total** | **46** | |

---

## Endpoint List

### Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/users/register` | No | Register new user |
| POST | `/api/users/login` | No | Login with email/password |
| GET | `/api/auth/github` | No | Initiate GitHub OAuth |
| GET | `/api/auth/google` | No | Initiate Google OAuth |
| GET | `/api/auth/github/callback` | No | GitHub OAuth callback |
| GET | `/api/auth/google/callback` | No | Google OAuth callback |
| POST | `/api/auth/refresh` | No | Refresh access token |
| POST | `/api/auth/logout` | No | Logout current session |
| GET | `/api/auth/profile` | Yes | Get authenticated user profile |
| POST | `/api/auth/logout-all` | Yes | Logout from all devices |

### Users

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/users` | Yes | List all users |
| GET | `/api/users/me` | Yes | Get current user |
| PUT | `/api/users/me` | Yes | Update current user |
| DELETE | `/api/users/me` | Yes | Delete account |
| GET | `/api/users/:id` | Yes | Get user by ID |
| GET | `/api/users/:id/wallpapers` | No | Get user wallpapers |
| POST | `/api/users/:id/ban` | Mod | Ban user |
| DELETE | `/api/users/:id/ban` | Mod | Unban user |
| GET | `/api/users/:id/stats` | Mod | Get user statistics |

### Wallpapers

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/wallpapers` | No | List wallpapers |
| GET | `/api/wallpapers?tag=:slug` | No | Filter by tag |
| GET | `/api/wallpapers/featured` | No | List featured wallpapers |
| GET | `/api/wallpapers/search?q=:query` | No | Search wallpapers |
| GET | `/api/wallpapers/trending` | No | Get trending wallpapers |
| GET | `/api/wallpapers/:id` | No | Get wallpaper by ID |
| POST | `/api/wallpapers` | Yes | Upload wallpaper |
| PUT | `/api/wallpapers/:id` | Yes | Update wallpaper |
| DELETE | `/api/wallpapers/:id` | Yes | Delete wallpaper |
| POST | `/api/wallpapers/:id/like` | Yes | Toggle like |
| GET | `/api/users/me/liked-wallpapers` | Yes | Get liked wallpapers |
| POST | `/api/wallpapers/:id/featured` | Mod | Set featured status |

### Collections

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/collections` | Yes | Create collection |
| GET | `/api/collections/me` | Yes | Get my collections |
| GET | `/api/collections/:id` | Yes | Get collection by ID |
| DELETE | `/api/collections/:id` | Yes | Delete collection |
| GET | `/api/collections/:id/wallpapers` | Yes | Get collection wallpapers |
| POST | `/api/collections/:id/wallpapers` | Yes | Add wallpaper |
| DELETE | `/api/collections/:id/wallpapers/:wid` | Yes | Remove wallpaper |

### Tags

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/tags` | No | List all tags |
| POST | `/api/tags` | Mod | Create tag |
| DELETE | `/api/tags/:id` | Mod | Delete tag |

### Reports

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/reports` | Yes | Create report |
| GET | `/api/reports` | Mod | List all reports |
| GET | `/api/reports/:id` | Mod | Get report by ID |
| PUT | `/api/reports/:id` | Mod | Update report status |

### Health

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | No | Health check with system stats |

**Access:** No = Public, Yes = Requires login, Mod = Moderator or Owner only

---

## Authentication

The API uses JWT tokens for authentication:
- **Access Token:** Short-lived (15 minutes), used for API requests
- **Refresh Token:** Long-lived (7 days), used to obtain new access tokens

OAuth 2.0 is supported for GitHub and Google providers.

---

## Role Permissions

| Role | Permissions |
|------|-------------|
| User | Upload, like, create collections, report content |
| Moderator | All user permissions + manage tags, review reports, set featured, ban users |
| Owner | All moderator permissions + full administrative access |

---

## Rate Limits

| Endpoint | Limit |
|----------|-------|
| Login | 5 requests per minute |
| Register | 3 requests per 5 minutes |
| OAuth | 10 requests per minute |
| General API | 100 requests per minute |

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.25 |
| Web Framework | Fiber v2 |
| Database | PostgreSQL 16 |
| Object Storage | Cloudflare R2 |
| Image Processing | disintegration/imaging |
| Authentication | golang-jwt/jwt |

---

## Project Structure

```
pixtify/
├── cmd/api/           # Application entrypoint
├── internal/
│   ├── config/        # Configuration management
│   ├── handler/       # HTTP request handlers
│   ├── middleware/    # JWT, rate limiting, CORS
│   ├── repository/    # Database access layer
│   ├── service/       # Business logic
│   ├── storage/       # Object storage (R2/S3)
│   └── processor/     # Image processing
├── database/
│   └── migrations/    # SQL migration files
└── Makefile           # Build and development commands
```

---

## License

MIT License
