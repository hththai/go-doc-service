// service.go - Contains the core business logic for user operations

package user

type UserService struct {
    repo UserRepository
}

// NewUserService creates a new instance of UserService
func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(name, email, password string) error {
    // Business logic for registering a user
    newUser := User{Name: name, Email: email, Password: password}
    return s.repo.Save(newUser)
}
