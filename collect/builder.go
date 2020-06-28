package collect

type Queryer struct {
	filter             Filter
	skipIfTypeNotMatch bool
	notExistStrategy   NotExistStrategy
}

func (q *Queryer) AddFilter(field, typ, op string, value interface{}) *Queryer {
	q.filter.Add(field, typ, op, value)
	return q
}

func (q *Queryer) SkipIfTypeNotMatch() *Queryer {
	q.skipIfTypeNotMatch = true
	return q
}

func (q *Queryer) NotExistStrategy(strategy NotExistStrategy) *Queryer {
	q.notExistStrategy = strategy
	return q
}

func (q *Queryer) Field(fieldName string, fieldNotExistStrategy NotExistStrategy) *ColumnQueryer {
	return &ColumnQueryer{
		Queryer:               *q,
		fieldName:             fieldName,
		fieldNotExistStrategy: fieldNotExistStrategy,
	}
}

func (q *Queryer) Count(value interface{}) (int, error) {
	var count int
	err := Select(value, &q.filter, q.skipIfTypeNotMatch, q.notExistStrategy, Count(&count))
	return count, err
}

func (q *Queryer) Run(value interface{}) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	err := Select(value, &q.filter, q.skipIfTypeNotMatch, q.notExistStrategy, func(idx Location, row map[string]interface{}) error {
		results = append(results, row)
		return nil
	})
	return results, err
}

type ColumnQueryer struct {
	Queryer

	fieldName             string
	fieldNotExistStrategy NotExistStrategy
}

func (q *ColumnQueryer) Aggregate(t, operator string) *AggregateQueryer {
	aggregateFactory, err := CreateAggregateFactory(t, operator)
	return &AggregateQueryer{
		ColumnQueryer:    *q,
		aggregateFactory: aggregateFactory,
		err:              err,
	}
}

func (q *ColumnQueryer) Run(value interface{}) ([]map[string]interface{}, []interface{}, error) {
	var rows []map[string]interface{}
	var results []interface{}

	err := Select(value, &q.filter, q.skipIfTypeNotMatch, q.notExistStrategy, Map(Field(q.fieldName, q.fieldNotExistStrategy),
		func(loc Location, row map[string]interface{}, fieldValue interface{}) error {
			rows = append(rows, row)
			results = append(results, fieldValue)
			return nil
		}))
	return rows, results, err
}

type AggregateQueryer struct {
	ColumnQueryer

	aggregateFactory AggregateFactory
	err              error
}

func (q *AggregateQueryer) Run(value interface{}) ([]map[string]interface{}, interface{}, error) {
	if q.err != nil {
		return nil, nil, q.err
	}
	agg := q.aggregateFactory.Create(8)
	err := Select(value, &q.filter, q.skipIfTypeNotMatch, q.notExistStrategy,
		Map(Field(q.fieldName, q.fieldNotExistStrategy), Aggregate(agg)))
	if err != nil {
		return nil, nil, err
	}
	return agg.Result()
}

// func (q *Queryer) Aggregate(typ, op string) *Queryer {
// 	f, err := CreateAggregateFactory(typ, op)
// 	if err != nil {
// 		panic(err)
// 	}
// 	q.aggregateFactory = f
// 	return q
// }
