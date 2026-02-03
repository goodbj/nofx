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

	// Update OTP verified status for test user
	result := st.GormDB().Table("users").Where("email = ?", "test@test.com").Update("otp_verified", true)
	if result.Error != nil {
		fmt.Printf("Failed to update user: %v\n", result.Error)
		return
	}

	fmt.Printf("Updated %d rows\n", result.RowsAffected)

	// Verify the update
	var user struct {
		ID          string `gorm:"column:id"`
		Email       string `gorm:"column:email"`
		OTPVerified bool   `gorm:"column:otp_verified"`
	}

	err = st.GormDB().Table("users").Where("email = ?", "test@test.com").First(&user).Error
	if err != nil {
		fmt.Printf("Failed to query user: %v\n", err)
		return
	}

	fmt.Printf("User updated - Email: %s, OTP Verified: %v\n", user.Email, user.OTPVerified)
}
