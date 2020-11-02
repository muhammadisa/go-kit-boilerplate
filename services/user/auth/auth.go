package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword hashing password
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
}

// VerifyPassword compare hashed password with password string
func VerifyPassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
}
