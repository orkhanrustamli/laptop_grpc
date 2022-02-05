package service

import (
	"fmt"

	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password string
	Role     string
}

func NewUser(username, password, role string) (*User, error) {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("cannot generate password hash: %v", err)
	}

	user := &User{
		Username: username,
		Password: string(hashedPass),
		Role:     role,
	}

	return user, nil
}

func (user *User) IsCorrectPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (user *User) Clone() (*User, error) {
	clone := &User{}
	return clone, copier.Copy(clone, user)
}
