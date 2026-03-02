# User Package

This package manages user-related operations including authentication, profile management, user data persistence, and configuration storage. It follows the repository pattern with clean separation of concerns between HTTP handlers, business logic, and data access.

## 🎯 Features

- **User Management**: CRUD operations for user accounts
- **Profile Management**: User preferences and configuration storage
- **Repository Pattern**: Clean separation of data access and business logic
- **Firestore Integration**: Persistent storage with Firestore
- **HTTP Handlers**: RESTful API endpoints for user operations
- **Service Layer**: Business logic and validation
- **Interface-Based**: Dependency injection and testability

## 📦 Key Components

### Interfaces

```go
// ServiceAPI defines the business logic interface
type ServiceAPI interface {
    CreateUser(ctx context.Context, user *common.User) error
    GetUser(ctx context.Context, userID string) (*common.User, error)
    UpdateUser(ctx context.Context, user *common.User) error
    DeleteUser(ctx context.Context, userID string) error
    ListUsers(ctx context.Context, limit int) ([]*common.User, error)
    
    GetProfile(ctx context.Context, userID string) (*common.UserProfile, error)
    UpdateProfile(ctx context.Context, profile *common.UserProfile) error
}

// RepositoryAPI defines the data access interface
type RepositoryAPI interface {
    Create(ctx context.Context, user *common.User) error
    Get(ctx context.Context, userID string) (*common.User, error)
    Update(ctx context.Context, user *common.User) error
    Delete(ctx context.Context, userID string) error
    List(ctx context.Context, limit int) ([]*common.User, error)
    
    GetProfile(ctx context.Context, userID string) (*common.UserProfile, error)
    SaveProfile(ctx context.Context, profile *common.UserProfile) error
}
```

### Service Implementation

```go
type Service struct {
    repo   RepositoryAPI
    logger *slog.Logger
}

func NewService(repo RepositoryAPI, logger *slog.Logger) *Service {
    return &Service{
        repo:   repo,
        logger: logger,
    }
}
```

### Repository Implementation

```go
type Repository struct {
    db     db.ServiceInterface
    logger *slog.Logger
}

func NewRepository(db db.ServiceInterface, logger *slog.Logger) *Repository {
    return &Repository{
        db:     db,
        logger: logger,
    }
}
```

## 📁 Files

### `interfaces.go`
Defines interfaces for service and repository layers.

**Contents:**
- `ServiceAPI`: Business logic interface
- `RepositoryAPI`: Data access interface
- Interface documentation and contracts

### `service.go`
Implements business logic and validation.

**Key Functions:**
- `CreateUser()`: Creates a new user with validation
- `GetUser()`: Retrieves user by ID
- `UpdateUser()`: Updates user information
- `DeleteUser()`: Soft or hard delete user
- `GetProfile()`: Retrieves user profile and preferences
- `UpdateProfile()`: Updates user profile

### `repository.go`
Implements data access layer with Firestore.

**Key Functions:**
- `Create()`: Persists new user to database
- `Get()`: Retrieves user from database
- `Update()`: Updates user in database
- `Delete()`: Removes user from database
- `List()`: Queries users with pagination
- `GetProfile()`: Retrieves profile from database
- `SaveProfile()`: Persists profile to database

### `handler.go`
Implements HTTP request handlers.

**Endpoints:**
- `POST /api/users`: Create user
- `GET /api/users/:id`: Get user
- `PUT /api/users/:id`: Update user
- `DELETE /api/users/:id`: Delete user
- `GET /api/users`: List users
- `GET /api/users/:id/profile`: Get profile
- `PUT /api/users/:id/profile`: Update profile

## 🔧 Usage Examples

### Service Initialization

```go
import (
    "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/db"
    "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/user"
)

// Initialize dependencies
dbService, err := db.NewService(ctx, projectID, logger)
if err != nil {
    log.Fatal(err)
}
defer dbService.Close()

// Create repository
repo := user.NewRepository(dbService, logger)

// Create service
userService := user.NewService(repo, logger)

// Create HTTP handler
userHandler := user.NewHandler(userService, logger)
```

### Creating a User

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/common"

// Create user object
newUser := &common.User{
    ID:    "user-123",
    Email: "user@example.com",
    Name:  "John Doe",
}

