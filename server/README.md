# Magic Stream Movies Server

A RESTful API server for managing a movie streaming catalog, built with **Go**, **Gin**, and **MongoDB**. Features include user authentication with JWT, AI-powered movie review sentiment analysis via OpenAI, and personalized movie recommendations based on user genre preferences.

---

## Table of Contents

- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Environment Variables](#environment-variables)
  - [Installation](#installation)
  - [Running the Server](#running-the-server)
- [API Reference](#api-reference)
  - [Public Endpoints](#public-endpoints)
  - [Protected Endpoints](#protected-endpoints)
- [Data Models](#data-models)
  - [User](#user)
  - [Movie](#movie)
  - [Genre](#genre)
  - [Ranking](#ranking)
- [Authentication](#authentication)
- [Architecture](#architecture)
  - [Middleware](#middleware)
  - [Token Management](#token-management)
  - [AI-Powered Review Ranking](#ai-powered-review-ranking)
  - [Personalized Recommendations](#personalized-recommendations)
- [MongoDB Collections](#mongodb-collections)

---

## Tech Stack

| Technology | Purpose |
|---|---|
| [Go 1.25](https://go.dev/) | Server language |
| [Gin](https://github.com/gin-gonic/gin) | HTTP web framework |
| [MongoDB (v2 Driver)](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2) | Database |
| [golang-jwt/jwt](https://github.com/golang-jwt/jwt) | JWT authentication |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Password hashing |
| [go-playground/validator](https://github.com/go-playground/validator) | Request validation |
| [langchaingo](https://github.com/tmc/langchaingo) | OpenAI LLM integration |
| [godotenv](https://github.com/joho/godotenv) | Environment variable loading |

---

## Project Structure

```
server/
├── main.go                          # Application entry point, DB client init
├── .env                             # Environment configuration (git-ignored)
├── .gitignore                       # Git ignore rules
├── go.mod                           # Go module definition
├── go.sum                           # Dependency checksums
├── controllers/
│   ├── movie_controller.go          # Movie CRUD, review ranking, recommendations
│   └── user_controller.go           # User registration & login
├── database/
│   └── database_connection.go       # MongoDB connection & collection helpers
├── middlewares/
│   └── auth_middleware.go           # JWT authentication middleware
├── models/
│   ├── movie_model.go               # Movie, Genre, Ranking structs
│   └── user_model.go                # User, UserLogin, UserResponse structs
├── routes/
│   ├── protected_routes.go          # Routes requiring authentication
│   └── unprotected_routes.go        # Public routes (no auth required)
└── utils/
    └── tokens_util.go               # JWT generation, validation & context helpers
```

---

## Getting Started

### Prerequisites

- **Go** 1.25+
- **MongoDB** running locally (default: `mongodb://localhost:27017/`)
- **OpenAI API Key** (required for AI-powered review sentiment analysis)

### Environment Variables

Create a `.env` file in the project root with the following variables:

```env
# Database
DATABASE_NAME=magic-stream-movies
MONGODB_URI=mongodb://localhost:27017/

# JWT Secrets
SECRET_KEY=<your-jwt-secret-key>
SECRET_REFRESH_KEY=<your-jwt-refresh-secret-key>

# OpenAI
OPENAI_API_KEY=<your-openai-api-key>
BASE_PROMPT_TEMPLATE=Return a response using one of these words: {rankings}. The response should be a single word and should not contain any other text. The response should be based on the following review:

# Recommendations
RECOMMENDED_MOVIE_LIMIT=5
```

> **Note:** Use strong, unique values for `SECRET_KEY` and `SECRET_REFRESH_KEY`. You can generate them with `openssl rand -hex 32`.

### Installation

```bash
# Clone the repository
git clone https://github.com/PaleBlueDot1990/magic-stream-movies.git
cd magic-stream-movies/server

# Download dependencies
go mod download
```

### Running the Server

```bash
go run .
```

The server starts on **port 8080** by default. Verify with:

```bash
curl http://localhost:8080/hello
# → Hello, Magic Stream Movies!
```

---

## API Reference

### Public Endpoints

These endpoints do **not** require authentication.

#### Health Check

```
GET /hello
```

Returns a simple greeting to verify the server is running.

**Response:** `200 OK`
```
Hello, Magic Stream Movies!
```

---

#### List All Movies

```
GET /movies
```

Returns all movies in the catalog.

**Response:** `200 OK`
```json
[
  {
    "_id": "...",
    "imdb_id": "tt0105695",
    "title": "Unforgiven",
    "poster_path": "https://image.tmdb.org/t/p/w300/...",
    "youtube_id": "6_UlfsdGiEc",
    "genre": [
      { "genre_id": 2, "genre_name": "Drama" },
      { "genre_id": 3, "genre_name": "Western" }
    ],
    "admin_review": "...",
    "ranking": { "ranking_value": 1, "ranking_name": "Excellent" }
  }
]
```

---

#### Register User

```
POST /register
```

Creates a new user account. Passwords are hashed with bcrypt before storage.

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "password": "secret123",
  "role": "USER",
  "favourite_genres": [
    { "genre_id": 1, "genre_name": "Action" },
    { "genre_id": 2, "genre_name": "Drama" }
  ]
}
```

| Field | Validation |
|---|---|
| `first_name` | Required, 2–100 characters |
| `last_name` | Required, 2–100 characters |
| `email` | Required, valid email format |
| `password` | Required, minimum 6 characters |
| `role` | Must be `ADMIN` or `USER` |
| `favourite_genres` | Required, array of Genre objects |

**Response:** `201 Created`

---

#### Login

```
POST /login
```

Authenticates a user and returns JWT access & refresh tokens.

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "secret123"
}
```

**Response:** `200 OK`
```json
{
  "user_id": "...",
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "role": "USER",
  "token": "<jwt-access-token>",
  "refresh_token": "<jwt-refresh-token>",
  "favourite_genres": [...]
}
```

---

### Protected Endpoints

These endpoints require a valid JWT token in the `Authorization` header:

```
Authorization: Bearer <access-token>
```

---

#### Get Movie by IMDB ID

```
GET /movie/:imdb_id
```

Returns a single movie by its IMDB identifier.

**Response:** `200 OK` — Single movie object

---

#### Add Movie

```
POST /addmovie
```

Adds a new movie to the catalog.

**Request Body:**
```json
{
  "imdb_id": "tt0133093",
  "title": "The Matrix",
  "poster_path": "https://image.tmdb.org/t/p/w300/...",
  "youtube_id": "vKQi3bBA1y8",
  "genre": [
    { "genre_id": 1, "genre_name": "Action" },
    { "genre_id": 4, "genre_name": "Sci-Fi" }
  ],
  "ranking": { "ranking_value": 1, "ranking_name": "Excellent" }
}
```

**Response:** `201 Created`

---

#### Update Admin Review (ADMIN only)

```
PATCH /updatereview/:imdb_id
```

Submits an admin review for a movie. The review text is analyzed by OpenAI to determine a sentiment ranking, which is then stored alongside the review.

> **Requires** `ADMIN` role.

**Request Body:**
```json
{
  "admin_review": "Clint Eastwood was magnificent. Amazing cast and story!"
}
```

**Response:** `200 OK`
```json
{
  "ranking_name": "Excellent",
  "admin_review": "Clint Eastwood was magnificent. Amazing cast and story!"
}
```

---

#### Get Recommended Movies

```
GET /recommendedmovies
```

Returns personalized movie recommendations based on the authenticated user's favourite genres. Results are sorted by ranking (best first) and limited by the `RECOMMENDED_MOVIE_LIMIT` environment variable (default: 5).

**Response:** `200 OK` — Array of movie objects

---

## Data Models

### User

| Field | Type | Description |
|---|---|---|
| `_id` | ObjectID | MongoDB auto-generated ID |
| `user_id` | string | Application-generated unique ID |
| `first_name` | string | First name (2–100 chars) |
| `last_name` | string | Last name (2–100 chars) |
| `email` | string | Unique email address |
| `password` | string | bcrypt-hashed password |
| `role` | string | `ADMIN` or `USER` |
| `created_at` | datetime | Account creation timestamp |
| `updated_at` | datetime | Last update timestamp |
| `token` | string | Current JWT access token |
| `refresh_token` | string | Current JWT refresh token |
| `favourite_genres` | Genre[] | User's preferred movie genres |

### Movie

| Field | Type | Description |
|---|---|---|
| `_id` | ObjectID | MongoDB auto-generated ID |
| `imdb_id` | string | IMDB identifier (e.g. `tt0105695`) |
| `title` | string | Movie title (2–500 chars) |
| `poster_path` | string | URL to poster image |
| `youtube_id` | string | YouTube trailer video ID |
| `genre` | Genre[] | Array of genre classifications |
| `admin_review` | string | Admin-written review text |
| `ranking` | Ranking | Sentiment-based ranking |

### Genre

| Field | Type | Description |
|---|---|---|
| `genre_id` | int | Numeric genre identifier |
| `genre_name` | string | Genre display name (2–100 chars) |

### Ranking

| Field | Type | Description |
|---|---|---|
| `ranking_value` | int | Numeric rank (lower = better) |
| `ranking_name` | string | Sentiment label (e.g. "Excellent", "Terrible") |

---

## Authentication

The server uses **JWT (JSON Web Tokens)** with HS256 signing for authentication.

| Token Type | Expiry | Signing Key |
|---|---|---|
| Access Token | 24 hours | `SECRET_KEY` |
| Refresh Token | 7 days | `SECRET_REFRESH_KEY` |

### Token Claims

Both tokens contain the following custom claims:

- `Email`
- `FirstName`
- `LastName`
- `Role`
- `UserID`

Plus standard registered claims: `Issuer` (`MagicStream`), `IssuedAt`, `ExpiresAt`.

### Flow

1. User calls `POST /login` with email and password
2. Server validates credentials against bcrypt hash
3. Server generates access + refresh tokens
4. Tokens are stored in the user document in MongoDB
5. Client sends the access token in `Authorization: Bearer <token>` header for protected routes

---

## Architecture

### Dependency Injection

The server uses **dependency injection** for database access. A single `*mongo.Client` is created in `main()` via `database.Connect()` and passed through the call chain:

```
main() → routes.Setup*(router, client) → controllers.*(client) → database.OpenCollection(name, client)
```

### Middleware

The **AuthMiddleWare** intercepts all protected route requests:

1. Extracts the Bearer token from the `Authorization` header
2. Validates the JWT signature and expiration
3. Injects `userId` and `role` into the Gin context for downstream handlers

### Token Management

- **Generation:** `GenerateAllTokens()` creates both access and refresh tokens with embedded user claims
- **Validation:** `ValidateToken()` parses and verifies the JWT, checks expiration and signing method
- **Persistence:** `UpdateAllTokens(userID, token, refreshToken, client)` saves the latest tokens to the user's MongoDB document

### AI-Powered Review Ranking

When an admin submits a movie review via `PATCH /updatereview/:imdb_id`:

1. Rankings are fetched from the `rankings` MongoDB collection
2. A prompt is constructed using `BASE_PROMPT_TEMPLATE` with available ranking names
3. The review text is sent to **OpenAI** via langchaingo
4. The LLM returns a single sentiment word (e.g. "Excellent", "Good", "Terrible")
5. The matching ranking value is looked up and stored with the movie

### Personalized Recommendations

The `GET /recommendedmovies` endpoint:

1. Extracts the authenticated user's ID from the JWT context
2. Queries the user's `favourite_genres` from MongoDB (using field projection for efficiency)
3. Finds movies matching any of the user's preferred genres
4. Sorts results by `ranking.ranking_value` (ascending — best ranked first)
5. Limits results to `RECOMMENDED_MOVIE_LIMIT` (configurable via `.env`)

---

## MongoDB Collections

| Collection | Description |
|---|---|
| `users` | User accounts, credentials, tokens, and genre preferences |
| `movies` | Movie catalog with metadata, reviews, and rankings |
| `rankings` | Sentiment ranking definitions (name → numeric value mapping) |

All collections live under the database specified by `DATABASE_NAME` in `.env` (default: `magic-stream-movies`).
