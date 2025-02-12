package domains

import "time"

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRepository interface {
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	CheckIfExistsByEmail(email string) bool
}

type UserService interface {
	GetByEmail(email string) (*User, error)
	Create(user *User) error
	VerifyCredentials(email string, password string) error
	CheckIfExistsByEmail(email string) bool
}
