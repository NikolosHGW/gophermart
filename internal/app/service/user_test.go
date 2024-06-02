package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Save(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByLogin(ctx context.Context, login string) (bool, error) {
	args := m.Called(ctx, login)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) FindByLogin(ctx context.Context, login string) (*entity.User, error) {
	args := m.Called(ctx, login)
	if usr, ok := args.Get(0).(*entity.User); ok {
		return usr, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestUserService_Register(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("ExistsByLogin", mock.Anything, "test_login").Return(false, nil)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)

	logger, _ := zap.NewDevelopment()
	service := NewUserService(mockRepo, logger, "test_secret")

	user, err := service.Register(context.Background(), "test_login", "test_password")
	assert.NoError(t, err)
	assert.NotNil(t, user)
}

func TestUserService_Register_ExistsByLoginError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("ExistsByLogin", mock.Anything, "test_login").Return(false, errors.New("database error"))

	logger, _ := zap.NewDevelopment()
	service := NewUserService(mockRepo, logger, "test_secret")

	_, err := service.Register(context.Background(), "test_login", "test_password")
	assert.Error(t, err)
	assert.EqualError(t, err, "ошибка сервера: database error")
}

func TestUserService_Register_SaveError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockRepo.On("ExistsByLogin", mock.Anything, "test_login").Return(false, nil)
	mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.User")).Return(errors.New("save error"))

	logger, _ := zap.NewDevelopment()
	service := NewUserService(mockRepo, logger, "test_secret")

	_, err := service.Register(context.Background(), "test_login", "test_password")
	assert.Error(t, err)
	assert.EqualError(t, err, "ошибка при сохранении пользователя: save error")
}

func TestUserService_GenerateJWT(t *testing.T) {
	user := &entity.User{
		ID: 1,
	}

	logger, _ := zap.NewDevelopment()
	service := NewUserService(nil, logger, "test_secret")

	token, err := service.GenerateJWT(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestUserService_GenerateJWT_Error(t *testing.T) {
	user := &entity.User{
		ID: 1,
	}

	logger, _ := zap.NewDevelopment()
	service := NewUserService(nil, logger, "")

	_, err := service.GenerateJWT(user)
	assert.Error(t, err)
}

func TestUserService_Authenticate(t *testing.T) {
	mockRepo := new(MockUserRepository)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("test_password"), bcrypt.DefaultCost)
	mockUser := &entity.User{
		ID:       1,
		Login:    "test_login",
		Password: string(hashedPassword),
	}

	mockRepo.On("FindByLogin", mock.Anything, "test_login").Return(mockUser, nil)
	mockRepo.On("FindByLogin", mock.Anything, "wrong_login").Return(nil, domain.ErrInvalidCredentials)

	logger, _ := zap.NewDevelopment()
	service := NewUserService(mockRepo, logger, "test_secret")

	t.Run("Положительный тест: успешная аутентификация", func(t *testing.T) {
		user, err := service.Authenticate(context.Background(), "test_login", "test_password")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("Отрицательный тест: неверные учетные данные", func(t *testing.T) {
		_, err := service.Authenticate(context.Background(), "test_login", "wrong_password")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
	})

	t.Run("Отрицательный тест: ошибка при поиске пользователя", func(t *testing.T) {
		_, err := service.Authenticate(context.Background(), "wrong_login", "test_password")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
	})
}
