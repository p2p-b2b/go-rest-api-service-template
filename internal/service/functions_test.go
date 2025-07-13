package service

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt" // Import bcrypt
)

func TestEncryptDecrypt(t *testing.T) {
	// Generate a random symmetric key
	symmetricKey := make([]byte, 32)
	if _, err := rand.Read(symmetricKey); err != nil {
		t.Fatalf("Failed to generate symmetric key: %v", err)
	}

	plaintext := []byte("This is a secret message")

	// Test Encrypt function
	ciphertext, err := Encrypt(plaintext, symmetricKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Test Decrypt function
	decryptedText, err := Decrypt(ciphertext, symmetricKey)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// Check if the decrypted text matches the original plaintext
	if !bytes.Equal(decryptedText, plaintext) {
		t.Errorf("Decrypted text = %s, want %s", decryptedText, plaintext)
	}
}

func TestEncrypt_ErrorCases(t *testing.T) {
	plaintext := []byte("This is a secret message")

	// Test Encrypt function with an invalid key size
	invalidKey := make([]byte, 10)
	_, err := Encrypt(plaintext, invalidKey)
	if err == nil {
		t.Error("Encrypt() error = nil, want error for invalid key size")
	}
}

func TestDecrypt_ErrorCases(t *testing.T) {
	// Generate a random symmetric key
	symmetricKey := make([]byte, 32)
	if _, err := rand.Read(symmetricKey); err != nil {
		t.Fatalf("Failed to generate symmetric key: %v", err)
	}

	// Test Decrypt function with an invalid ciphertext
	invalidCiphertext := []byte("invalid ciphertext")
	_, err := Decrypt(invalidCiphertext, symmetricKey)
	if err == nil {
		t.Error("Decrypt() error = nil, want error for invalid ciphertext")
	}

	// Test Decrypt function with a valid ciphertext but wrong key
	plaintext := []byte("This is a secret message")
	ciphertext, err := Encrypt(plaintext, symmetricKey)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	wrongKey := make([]byte, 32)
	if _, err := rand.Read(wrongKey); err != nil {
		t.Fatalf("Failed to generate wrong symmetric key: %v", err)
	}

	_, err = Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Error("Decrypt() error = nil, want error for wrong key")
	}
}

