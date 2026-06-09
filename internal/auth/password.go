package auth

import "golang.org/x/crypto/bcrypt"

// PasswordHasher 定义密码哈希和校验能力。
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password string, passwordHash string) bool
}

// BcryptPasswordHasher 是默认的密码哈希实现。
type BcryptPasswordHasher struct {
	cost int
}

// NewPasswordHasher 返回默认 bcrypt 哈希器。
func NewPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

// NewBcryptPasswordHasher 允许指定 bcrypt cost。
func NewBcryptPasswordHasher(cost int) *BcryptPasswordHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	return &BcryptPasswordHasher{cost: cost}
}

// Hash 生成 bcrypt 哈希，数据库只保存该值，不保存明文密码。
func (h *BcryptPasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Compare 校验明文密码和数据库密码哈希是否匹配。
func (h *BcryptPasswordHasher) Compare(password string, passwordHash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) == nil
}

// HashPassword 是简单场景下可直接调用的便捷函数。
func HashPassword(password string) (string, error) {
	return NewPasswordHasher().Hash(password)
}

// ComparePassword 是简单场景下可直接调用的便捷函数。
func ComparePassword(password string, passwordHash string) bool {
	return NewPasswordHasher().Compare(password, passwordHash)
}
