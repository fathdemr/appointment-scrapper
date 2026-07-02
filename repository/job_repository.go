package repository

import (
	"context"
	"errors"

	"appointment-scrapper/model"
)

var ErrNotFound = errors.New("job not found")

// JobRepository booking job'ları için veri erişim arayüzü.
type JobRepository interface {
	Create(ctx context.Context, job *model.BookingJob) error
	GetByID(ctx context.Context, id string) (*model.BookingJob, error)
	List(ctx context.Context) ([]*model.BookingJob, error)
	ListByStatus(ctx context.Context, status model.JobStatus) ([]*model.BookingJob, error)
	Update(ctx context.Context, job *model.BookingJob) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status model.JobStatus) error
	IncrementRunStats(ctx context.Context, id string, found bool, result string) error
}
