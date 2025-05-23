package models

type User struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid"`
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
	Email    string `json:"email" gorm:"unique;not null"`
	Role     string `json:"role" gorm:"not null"`
}
