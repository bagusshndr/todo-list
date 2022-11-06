package mocks

import (
	context "context"

	domain "github.com/bxcodec/go-clean-arch/domain"
	mock "github.com/stretchr/testify/mock"
)

type TodoRepositoryMock struct {
	mock.Mock
}

func (m *TodoRepositoryMock) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)

	return args.Error(0)
}

func (m *TodoRepositoryMock) Fetch(ctx context.Context, cursor string, num int64) ([]domain.Todo, string, error) {
	args := m.Called(ctx, cursor, num)

	return args.Get(0).([]domain.Todo), args.Get(0).(string), args.Error(1)
}

func (m *TodoRepositoryMock) GetByID(ctx context.Context, id int64) (domain.Todo, error) {
	args := m.Called(ctx, id)

	return args.Get(0).(domain.Todo), args.Error(1)
}

func (m *TodoRepositoryMock) GetByTitle(ctx context.Context, title string) (domain.Todo, error) {
	args := m.Called(ctx, title)

	return args.Get(0).(domain.Todo), args.Error(1)
}

func (m *TodoRepositoryMock) Store(_a0 context.Context, _a1 *domain.Todo) error {
	args := m.Called(_a0, _a1)

	return args.Error(0)
}

func (m *TodoRepositoryMock) Update(ctx context.Context, ar *domain.Todo) error {
	args := m.Called(ctx, ar)

	return args.Error(0)
}
