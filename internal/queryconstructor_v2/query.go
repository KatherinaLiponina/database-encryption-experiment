package queryconstructorv2

type Query interface {
	GetQuery() string
	GetArguments() []any
}

type exec struct {
	query string
}

func (q *exec) GetQuery() string {
	return q.query
}
func (q *exec) GetArguments() []any {
	return nil
}

func NewExecQuery(query string) Query {
	return &exec{query: query}
}

type insert struct {
	query string
	args  []any
}

func (q *insert) GetQuery() string {
	return q.query
}
func (q *insert) GetArguments() []any {
	return q.args
}
func NewInsertQuery(query string, args []any) Query {
	return &insert{query: query, args: args}
}

type selectQuery struct {
	query string
	args  []any
}

func (q *selectQuery) GetQuery() string {
	return q.query
}
func (q *selectQuery) GetArguments() []any {
	return q.args
}
func NewSelectQuery(query string, args []any) Query {
	return &selectQuery{query: query, args: args}
}
