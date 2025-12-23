package autho

type AuthoRepo interface {
	Register(user User) (User, error)
	ChangePassword(user *User, newPassword string) (*User, error)
	ValidatePassword(user *User, inputPassword string) (bool, error)
}

type AuthoRepoImpl struct {
}

// Constructor return interface.
func NewAuthoRepo() AuthoRepoImpl {
	return AuthoRepoImpl{}
}

func (r *AuthoRepoImpl) Register(user User) (User, error) {
	return user, nil
}

func (r *AuthoRepoImpl) ChangePassword(user *User, newPassword string) (*User, error) {
	return user, nil
}

func (r *AuthoRepoImpl) ValidatePassword(user *User, inputPassword string) (bool, error) {
	return true, nil
}
