package model

type User struct {
	ID       int       `db:"id"`
	Fullname string    `json:"name"`
	Email    string    `json:"email"`
	Username string    `db:"username"`
	Password string    `db:"password"`
	Role     string    `db:"role"`
	Payments []Payment `json:"payments"`
}
