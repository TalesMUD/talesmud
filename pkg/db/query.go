package db

// QueryParam represents a single key-value query parameter.
type QueryParam struct {
	Key   string
	Value interface{}
}

// QueryParams holds a collection of query parameters for filtering.
type QueryParams struct {
	params []QueryParam
}

// NewQueryParams creates a new QueryParams instance with optional initial parameters.
func NewQueryParams(initial ...QueryParam) *QueryParams {
	return &QueryParams{params: initial}
}

// With adds a parameter and returns the QueryParams for chaining.
func (qp *QueryParams) With(param QueryParam) *QueryParams {
	qp.params = append(qp.params, param)
	return qp
}

// Params returns the underlying parameter slice.
func (qp *QueryParams) Params() []QueryParam {
	return qp.params
}