func TestConvertToSQLRegex(t *testing.T) {
	testCases := []struct {
		name     string
		resource string
		expected string
	}{
		{"login", "/auth/login", "^/auth/login$"},
		{"logout", "/auth/logout", "^/auth/logout$"},
		{"refresh", "/auth/refresh", "^/auth/refresh$"},
		{"register", "/auth/register", "^/auth/register$"},
		{"verify", "/auth/verify", "^/auth/verify$"},
		{"verify_star", "/auth/verify/*", "^/auth/verify/\\{[a-z_]{1,50}\\}$"},
		{"embeddings", "/embeddings", "^/embeddings$"},
		{"languages", "/languages", "^/languages$"},
		{"languages_uuid", "/languages/af2a595c-762d-4e71-ab7e-a64740653fa6", "^/languages/\\{[a-z_]{1,50}\\}$"},
		{"languages_uuid2", "/languages/8bf97eb1-e82d-4151-8b28-8be207e752a2", "^/languages/\\{[a-z_]{1,50}\\}$"},
		{"languages_star", "/languages/*", "^/languages/\\{[a-z_]{1,50}\\}$"},
		{"llm_engine_types", "/llm_engine_types", "^/llm_engine_types$"},
		{"llm_engine_types_uuid", "/llm_engine_types/af2a595c-762d-4e71-ab7e-a64740653fa6", "^/llm_engine_types/\\{[a-z_]{1,50}\\}$"},
		{"llm_engine_types_uuid2", "/llm_engine_types/8bf97eb1-e82d-4151-8b28-8be207e752a2", "^/llm_engine_types/\\{[a-z_]{1,50}\\}$"},
		{"llm_engine_types_star", "/llm_engine_types/*", "^/llm_engine_types/\\{[a-z_]{1,50}\\}$"},
		{"llm_engines", "/llm_engines", "^/llm_engines$"},
		{"llm_engines_uuid", "/llm_engines/af2a595c-762d-4e71-ab7e-a64740653fa6", "^/llm_engines/\\{[a-z_]{1,50}\\}$"},
		{"llm_engines_uuid2", "/llm_engines/8bf97eb1-e82d-4151-8b28-8be207e752a2", "^/llm_engines/\\{[a-z_]{1,50}\\}$"},
		{"llm_engines_star", "/llm_engines/*", "^/llm_engines/\\{[a-z_]{1,50}\\}$"},
		{"llm_translators_uuid", "/llm_translators/af2a595c-762d-4e71-ab7e-a64740653fa6", "^/llm_translators/\\{[a-z_]{1,50}\\}$"},
		{"llm_translators_star", "/llm_translators/*", "^/llm_translators/\\{[a-z_]{1,50}\\}$"},
		{"model_types_uuid", "/model_types/af2a595c-762d-4e71-ab7e-a64740653fa6", "^/model_types/\\{[a-z_]{1,50}\\}$"},
		{"model_types_uuid2", "/model_types/8bf97eb1-e82d-4151-8b28-8be207e752a2", "^/model_types/\\{[a-z_]{1,50}\\}$"},
		{"model_types_star", "/model_types/*", "^/model_types/\\{[a-z_]{1,50}\\}$"},
		{"models", "/models", "^/models$"},
		{"resources", "/resources", "^/resources$"},
		{"resources_uuid", "/resources/321d9da0-72b8-44fc-9549-8cdf55497e5d", "^/resources/\\{[a-z_]{1,50}\\}$"},
		{"resources_star", "/resources/*", "^/resources/\\{[a-z_]{1,50}\\}$"},
		{"resources_uuid_roles", "/resources/38963c2c-ea09-410a-b4ce-e30127906c06/roles", "^/resources/\\{[a-z_]{1,50}\\}/roles$"},
		{"resources_star_roles", "/resources/*/roles", "^/resources/\\{[a-z_]{1,50}\\}/roles$"},
		{"projects", "/projects", "^/projects$"},
		{"projects_uuid", "/projects/8bf97eb1-e82d-4151-8b28-8be207e752a2", "^/projects/\\{[a-z_]{1,50}\\}$"},
		{"projects_star", "/projects/*", "^/projects/\\{[a-z_]{1,50}\\}$"},
		{"projects_uuid_embeddings", "/projects/38963c2c-ea09-410a-b4ce-e30127906c06/embeddings", "^/projects/\\{[a-z_]{1,50}\\}/embeddings$"},
		{"projects_star_embeddings", "/projects/*/embeddings", "^/projects/\\{[a-z_]{1,50}\\}/embeddings$"},
		{"projects_uuid_embeddings_uuid", "/projects/bc53f7e8-c8f1-4023-8f3c-d19dbedb9c2a/embeddings/4685085f-e45d-415d-a5cb-bcf339a0aa1a", "^/projects/\\{[a-z_]{1,50}\\}/embeddings/\\{[a-z_]{1,50}\\}$"},
		{"projects_star_embeddings_uuid", "/projects/*/embeddings/4685085f-e45d-415d-a5cb-bcf339a0aa1a", "^/projects/\\{[a-z_]{1,50}\\}/embeddings/\\{[a-z_]{1,50}\\}$"},
		{"projects_uuid_embeddings_star", "/projects/bc53f7e8-c8f1-4023-8f3c-d19dbedb9c2a/embeddings/*", "^/projects/\\{[a-z_]{1,50}\\}/embeddings/\\{[a-z_]{1,50}\\}$"},
		{"projects_uuid_embeddings_uuid_ingest", "/projects/964e6b16-1def-4462-948c-d4bbe16cc83c/embeddings/38963c2c-ea09-410a-b4ce-e30127906c06/ingest", "^/projects/\\{[a-z_]{1,50}\\}/embeddings/\\{[a-z_]{1,50}\\}/ingest$"},
		{"projects_star_embeddings_uuid_ingest", "/projects/*/embeddings/38963c2c-ea09-410a-b4ce-e30127906c06/ingest", "^/projects/\\{[a-z_]{1,50}\\}/embeddings/\\{[a-z_]{1,50}\\}/ingest$"},
		{"projects_uuid_embeddings_star_ingest", "/projects/964e6b16-1def-4462-948c-d4bbe16cc83c/embeddings/*/ingest", "^/projects/\\{[a-z_]{1,50}\\}/embeddings/\\{[a-z_]{1,50}\\}/ingest$"},
		{"roles", "/roles", "^/roles$"},
		{"tokens", "/tokens", "^/tokens$"},
		{"users", "/users", "^/users$"},
		{"users_uuid", "/users/f8a8204f-ad6b-4fae-b4fb-335b7c64983c", "^/users/\\{[a-z_]{1,50}\\}$"},
		{"users_star", "/users/*", "^/users/\\{[a-z_]{1,50}\\}$"},
		{"users_uuid_authz", "/users/c2aac967-a79a-4fe4-9134-8c237534a33c/authz", "^/users/\\{[a-z_]{1,50}\\}/authz$"},
		{"users_star_authz", "/users/*/authz", "^/users/\\{[a-z_]{1,50}\\}/authz$"},
		{"users_uuid_projects", "/users/bc53f7e8-c8f1-4023-8f3c-d19dbedb9c2a/projects", "^/users/\\{[a-z_]{1,50}\\}/projects$"},
		{"users_star_projects", "/users/*/projects", "^/users/\\{[a-z_]{1,50}\\}/projects$"},
		{"star", "*", "^\\{[a-z_]{1,50}\\}$"},
		{"star_test", "/*/test", "^/\\{[a-z_]{1,50}\\}/test$"},
		{"test_star", "/test/*", "^/test/\\{[a-z_]{1,50}\\}$"},
		{"test_star_sub", "/test/*/sub", "^/test/\\{[a-z_]{1,50}\\}/sub$"},
		{"test_uuid_sub", "/test/123e4567-e89b-12d3-a456-426614174000/sub", "^/test/\\{[a-z_]{1,50}\\}/sub$"},
		{"test_uuid_sub_abc", "/test/123e4567-e89b-12d3-a456-426614174000/sub/abc", "^/test/\\{[a-z_]{1,50}\\}/sub/abc$"},
		{"test_star_sub_star", "/test/*/sub/*", "^/test/\\{[a-z_]{1,50}\\}/sub/\\{[a-z_]{1,50}\\}$"},
		{"test_uuid_sub_long_string", "/test/123e4567-e89b-12d3-a456-426614174000/sub/abc-def-ghi-jkl-mno", "^/test/\\{[a-z_]{1,50}\\}/sub/abc-def-ghi-jkl-mno$"},
		{"test_uuid_sub_uuid", "/test/123e4567-e89b-12d3-a456-426614174000/sub/123e4567-e89b-12d3-a456-426614174000", "^/test/\\{[a-z_]{1,50}\\}/sub/\\{[a-z_]{1,50}\\}$"},
		{"root", "/", "^/$"},
		{"empty", "", "^$"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := convertToSQLRegex(tc.resource)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestHashAndSaltPassword(t *testing.T) {
	password := "mysecretpassword"

	t.Run("success_default_cost", func(t *testing.T) {
		hashed, err := HashAndSaltPassword(password)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)

		// Verify the hash using bcrypt's built-in comparison
		err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
		assert.NoError(t, err, "Hashed password should match original password")
	})

	t.Run("success_specific_cost", func(t *testing.T) {
		cost := bcrypt.MinCost + 1 // Use a valid cost other than default
		hashed, err := HashAndSaltPassword(password, cost)
		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)

		// Verify the hash
		err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
		assert.NoError(t, err, "Hashed password with specific cost should match original password")

		// Optional: Check if the cost is embedded (though bcrypt handles this internally)
		costFromHash, err := bcrypt.Cost([]byte(hashed))
		assert.NoError(t, err)
		assert.Equal(t, cost, costFromHash, "Embedded cost should match the specified cost")
	})

	t.Run("error_cost_too_low", func(t *testing.T) {
		invalidCost := bcrypt.MinCost - 1
		_, err := HashAndSaltPassword(password, invalidCost)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cost value must be between")
	})

	t.Run("error_cost_too_high", func(t *testing.T) {
		invalidCost := bcrypt.MaxCost + 1
		_, err := HashAndSaltPassword(password, invalidCost)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cost value must be between")
	})
}

func TestComparePasswords(t *testing.T) {
	password := "anotherpassword123"
	hashedPassword, err := HashAndSaltPassword(password, bcrypt.MinCost) // Use min cost for faster tests
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)

	t.Run("success_match", func(t *testing.T) {
		match := ComparePasswords(hashedPassword, password)
		assert.True(t, match, "Correct password should match the hash")
	})

	t.Run("failure_no_match", func(t *testing.T) {
		wrongPassword := "wrongpassword"
		match := ComparePasswords(hashedPassword, wrongPassword)
		assert.False(t, match, "Incorrect password should not match the hash")
	})

	t.Run("failure_invalid_hash", func(t *testing.T) {
		invalidHash := "thisisnotavalidhash"
		match := ComparePasswords(invalidHash, password)
		assert.False(t, match, "Invalid hash should result in no match")
	})

	t.Run("failure_empty_hash", func(t *testing.T) {
		emptyHash := ""
		match := ComparePasswords(emptyHash, password)
		assert.False(t, match, "Empty hash should result in no match")
	})

	t.Run("failure_empty_password", func(t *testing.T) {
		emptyPassword := ""
		match := ComparePasswords(hashedPassword, emptyPassword)
		assert.False(t, match, "Empty password should not match a valid hash")
	})
}
