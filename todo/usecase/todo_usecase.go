package usecase

import (
	"context"
	"time"

	"github.com/bxcodec/go-clean-arch/domain"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type todoUsecase struct {
	activity       domain.ActivityRepository
	todo           domain.TodoRepository
	contextTimeout time.Duration
}

func NewTodoUsecase(a domain.ActivityRepository, td domain.TodoRepository, timeout time.Duration) domain.TodoUsecase {
	return &todoUsecase{
		activity:       a,
		todo:           td,
		contextTimeout: timeout,
	}
}
func (a *todoUsecase) fillAuthorDetails(c context.Context, data []domain.Todo) ([]domain.Todo, error) {
	g, ctx := errgroup.WithContext(c)

	// Get the author's id
	mapActivities := map[int64]domain.Activity{}

	for _, todo := range data {
		mapActivities[todo.ActivityGroupID.ID] = domain.Activity{}
	}
	// Using goroutine to fetch the author's detail
	chanActivity := make(chan domain.Activity)
	for activityID := range mapActivities {
		activityID := activityID
		g.Go(func() error {
			res, err := a.activity.GetByID(ctx, activityID)
			if err != nil {
				return err
			}
			chanActivity <- res
			return nil
		})
	}

	go func() {
		err := g.Wait()
		if err != nil {
			logrus.Error(err)
			return
		}
		close(chanActivity)
	}()

	for activity := range chanActivity {
		if activity != (domain.Activity{}) {
			mapActivities[activity.ID] = activity
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// merge the author's data
	for index, item := range data {
		if a, ok := mapActivities[item.ActivityGroupID.ID]; ok {
			data[index].ActivityGroupID = a
		}
	}
	return data, nil
}

func (a *todoUsecase) Fetch(c context.Context, cursor string, num int64) (res []domain.Todo, nextCursor string, err error) {
	if num == 0 {
		num = 10
	}

	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()

	res, nextCursor, err = a.todo.Fetch(ctx, cursor, num)
	if err != nil {
		return nil, "", err
	}

	res, err = a.fillAuthorDetails(ctx, res)
	if err != nil {
		nextCursor = ""
	}
	return
}

func (a *todoUsecase) GetByID(c context.Context, id int64) (res domain.Todo, err error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()

	res, err = a.todo.GetByID(ctx, id)
	if err != nil {
		return
	}

	resActivity, err := a.activity.GetByID(ctx, res.ActivityGroupID.ID)
	if err != nil {
		return domain.Todo{}, err
	}
	res.ActivityGroupID = resActivity
	return
}

func (a *todoUsecase) Update(c context.Context, ar *domain.Todo) (err error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()

	ar.UpdatedAt = time.Now()
	return a.todo.Update(ctx, ar)
}

func (a *todoUsecase) GetByTitle(c context.Context, title string) (res domain.Todo, err error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()
	res, err = a.todo.GetByTitle(ctx, title)
	if err != nil {
		return
	}

	resActivity, err := a.activity.GetByID(ctx, res.ActivityGroupID.ID)
	if err != nil {
		return domain.Todo{}, err
	}

	res.ActivityGroupID = resActivity
	return
}

func (a *todoUsecase) Store(c context.Context, m *domain.Todo) (err error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()
	existedArticle, _ := a.GetByTitle(ctx, m.Title)
	if existedArticle != (domain.Todo{}) {
		return domain.ErrConflict
	}

	err = a.todo.Store(ctx, m)
	return
}

func (a *todoUsecase) Delete(c context.Context, id int64) (err error) {
	ctx, cancel := context.WithTimeout(c, a.contextTimeout)
	defer cancel()
	existedArticle, err := a.todo.GetByID(ctx, id)
	if err != nil {
		return
	}
	if existedArticle == (domain.Todo{}) {
		return domain.ErrNotFound
	}
	return a.todo.Delete(ctx, id)
}
