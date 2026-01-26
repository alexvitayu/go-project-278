package mocks

import (
	"code/internal/service"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockLinkService struct {
	mock.Mock
}

func (m *MockLinkService) CreateShortLink(ctx context.Context, shortname, originalUrl string) (*service.Link, error) {
	args := m.Called(ctx, shortname, originalUrl)
	return args.Get(0).(*service.Link), args.Error(1)
}

func (m *MockLinkService) GetLinks(ctx context.Context, limit, offset int32) ([]*service.Link, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*service.Link), args.Get(1).(int64), args.Error(2)
}

func (m *MockLinkService) GetLinkByID(ctx context.Context, id int64) (*service.Link, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*service.Link), args.Error(1)
}

func (m *MockLinkService) UpdateLinkByID(ctx context.Context, shortName, originalUrl string, id int64) (*service.Link, error) {
	args := m.Called(ctx, shortName, originalUrl, id)
	return args.Get(0).(*service.Link), args.Error(1)
}

func (m *MockLinkService) DeleteLinkByID(ctx context.Context, id int64) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLinkService) GetOriginalURLByShortName(ctx context.Context, shortName string) (*service.Link, error) {
	args := m.Called(ctx, shortName)
	return args.Get(0).(*service.Link), args.Error(1)
}
