package models

type User struct {
	ID       uint   `json:"id" gorm:"id"`
	Username string `json:"username" gorm:"username"`
	Password string `json:"password" gorm:"password"`
}

func (User) TableName() string {
	return "users"
}
