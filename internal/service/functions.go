package service

import (
	"golang.org/x/crypto/bcrypt"
)

// hashAndSaltPassword hashes and salts the password.
func hashAndSaltPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// comparePasswords compares the hashed password and the plain password.
// func comparePasswords(hashedPwd string, plainPwd string) bool {
// 	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
// 	return err == nil
// }