// Create via service
err := userService.CreateUser(ctx, newUser)
if err != nil {
    // Handle error (validation, duplicate, etc.)
    return err
}
```

### Getting a User

```go
// Get user by ID
user, err := userService.GetUser(ctx, "user-123")
if err != nil {
    // Handle error (not found, etc.)
    return err
}

fmt.Printf("User: %s (%s)\n", user.Name, user.Email)
```

### Updating a User

```go
// Get existing user
user, err := userService.GetUser(ctx, "user-123")
if err != nil {
    return err
}

// Modify user
user.Name = "Jane Doe"
user.UpdatedAt = time.Now()

// Update via service
err = userService.UpdateUser(ctx, user)
if err != nil {
    return err
}
```

### Managing Profiles

```go
// Get user profile
profile, err := userService.GetProfile(ctx, "user-123")
if err != nil {
    return err
}

// Update preferences
profile.Preferences["theme"] = "dark"
profile.Preferences["notifications"] = true
profile.LastActivity = time.Now()

// Save profile
err = userService.UpdateProfile(ctx, profile)
if err != nil {
    return err
}
```

### HTTP Handler Usage

```go
// Register routes with handler
router := chi.NewRouter()

// User routes
router.Post("/api/users", userHandler.CreateUser)
router.Get("/api/users/{userID}", userHandler.GetUser)
router.Put("/api/users/{userID}", userHandler.UpdateUser)
router.Delete("/api/users/{userID}", userHandler.DeleteUser)
router.Get("/api/users", userHandler.ListUsers)

// Profile routes
router.Get("/api/users/{userID}/profile", userHandler.GetProfile)
router.Put("/api/users/{userID}/profile", userHandler.UpdateProfile)
```

## 🌐 API Reference

### Create User

**Request:**
```http
POST /api/users
Content-Type: application/json

{
  "id": "user-123",
  "email": "user@example.com",
  "name": "John Doe"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "name": "John Doe",
    "createdAt": "2025-01-27T10:30:00Z",
    "updatedAt": "2025-01-27T10:30:00Z"
  }
}
```

### Get User

**Request:**
```http
GET /api/users/user-123
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "id": "user-123",
    "email": "user@example.com",
    "name": "John Doe",
    "createdAt": "2025-01-27T10:30:00Z",
    "updatedAt": "2025-01-27T10:30:00Z"
  }
}
```

### Update Profile

**Request:**
```http
PUT /api/users/user-123/profile
Content-Type: application/json

{
  "preferences": {
    "theme": "dark",
    "language": "en",
    "notifications": true
  }
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "userId": "user-123",
    "preferences": {
      "theme": "dark",
      "language": "en",
      "notifications": true
    },
    "lastActivity": "2025-01-27T10:35:00Z"
  }
}
```

## 🏗️ Architecture

### Layered Architecture

```
┌─────────────────────────────────────────┐
│         HTTP Handler Layer              │
│  (handler.go - Request/Response)        │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│         Service Layer                   │
│  (service.go - Business Logic)          │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│         Repository Layer                │
│  (repository.go - Data Access)          │
└────────────────┬────────────────────────┘
                 │
┌────────────────▼────────────────────────┐
│         Database Layer                  │
│  (Firestore - Persistent Storage)       │
└─────────────────────────────────────────┘
```

### Dependency Flow

1. **Handler** receives HTTP requests
2. **Handler** calls **Service** with validated data
3. **Service** implements business logic and validation
4. **Service** calls **Repository** for data operations
5. **Repository** interacts with **Database**
6. Data flows back up through the layers

### Benefits

- **Testability**: Each layer can be tested independently
- **Maintainability**: Clear separation of concerns
- **Flexibility**: Easy to swap implementations (e.g., different database)
- **Reusability**: Service layer can be used by multiple interfaces (HTTP, CLI, gRPC)

## 🔒 Security

### Input Validation

```go
func (s *Service) CreateUser(ctx context.Context, user *common.User) error {
    // Validate email format
    if !isValidEmail(user.Email) {
        return errors.NewValidationError("invalid email format", nil)
    }
    
    // Validate name length
    if len(user.Name) < 2 || len(user.Name) > 100 {
        return errors.NewValidationError("name must be 2-100 characters", nil)
    }
    
    // Check for duplicate email
    existing, _ := s.repo.GetByEmail(ctx, user.Email)
    if existing != nil {
        return errors.NewValidationError("email already exists", nil)
    }
    
    return s.repo.Create(ctx, user)
}
```

### Authentication Context

```go
// Extract user from context
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
    // Extract authenticated user
    user, ok := common.UserFromContext(r.Context())
    if !ok {
        errors.HTTPErrorHandler(w, errors.NewUnauthorizedError("not authenticated"), h.logger, "")
        return
    }
    
    // Verify user has permission
    requestedUserID := chi.URLParam(r, "userID")
    if user.ID != requestedUserID {
        errors.HTTPErrorHandler(w, errors.NewForbiddenError("access denied"), h.logger, "")
        return
    }
    
    // Get profile
    profile, err := h.service.GetProfile(r.Context(), requestedUserID)
    if err != nil {
        errors.HTTPErrorHandler(w, err, h.logger, "")
        return
    }
    
    errors.SafeJSONResponse(w, profile)
}
```

### Data Sanitization

```go
import "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/security"

