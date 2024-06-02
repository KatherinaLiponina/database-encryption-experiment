package queryconstructorv2

import (
	"fmt"
)

type QueryBuilder interface {
	Next() (*Query, error)
}

var ErrNoMoreQueries = fmt.Errorf("no more queries to build")
