package service

import (
	"context"
	"errors"
	"testing"

	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
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
