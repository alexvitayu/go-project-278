// здесь реализуем все методы интерфейса Querier
// этот мок может быть ручным или сгенерированным
package mocks

import (
	"code/internal/db/postgres_db"
	"code/internal/db/visits"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockQuerier struct {
	mock.Mock
}

func (m *MockQuerier) CreateLink(ctx context.Context, arg postgres_db.CreateLinkParams) (postgres_db.CreateLinkRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(postgres_db.CreateLinkRow), args.Error(1)
}

func (m *MockQuerier) DeleteLinkByID(ctx context.Context, id int64) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) GetLinkByID(ctx context.Context, id int64) (postgres_db.GetLinkByIDRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(postgres_db.GetLinkByIDRow), args.Error(1)
}

func (m *MockQuerier) GetLinks(ctx context.Context, arg postgres_db.GetLinksParams) ([]postgres_db.GetLinksRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]postgres_db.GetLinksRow), args.Error(1)
}

func (m *MockQuerier) GetOriginalURLByShortName(ctx context.Context, shortName string) (postgres_db.GetOriginalURLByShortNameRow, error) {
	args := m.Called(ctx, shortName)
	return args.Get(0).(postgres_db.GetOriginalURLByShortNameRow), args.Error(1)
}

func (m *MockQuerier) GetTotalLinks(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockQuerier) UpdateLinkByID(ctx context.Context, arg postgres_db.UpdateLinkByIDParams) (postgres_db.UpdateLinkByIDRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(postgres_db.UpdateLinkByIDRow), args.Error(1)
}

type MockVisits struct {
	mock.Mock
}

func (mv *MockVisits) CreateVisit(ctx context.Context, arg visits.CreateVisitParams) error {
	args := mv.Called(ctx, arg)
	return args.Error(0)
}

func (mv *MockVisits) GetVisits(ctx context.Context, arg visits.GetVisitsParams) ([]visits.GetVisitsRow, error) {
	args := mv.Called(ctx, arg)
	return args.Get(0).([]visits.GetVisitsRow), args.Error(1)
}

func (mv *MockVisits) GetTotalVisits(ctx context.Context) (int64, error) {
	args := mv.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}
