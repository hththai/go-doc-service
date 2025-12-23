package autho

type AuthoRepo interface {
	Register(user User) (User, error)
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
