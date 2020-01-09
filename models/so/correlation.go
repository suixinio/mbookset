package models

// Correlation types.
const (
	CorrelationQuestionTag = iota
)

// Correlation model.
//   id1(question_id) - id2(tag_id)
type Correlation struct {
	Model

	//ID1  uint64 `json:"id1"`
	ID1  uint64 `orm:"column(id1);null" json:"id1"`
	ID2  uint64 `orm:"column(id2);null" json:"id2"`
	Str1 string `orm:"column(str1);null;size(255)" json:"str1"`
	Str2 string `orm:"column(str2);null;size(255)" json:"str2"`
	Str3 string `orm:"column(str3);null;size(255)" json:"str3"`
	Str4 string `orm:"column(str4);null;size(255)" json:"str4"`
	Int1 int    `orm:"column(int1);null" json:"int1"`
	Int2 int    `orm:"column(int2);null" json:"int2"`
	Int3 int    `orm:"column(int3);null" json:"int3"`
	Int4 int    `orm:"column(int4);null" json:"int4"`
	Type int    `orm:"column(type);null" json:"type"`
}

func (m *Correlation) TableName() string {
	return TNCorrelation()
}

func NewCorrelation() *Correlation {
	return &Correlation{}
}
