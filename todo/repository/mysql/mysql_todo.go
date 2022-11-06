package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/bxcodec/go-clean-arch/activity/repository"
	"github.com/bxcodec/go-clean-arch/domain"
)

type mysqlTodoRepository struct {
	Conn *sql.DB
}

// NewMysqlArticleRepository will create an object that represent the article.Repository interface
func NewMysqlTodoRepository(Conn *sql.DB) domain.TodoRepository {
	return &mysqlTodoRepository{Conn}
}

func (m *mysqlTodoRepository) fetch(ctx context.Context, query string, args ...interface{}) (result []domain.Todo, err error) {
	rows, err := m.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	defer func() {
		errRow := rows.Close()
		if errRow != nil {
			logrus.Error(errRow)
		}
	}()

	result = make([]domain.Todo, 0)
	for rows.Next() {
		t := domain.Todo{}
		activityID := int64(0)
		err = rows.Scan(
			&t.ID,
			&activityID,
			&t.Title,
			&t.IsActive,
			&t.Priority,
		)

		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		t.ActivityGroupID = domain.Activity{
			ID: activityID,
		}
		result = append(result, t)
	}

	return result, nil
}

func (m *mysqlTodoRepository) Fetch(ctx context.Context, cursor string, num int64) (res []domain.Todo, nextCursor string, err error) {
	query := `SELECT id, activity_group_id, title, is_active, priority
  						FROM todo WHERE created_at > ? ORDER BY created_at LIMIT ? `

	decodedCursor, err := repository.DecodeCursor(cursor)
	if err != nil && cursor != "" {
		return nil, "", domain.ErrBadParamInput
	}

	res, err = m.fetch(ctx, query, decodedCursor, num)
	if err != nil {
		return nil, "", err
	}

	if len(res) == int(num) {
		nextCursor = repository.EncodeCursor(res[len(res)-1].CreatedAt)
	}

	return
}
func (m *mysqlTodoRepository) GetByID(ctx context.Context, id int64) (res domain.Todo, err error) {
	query := `SELECT id, activity_group_id, title, is_active, priority
  						FROM todo WHERE ID = ?`

	list, err := m.fetch(ctx, query, id)
	if err != nil {
		return domain.Todo{}, err
	}

	if len(list) > 0 {
		res = list[0]
	} else {
		return res, domain.ErrNotFound
	}

	return
}

func (m *mysqlTodoRepository) GetByTitle(ctx context.Context, title string) (res domain.Todo, err error) {
	query := `SELECT id, activity_group_id, title, is_active, priority
  						FROM todo WHERE title = ?`

	list, err := m.fetch(ctx, query, title)
	if err != nil {
		return
	}

	if len(list) > 0 {
		res = list[0]
	} else {
		return res, domain.ErrNotFound
	}
	return
}

func (m *mysqlTodoRepository) Store(ctx context.Context, a *domain.Todo) (err error) {
	query := `INSERT todo SET activity_group_id=?, title=?, is_active=?, priority=?, updated_at=?, created_at=?`
	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, a.ActivityGroupID, a.Title, a.IsActive, a.Priority, a.UpdatedAt, a.CreatedAt)
	if err != nil {
		return
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return
	}
	a.ID = lastID
	return
}

func (m *mysqlTodoRepository) Delete(ctx context.Context, id int64) (err error) {
	query := "DELETE FROM todo WHERE id = ?"

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return
	}

	rowsAfected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if rowsAfected != 1 {
		err = fmt.Errorf("weird  Behavior. Total Affected: %d", rowsAfected)
		return
	}

	return
}
func (m *mysqlTodoRepository) Update(ctx context.Context, ar *domain.Todo) (err error) {
	query := `UPDATE activity set email=?, title=?, updated_at=? WHERE ID = ?`

	stmt, err := m.Conn.PrepareContext(ctx, query)
	if err != nil {
		return
	}

	res, err := stmt.ExecContext(ctx, ar.ActivityGroupID, ar.Title, ar.IsActive, ar.Priority, ar.UpdatedAt, ar.ID)
	if err != nil {
		return
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return
	}
	if affect != 1 {
		err = fmt.Errorf("weird  Behavior. Total Affected: %d", affect)
		return
	}

	return
}
