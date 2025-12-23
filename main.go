func main() {
	authoRepo := autho.NewAuthoRepo()
	authoService := autho.NewAuthoService(&authoRepo)

	newAccount, err := authoService.RegisterService(autho.User{Email: "hello", Username: "hello", Password: "Hello"})

	if err != nil {
		fmt.Println("Cannot create an account because:", newAccount.ErrorResult.Error)
		return
	}

	fmt.Println("the current status is:::", newAccount.Username)
	fmt.Println("Password is:::", newAccount.Password)

	fmt.Println("Change Password")

	_, err = authoService.ChangePasswordService(&newAccount, "123")
	if err != nil {
		fmt.Println("Cannot change password:::", err)
		return
	}

	fmt.Println("New password is:::", newAccount.Password)

	// inputPass, _ := authoService.HashPassword("Hello")
	// fmt.Println("inputPassword is :::", inputPass)

	_, err = authoService.ValidatePassword(&newAccount, "123")
	if err != nil {
		fmt.Println("Password NOT match")
		return
	}

	fmt.Println("Current password is:", newAccount.Password)

	fmt.Println("Password matched")
}
