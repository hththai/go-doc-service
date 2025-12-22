package auth

type AuthService struct {
	authRepo AuthRepository
}

func NewAuthService(authRepo AuthRepository) *AuthService {
	return &AuthService{authRepo: authRepo}
}

func (s *AuthService) Register(username string, password string) (Account, error) {
	return s.authRepo.Register(username, password)
}