func (s *Service) UpdateProfile(ctx context.Context, profile *common.UserProfile) error {
    // Sanitize preference values
    for key, value := range profile.Preferences {
        if str, ok := value.(string); ok {
            profile.Preferences[key] = security.SanitizeInput(str)
        }
    }
    
    return s.repo.SaveProfile(ctx, profile)
}
```

## 🧪 Testing

### Unit Tests

```go
func TestCreateUser(t *testing.T) {
    // Create mock repository
    mockRepo := &MockRepository{}
    logger := slog.Default()
    
    // Create service
    service := user.NewService(mockRepo, logger)
    
    // Test user creation
    testUser := &common.User{
        ID:    "test-123",
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    err := service.CreateUser(context.Background(), testUser)
    assert.NoError(t, err)
    assert.True(t, mockRepo.CreateCalled)
}
```

### Integration Tests

```bash
# Start Firestore emulator
firebase emulators:start --only firestore

# Run integration tests
export FIRESTORE_EMULATOR_HOST=localhost:8080
go test -tags=integration ./file_server/advanced_file_operations/infrastructure/user/
```

### Mock Repository

```go
type MockRepository struct {
    users       map[string]*common.User
    profiles    map[string]*common.UserProfile
    CreateCalled bool
}

func (m *MockRepository) Create(ctx context.Context, user *common.User) error {
    m.CreateCalled = true
    m.users[user.ID] = user
    return nil
}

func (m *MockRepository) Get(ctx context.Context, userID string) (*common.User, error) {
    user, ok := m.users[userID]
    if !ok {
        return nil, errors.NewNotFoundError("user", userID)
    }
    return user, nil
}
```

## 📊 Data Models

### User Document

```go
type UserDocument struct {
    ID        string    `firestore:"id" json:"id"`
    Email     string    `firestore:"email" json:"email"`
    Name      string    `firestore:"name" json:"name"`
    Active    bool      `firestore:"active" json:"active"`
    CreatedAt time.Time `firestore:"createdAt" json:"createdAt"`
    UpdatedAt time.Time `firestore:"updatedAt" json:"updatedAt"`
}
```

### Profile Document

```go
type ProfileDocument struct {
    UserID       string                 `firestore:"userId" json:"userId"`
    Preferences  map[string]interface{} `firestore:"preferences" json:"preferences"`
    JobHistory   []JobSummary           `firestore:"jobHistory" json:"jobHistory"`
    LastActivity time.Time              `firestore:"lastActivity" json:"lastActivity"`
}
```

## 🐛 Troubleshooting

### Common Issues

**"user not found" errors:**
- Verify user ID is correct
- Check if user was created successfully
- Ensure database connection is working

**"duplicate email" errors:**
- Email addresses must be unique
- Check for existing user with same email
- Consider soft-delete for user removal

**"permission denied" errors:**
- Check Firestore security rules
- Verify service account has proper permissions
- Ensure user is authenticated

## 📚 Related Documentation

- [Infrastructure Overview](../README.md)
- [Database Package](../db/README.md) - Database service used by repository
- [Common Package](../common/README.md) - User data structures
- [Errors Package](../errors/README.md) - Error handling

## 🤝 Contributing

When enhancing the user package:

1. Follow the repository pattern
2. Add validation in the service layer
3. Keep handlers thin (delegate to service)
4. Write unit tests for business logic
5. Write integration tests for database operations
6. Update API documentation for new endpoints
7. Maintain backwards compatibility

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
