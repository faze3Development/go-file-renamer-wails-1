# Database Package

This package manages database connections and provides access to Firestore for persistent storage of user data, job histories, and application state. It follows clean architecture principles with repository patterns and dependency injection.

## 🎯 Features

- **Firestore Integration**: Cloud-native NoSQL database access
- **Connection Management**: Singleton pattern for efficient connection handling
- **Service Interface**: Clean abstraction for database operations
- **Lifecycle Management**: Proper initialization and cleanup
- **Error Handling**: Consistent error types from infrastructure/errors
- **Context Support**: Request-scoped database operations

## 📦 Key Components

### Database Service

The service provides access to the Firestore client and manages its lifecycle:

```go
type ServiceInterface interface {
    // GetClient returns the Firestore client
    GetClient() *firestore.Client
    
    // Close gracefully closes the database connection
    Close() error
    
    // IsHealthy checks if the database connection is healthy
    IsHealthy(ctx context.Context) bool
}
```

### Service Implementation

```go
type Service struct {
    client *firestore.Client
    logger *slog.Logger
}

// NewService initializes the Firestore client
func NewService(ctx context.Context, projectID string, logger *slog.Logger) (*Service, error)

// GetClient returns the Firestore client for database operations
func (s *Service) GetClient() *firestore.Client

// Close gracefully shuts down the Firestore client
func (s *Service) Close() error
```

## 📁 Files

### `service.go`
Main service implementation with initialization and lifecycle management.

**Key Functions:**
- `NewService()`: Creates and initializes the Firestore service
- `GetClient()`: Returns the Firestore client for operations
- `Close()`: Gracefully closes the connection
- `IsHealthy()`: Checks connection health

### `db.go`
Database utilities and helper functions.

**Contents:**
- Collection name constants
- Document path helpers
- Query builders
- Transaction helpers

## 🔧 Usage

### Service Initialization

Initialize the database service during application startup:

```go
import (
    "context"
    "go-file-renamer-wails/file_server/advanced_file_operations/infrastructure/db"
)

// In main.go or initialization code
ctx := context.Background()
projectID := os.Getenv("FIRESTORE_PROJECT_ID")
if projectID == "" {
    projectID = "your-project-id"
}

dbService, err := db.NewService(ctx, projectID, logger)
if err != nil {
    logger.Error("failed to initialize database", "error", err)
    return err
}
defer dbService.Close()

// Pass to services via dependency injection
userService := user.NewService(dbService, logger)
jobService := jobs.NewService(dbService, logger)
```

### Basic Database Operations

```go
// Get Firestore client
client := dbService.GetClient()

// Create a document
doc := client.Collection("users").Doc(userID)
_, err := doc.Set(ctx, map[string]interface{}{
    "name":      "John Doe",
    "email":     "john@example.com",
    "createdAt": time.Now(),
})
if err != nil {
    return errors.NewDatabaseError("failed to create user", err)
}

// Read a document
snapshot, err := doc.Get(ctx)
if err != nil {
    if status.Code(err) == codes.NotFound {
        return errors.NewNotFoundError("user", userID)
    }
    return errors.NewDatabaseError("failed to get user", err)
}

var user User
if err := snapshot.DataTo(&user); err != nil {
    return errors.NewSystemError("failed to parse user data", err)
}

// Update a document
_, err = doc.Update(ctx, []firestore.Update{
    {Path: "name", Value: "Jane Doe"},
    {Path: "updatedAt", Value: time.Now()},
})
if err != nil {
    return errors.NewDatabaseError("failed to update user", err)
}

// Delete a document
_, err = doc.Delete(ctx)
if err != nil {
    return errors.NewDatabaseError("failed to delete user", err)
}
```

### Query Operations

```go
// Simple query
client := dbService.GetClient()
iter := client.Collection("users").
    Where("active", "==", true).
    OrderBy("createdAt", firestore.Desc).
    Limit(10).
    Documents(ctx)

defer iter.Stop()
for {
    doc, err := iter.Next()
    if err == iterator.Done {
        break
    }
    if err != nil {
        return errors.NewDatabaseError("query failed", err)
    }
    
    var user User
    if err := doc.DataTo(&user); err != nil {
        logger.Warn("failed to parse user", "error", err)
        continue
    }
    users = append(users, user)
}
```

### Transactions

```go
client := dbService.GetClient()

err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
    // Read operations
    doc := client.Collection("counters").Doc("users")
    snapshot, err := tx.Get(doc)
    if err != nil {
        return err
    }
    
    var counter int64
    if err := snapshot.DataTo(&counter); err != nil {
        return err
    }
    
    // Write operations
    return tx.Set(doc, map[string]interface{}{
        "count": counter + 1,
        "updatedAt": firestore.ServerTimestamp,
    })
})

if err != nil {
    return errors.NewDatabaseError("transaction failed", err)
}
```

### Batch Operations

```go
client := dbService.GetClient()
batch := client.Batch()

// Add multiple operations to batch
for _, user := range users {
    doc := client.Collection("users").Doc(user.ID)
    batch.Set(doc, user)
}

// Commit all operations atomically
_, err := batch.Commit(ctx)
if err != nil {
    return errors.NewDatabaseError("batch commit failed", err)
}
```

## 🏗️ Architecture Patterns

### Repository Pattern

Use the repository pattern for data access:

```go
// Repository interface
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    Get(ctx context.Context, id string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit int) ([]*User, error)
}

// Repository implementation
type userRepository struct {
    db db.ServiceInterface
}

func NewUserRepository(db db.ServiceInterface) UserRepository {
    return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *User) error {
    client := r.db.GetClient()
    doc := client.Collection("users").Doc(user.ID)
    _, err := doc.Set(ctx, user)
    if err != nil {
        return errors.NewDatabaseError("failed to create user", err)
    }
    return nil
}
```

