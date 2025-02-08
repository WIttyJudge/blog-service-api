package domains

import "time"

type User struct {
	ID        int       `json:"id"`
	FirstName string    `json:"first_name" validate:"required,max=50"`
	LastName  string    `json:"last_name" validate:"required,max=50"`
	Email     string    `json:"email" validate:"required,email,max=255"`
	Password  string    `json:"password" validate:"required,min=3,max=100"`
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
