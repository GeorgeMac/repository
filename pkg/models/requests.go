package models

type RepositoryRequest struct {
	Count  int
	Unique bool
}

func NewRepositoryRequest(opts ...Option) RepositoryRequest {
	r := RepositoryRequest{Count: 1}

	Options(opts).Apply(&r)

	return r
}

type Option func(r *RepositoryRequest)

type Options []Option

func (o Options) Apply(r *RepositoryRequest) {
	for _, opt := range o {
		opt(r)
	}
}

func WithCount(count int) Option {
	return func(r *RepositoryRequest) {
		r.Count = count
	}
}

func Unique(r *RepositoryRequest) {
	r.Unique = true
}
