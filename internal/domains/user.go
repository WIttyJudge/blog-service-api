package domains

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name" validate:"required,max=50"`
	LastName  string `json:"last_name" validate:"required,max=50"`
	Username  string `json:"username" validate:"required,max=50"`
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate"required,min=3,max=100"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserRepository interface {
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	Create(*User) error
}
