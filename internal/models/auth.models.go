package models

type UserRegist struct {
	Firstname string `json:"first_name" binding:"required"`
	Lastname  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone_number" binding:"required"`
	Address   string `json:"address" binding:"required"`
	Pin       string `json:"pin" binding:"required,len=6"`
}

type UserLogin struct {
	Phone string `json:"phone_number" binding:"required"`
	Pin   string `json:"pin" binding:"required,len=6"`
}
