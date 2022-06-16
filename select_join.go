package crate

type Join interface {
	run(args *[]any) string
}

type joinTemplate struct {
	Table     string
	Condition Condition
}

type InnerJoin struct {
	joinTemplate
}

func (j *InnerJoin) run(args *[]any) string {
	return "INNER JOIN " + j.Table + " ON " + j.Condition.run(args)
}

type OuterJoin struct {
	joinTemplate
}

func (j *OuterJoin) run(args *[]any) string {
	return "OUTER JOIN " + j.Table + " ON " + j.Condition.run(args)
}

type LeftJoin struct {
	joinTemplate
}

func (j *LeftJoin) run(args *[]any) string {
	return "LEFT JOIN " + j.Table + " ON " + j.Condition.run(args)
}

type RightJoin struct {
	joinTemplate
}

func (j *RightJoin) run(args *[]any) string {
	return "RIGHT JOIN " + j.Table + " ON " + j.Condition.run(args)
}
