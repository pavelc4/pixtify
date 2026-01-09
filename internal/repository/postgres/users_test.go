package postgres

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	dsn := "host=localhost port=5432 user=pixtify password=pixtify_dev_password dbname=pixtify_db sslmode=disable"

	var err error
	testDB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := testDB.Ping(); err != nil {
		log.Fatalf("Failed to ping test database: %v", err)
	}

	log.Println("Test database connected successfully")

	code := m.Run()

	testDB.Close()

	os.Exit(code)
}

func cleanupUsers(t *testing.T) {
	_, err := testDB.Exec("DELETE FROM users")
	require.NoError(t, err, "Failed to cleanup users table")
}

func createTestUser(t *testing.T, username, email string) *User {
	return &User{
		Username:     username,
		Email:        email,
		PasswordHash: "hashed_password_123",
		FullName:     stringPtr("Test User"),
		Role:         "user",
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestUserRepository_Create(t *testing.T) {
	cleanupUsers(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	t.Run("successfully create user", func(t *testing.T) {
		user := createTestUser(t, "testuser", "test@example.com")

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		assert.NotEqual(t, uuid.Nil, user.ID)

		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())

		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("fail to create user with duplicate email", func(t *testing.T) {
		user1 := createTestUser(t, "user1", "duplicate@example.com")
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := createTestUser(t, "user2", "duplicate@example.com")
		err = repo.Create(ctx, user2)
		assert.Error(t, err, "Should fail with duplicate email")
	})

	t.Run("fail to create user with duplicate username", func(t *testing.T) {
		user1 := createTestUser(t, "duplicateuser", "email1@example.com")
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := createTestUser(t, "duplicateuser", "email2@example.com")
		err = repo.Create(ctx, user2)
		assert.Error(t, err, "Should fail with duplicate username")
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	cleanupUsers(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	t.Run("successfully get user by ID", func(t *testing.T) {

		user := createTestUser(t, "getbyid", "getbyid@example.com")
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found)

		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Username, found.Username)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("return nil for non-existent user", func(t *testing.T) {
		randomID := uuid.New()
		found, err := repo.GetByID(ctx, randomID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	cleanupUsers(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	t.Run("successfully get user by email", func(t *testing.T) {

		user := createTestUser(t, "getbyemail", "getbyemail@example.com")
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.GetByEmail(ctx, "getbyemail@example.com")
		require.NoError(t, err)
		require.NotNil(t, found)

		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("return nil for non-existent email", func(t *testing.T) {
		found, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

// Test Update
func TestUserRepository_Update(t *testing.T) {
	cleanupUsers(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	t.Run("successfully update user", func(t *testing.T) {

		user := createTestUser(t, "updateuser", "update@example.com")
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		user.Username = "updatedusername"
		user.FullName = stringPtr("Updated Name")
		user.Bio = stringPtr("New bio")

		err = repo.Update(ctx, user)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)

		assert.Equal(t, "updatedusername", updated.Username)
		assert.Equal(t, "Updated Name", *updated.FullName)
		assert.Equal(t, "New bio", *updated.Bio)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	cleanupUsers(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	t.Run("successfully delete user", func(t *testing.T) {

		user := createTestUser(t, "deleteuser", "delete@example.com")
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		err = repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("delete non-existent user should not error", func(t *testing.T) {
		randomID := uuid.New()
		err := repo.Delete(ctx, randomID)
		assert.NoError(t, err)
	})
}
