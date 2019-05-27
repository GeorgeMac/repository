package models

type RepositoriesRequest struct {
	Count  int
	Unique bool
}

func NewRepositoriesRequest(opts ...Option) RepositoriesRequest {
	r := RepositoriesRequest{Count: 1}

	Options(opts).Apply(&r)

	return r
}

type Option func(r *RepositoriesRequest)

type Options []Option

func (o Options) Apply(r *RepositoriesRequest) {
	for _, opt := range o {
		opt(r)
	}
}

func WithCount(count int) Option {
	return func(r *RepositoriesRequest) {
		r.Count = count
	}
}

func Unique(r *RepositoriesRequest) {
	r.Unique = true
}