### Dependency Injection

Inject the database service through constructors:

```go
// Service with database dependency
type UserService struct {
    db     db.ServiceInterface
    logger *slog.Logger
}

func NewUserService(db db.ServiceInterface, logger *slog.Logger) *UserService {
    return &UserService{
        db:     db,
        logger: logger,
    }
}

func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    // Use injected database service
    client := s.db.GetClient()
    // ... database operations
}
```

## 🗄️ Data Models

### Collections

Standard collections used by the application:

```go
const (
    UsersCollection      = "users"
    ProfilesCollection   = "profiles"
    JobsCollection       = "jobs"
    JobHistoryCollection = "job_history"
    ConfigsCollection    = "configs"
)
```

### Document Structure

Example document structures:

```go
// User document
type UserDocument struct {
    ID        string                 `firestore:"id"`
    Email     string                 `firestore:"email"`
    Name      string                 `firestore:"name"`
    Active    bool                   `firestore:"active"`
    CreatedAt time.Time              `firestore:"createdAt"`
    UpdatedAt time.Time              `firestore:"updatedAt"`
    Metadata  map[string]interface{} `firestore:"metadata,omitempty"`
}

// Job document
type JobDocument struct {
    ID          string    `firestore:"id"`
    UserID      string    `firestore:"userId"`
    Status      string    `firestore:"status"`
    TotalFiles  int       `firestore:"totalFiles"`
    ProcessedFiles int    `firestore:"processedFiles"`
    CreatedAt   time.Time `firestore:"createdAt"`
    CompletedAt time.Time `firestore:"completedAt,omitempty"`
}
```

## 🔒 Security

### Firestore Security Rules

Example security rules for Firestore:

```javascript
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Users can only read/write their own data
    match /users/{userId} {
      allow read, write: if request.auth != null && request.auth.uid == userId;
    }
    
    // Jobs can only be read/written by the owner
    match /jobs/{jobId} {
      allow read, write: if request.auth != null && 
                             resource.data.userId == request.auth.uid;
    }
    
    // Admin-only collections
    match /admin/{document=**} {
      allow read, write: if request.auth != null && 
                             request.auth.token.admin == true;
    }
  }
}
```

### Connection Security

- Uses Google Cloud IAM for authentication
- Supports service account credentials
- Enforces TLS/SSL for all connections
- Supports VPC Service Controls

## 📊 Performance

### Best Practices

1. **Connection Pooling**: Reuse the Firestore client (singleton pattern)
2. **Batch Operations**: Use batch writes for multiple operations
3. **Indexes**: Create composite indexes for complex queries
4. **Pagination**: Use cursor-based pagination for large result sets
5. **Caching**: Cache frequently accessed data

### Query Optimization

```go
// Use indexes for better performance
client.Collection("users").
    Where("status", "==", "active").
    Where("createdAt", ">", yesterday).
    OrderBy("createdAt", firestore.Desc).
    Limit(100)

// Use cursors for pagination
query := client.Collection("users").
    OrderBy("createdAt", firestore.Desc).
    Limit(20)

// Get first page
docs := query.Documents(ctx)

// Get next page using cursor
lastDoc := docs[len(docs)-1]
nextQuery := query.StartAfter(lastDoc)
```

## 🧪 Testing

### Mock Service

For testing, use a mock database service:

```go
type MockDBService struct {
    client *firestore.Client
}

func NewMockDBService() *MockDBService {
    // Use Firestore emulator or in-memory mock
    return &MockDBService{}
}

func (m *MockDBService) GetClient() *firestore.Client {
    return m.client
}

func (m *MockDBService) Close() error {
    return nil
}
```

### Integration Tests

```bash
# Start Firestore emulator
firebase emulators:start --only firestore

# Run tests against emulator
export FIRESTORE_EMULATOR_HOST=localhost:8080
go test ./file_server/advanced_file_operations/infrastructure/db/
```

## 🔧 Configuration

### Environment Variables

```bash
# Required
FIRESTORE_PROJECT_ID=your-project-id

# Optional
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
FIRESTORE_EMULATOR_HOST=localhost:8080  # For testing
```

### Service Account Setup

1. Create a service account in Google Cloud Console
2. Grant "Cloud Datastore User" role
3. Download JSON key file
4. Set `GOOGLE_APPLICATION_CREDENTIALS` environment variable

## 🐛 Troubleshooting

### Common Issues

**"permission denied" errors:**
- Check Firestore security rules
- Verify service account permissions
- Ensure user is authenticated

**"not found" errors:**
- Document may not exist
- Check collection/document path
- Verify query conditions

**"deadline exceeded" errors:**
- Network connectivity issues
- Firestore quota limits reached
- Query is too complex or slow

## 📚 Related Documentation

- [Firestore Documentation](https://firebase.google.com/docs/firestore)
- [Go Client Library](https://pkg.go.dev/cloud.google.com/go/firestore)
- [Infrastructure Overview](../README.md)
- [User Package](../user/README.md) - Uses database service

## 🤝 Contributing

When working with the database:

1. Follow the repository pattern for data access
2. Use proper error handling with structured errors
3. Always close resources (defer client.Close())
4. Add indexes for complex queries
5. Document new collections and structures
6. Write integration tests

## 📄 License

Copyright (c) 2025 FAZE3 DEVELOPMENT LLC. All rights reserved.
