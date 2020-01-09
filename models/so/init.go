package models

import (
	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(
		new(Answer),
		new(Question),
		new(Tag),
		new(Correlation),
	)
}

/*
* Table Names
* */

func TNAnswer() string {
	return "so_answers"
}

func TNQuestion() string {
	return "so_questions"
}

func TNTag() string {
	return "so_tags"
}

func TNCorrelation() string {
	return "so_correlations"
}
