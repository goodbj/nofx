package main

import (
	"fmt"
	"nofx/store"
)

func main() {
	st, err := store.New("data/data.db")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer st.Close()

	// Query all users
	var users []struct {
		ID          string `gorm:"column:id"`
		Email       string `gorm:"column:email"`
		OTPVerified bool   `gorm:"column:otp_verified"`
	}

	err = st.GormDB().Table("users").Find(&users).Error
	if err != nil {
		fmt.Printf("Failed to query users: %v\n", err)
		return
	}

	fmt.Println("现有用户:")
	fmt.Println("========")
	for _, user := range users {
		fmt.Printf("ID: %s\n", user.ID)
		fmt.Printf("Email: %s\n", user.Email)
		fmt.Printf("OTP Verified: %v\n", user.OTPVerified)
		fmt.Println("--------")
	}
}
