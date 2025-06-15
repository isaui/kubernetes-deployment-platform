# PenDeploy-Handal

A robust Go REST API for managing deployments with authentication and authorization.

## Features

- User authentication and authorization
- Deployment management (create, view, update, delete)
- Role-based access control
- CORS support
- Environment configuration

## Project Structure

```
pendeploy-handal/
├── config/             # Configuration functions
├── controllers/        # Request handlers
├── middleware/         # Middleware functions
├── models/             # Data models
├── routes/             # API routes
├── .env                # Environment variables
├── go.mod              # Go module file
├── main.go             # Entry point
└── README.md           # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.21 or higher

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/isabu/pendeploy-handal.git
   cd pendeploy-handal
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create a `.env` file with your configuration (example provided)

4. Run the application:
   ```bash
   go run main.go
   ```

## API Endpoints

### Public Endpoints

- `GET /` - Health check
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get token

### Protected Endpoints (require authentication)

**Deployments:**
- `GET /api/deployments` - List all deployments
- `GET /api/deployments/:id` - Get deployment details
- `POST /api/deployments` - Create a new deployment
- `PUT /api/deployments/:id` - Update a deployment
- `DELETE /api/deployments/:id` - Delete a deployment
- `PATCH /api/deployments/:id/status` - Update deployment status

**Users (Admin only):**
- `GET /api/users` - List all users
- `GET /api/users/:id` - Get user details
- `PUT /api/users/:id` - Update a user
- `DELETE /api/users/:id` - Delete a user

## Environment Variables

- `PORT` - Server port (default: 8080)
- `GIN_MODE` - Gin mode (debug or release)
- `DATABASE_URL` - Database connection string
- `JWT_SECRET` - Secret key for JWT token generation

## Future Enhancements

- Database integration (PostgreSQL, MongoDB)
- Logging system
- Unit and integration tests
- Docker containerization
- CI/CD pipeline integration
