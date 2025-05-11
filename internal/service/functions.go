package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
)

func Encrypt(plaintext, symmetricKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func Decrypt(ciphertext, symmetricKey []byte) ([]byte, error) {
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func CiphertextToString(ciphertext []byte) (string, error) {
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func StringToCiphertext(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// HashAndSaltPassword hashes and salts the password
// It uses bcrypt to hash the password with a cost of 10.
// The hashed password is returned as a string.
func HashAndSaltPassword(password string, cost ...int) (string, error) {
	var costVal int
	if len(cost) > 0 {
		if cost[0] < bcrypt.MinCost || cost[0] > bcrypt.MaxCost {
			return "", fmt.Errorf("cost value must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
		}
		costVal = cost[0]
	} else {
		costVal = bcrypt.DefaultCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), costVal)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// ComparePasswords compares the hashed password and the plain password
// It uses bcrypt to compare the hashed password with the plain password.
// It returns true if the passwords match, false otherwise.
func ComparePasswords(hashedPwd string, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	return err == nil
}
