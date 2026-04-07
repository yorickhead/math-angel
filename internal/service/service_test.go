package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/osamikoyo/math-angel/internal/errors"
	"github.com/osamikoyo/math-angel/internal/mocks"
	"github.com/osamikoyo/math-angel/internal/model"
)

func TestNewService(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	timeout := time.Second * 5

	service := NewService(mockRepo, mockCash, timeout)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
	assert.Equal(t, mockCash, service.cash)
	assert.Equal(t, timeout, service.timeout)
}

func TestCreateTask(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	taskType := "algebra"
	problem := "2+2=?"
	solution := "4"
	boxed := "4"
	level := "easy"

	mockCash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*model.Task")).Return(nil)
	mockRepo.On("CreateTask", mock.Anything, mock.AnythingOfType("*model.Task")).Return(nil)

	err := service.CreateTask(ctx, taskType, problem, solution, boxed, level)

	assert.NoError(t, err)
	mockCash.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestIncLike(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	id := uuid.New().String()
	task := &model.Task{ID: uuid.MustParse(id), Likes: 5}

	mockRepo.On("GetTask", mock.Anything, uuid.MustParse(id)).Return(task, nil)
	mockCash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), task).Return(nil)
	mockRepo.On("UpdateTask", mock.Anything, uuid.MustParse(id), "likes", uint(6)).Return(nil)

	err := service.IncLike(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, uint(6), task.Likes)
	mockCash.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestIncLike_InvalidUUID(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	invalidID := "invalid-uuid"

	err := service.IncLike(ctx, invalidID)

	assert.Equal(t, errors.ErrBadUID, err)
}

func TestDecLike(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	id := uuid.New().String()
	task := &model.Task{ID: uuid.MustParse(id), Likes: 5}

	mockRepo.On("GetTask", mock.Anything, uuid.MustParse(id)).Return(task, nil)
	mockCash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), task).Return(nil)
	mockRepo.On("UpdateTask", mock.Anything, uuid.MustParse(id), "likes", uint(4)).Return(nil)

	err := service.DecLike(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, uint(4), task.Likes)
	mockCash.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestIncDislike(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	id := uuid.New().String()
	task := &model.Task{ID: uuid.MustParse(id), Dislikes: 3}

	mockRepo.On("GetTask", mock.Anything, uuid.MustParse(id)).Return(task, nil)
	mockCash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), task).Return(nil)
	mockRepo.On("UpdateTask", mock.Anything, uuid.MustParse(id), "dislikes", uint(4)).Return(nil)

	err := service.IncDislike(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, uint(4), task.Dislikes)
	mockCash.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestDecDislike(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	id := uuid.New().String()
	task := &model.Task{ID: uuid.MustParse(id), Dislikes: 3}

	mockRepo.On("GetTask", mock.Anything, uuid.MustParse(id)).Return(task, nil)
	mockCash.On("SetTask", mock.Anything, mock.AnythingOfType("string"), task).Return(nil)
	mockRepo.On("UpdateTask", mock.Anything, uuid.MustParse(id), "dislikes", uint(2)).Return(nil)

	err := service.DecDislike(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, uint(2), task.Dislikes)
	mockCash.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestGetRandomTask_FromCache(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	taskType := "algebra"
	level := "easy"
	tasks := []model.Task{
		{ID: uuid.New(), Type: taskType, Level: level, Likes: 1},
		{ID: uuid.New(), Type: taskType, Level: level, Likes: 2},
	}

	mockCash.On("GetTasks", mock.Anything, "algebra:easy").Return(tasks, nil)

	task, err := service.GetRandomTask(ctx, taskType, level)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Contains(t, tasks, *task)
	mockCash.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetTasksByTypeAndLevel")
}

func TestGetRandomTask_FromRepo(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	taskType := "algebra"
	level := "easy"
	tasks := []model.Task{
		{ID: uuid.New(), Type: taskType, Level: level, Likes: 1},
	}

	mockCash.On("GetTasks", mock.Anything, "algebra:easy").Return(nil, assert.AnError)
	mockRepo.On("GetTasksByTypeAndLevel", mock.Anything, taskType, level).Return(tasks, nil)
	mockCash.On("SetTasks", mock.Anything, "algebra:easy", tasks).Return(nil)

	task, err := service.GetRandomTask(ctx, taskType, level)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, tasks[0], *task)
	mockCash.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestGetTask_FromCache(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	id := uuid.New().String()
	task := &model.Task{ID: uuid.MustParse(id)}

	mockCash.On("GetTask", mock.Anything, "one:"+id).Return(task, nil)

	result, err := service.GetTask(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, task, result)
	mockCash.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetTask")
}

func TestGetTask_FromRepo(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	id := uuid.New().String()
	task := &model.Task{ID: uuid.MustParse(id)}

	mockCash.On("GetTask", mock.Anything, "one:"+id).Return(nil, assert.AnError)
	mockRepo.On("GetTask", mock.Anything, uuid.MustParse(id)).Return(task, nil)
	mockCash.On("SetTask", mock.Anything, "one:"+id, task).Return(nil)

	result, err := service.GetTask(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, task, result)
	mockCash.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestGetTask_InvalidUUID(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	invalidID := "invalid"

	mockCash.On("GetTask", mock.Anything, "one:invalid").Return(nil, assert.AnError)

	_, err := service.GetTask(ctx, invalidID)

	assert.Equal(t, errors.ErrBadUID, err)
}

func TestGetBests_FromCache(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	taskType := "algebra"
	level := "easy"
	tasks := []model.Task{
		{ID: uuid.New(), Likes: 3},
		{ID: uuid.New(), Likes: 1},
		{ID: uuid.New(), Likes: 2},
	}
	sortedTasks := []model.Task{tasks[1], tasks[2], tasks[0]} // sorted by likes

	mockCash.On("GetTasks", mock.Anything, "sorted:algebra:easy").Return(sortedTasks, nil)

	result, err := service.GetBests(ctx, taskType, level, 2, 1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, sortedTasks[0], result[0])
	assert.Equal(t, sortedTasks[1], result[1])
	mockCash.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetTasksByTypeAndLevel")
}

func TestGetBests_FromRepo(t *testing.T) {
	mockRepo := &mocks.Repository{}
	mockCash := &mocks.Cash{}
	service := NewService(mockRepo, mockCash, time.Second)

	ctx := context.Background()
	taskType := "algebra"
	level := "easy"
	tasks := []model.Task{
		{ID: uuid.New(), Likes: 1},
		{ID: uuid.New(), Likes: 3},
		{ID: uuid.New(), Likes: 2},
	}

	mockCash.On("GetTasks", mock.Anything, "sorted:algebra:easy").Return(nil, assert.AnError)
	mockCash.On("GetTasks", mock.Anything, "algebra:easy").Return(nil, assert.AnError)
	mockRepo.On("GetTasksByTypeAndLevel", mock.Anything, taskType, level).Return(tasks, nil)
	mockCash.On("SetTasks", mock.Anything, "algebra:easy", mock.AnythingOfType("[]model.Task")).Return(nil)

	result, err := service.GetBests(ctx, taskType, level, 2, 1)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	// Since sorted, likes should be 1,2,3 but we take first 2: 1 and 2
	assert.Equal(t, uint(1), result[0].Likes)
	assert.Equal(t, uint(2), result[1].Likes)
	mockCash.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
