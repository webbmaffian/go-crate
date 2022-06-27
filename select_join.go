package crate

type Join interface {
	run(args *[]any) string
}

type InnerJoin struct {
	Table     string
	Condition Condition
}

func (j *InnerJoin) run(args *[]any) string {
	return "INNER JOIN " + j.Table + " ON " + j.Condition.run(args)
}

type OuterJoin InnerJoin

func (j *OuterJoin) run(args *[]any) string {
	return "OUTER JOIN " + j.Table + " ON " + j.Condition.run(args)
}

type LeftJoin InnerJoin

func (j *LeftJoin) run(args *[]any) string {
	return "LEFT JOIN " + j.Table + " ON " + j.Condition.run(args)
}

type RightJoin InnerJoin

func (j *RightJoin) run(args *[]any) string {
	return "RIGHT JOIN " + j.Table + " ON " + j.Condition.run(args)
}
