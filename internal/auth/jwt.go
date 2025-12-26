package auth

type JWTService struct {
	secret string
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: secret}
}

func (j *JWTService) GenerateToken(userID int, role string) (string, error) {
	// TODO: Implement token generation
	return "", nil
}
