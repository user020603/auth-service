package models

type User struct {
	ID       uint `json:"id" gorm:"primaryKey;"`
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
	Name     string `json:"name" gorm:"not null"`
	Email    string `json:"email" gorm:"unique;not null"`
	Role     string `json:"role" gorm:"not null"`
}
