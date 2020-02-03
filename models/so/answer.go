package models

import (
	"github.com/astaxie/beego/orm"
	"strconv"
	"time"
)

type Answer struct {
	Model

	QuestionID  uint64 `orm:"column(question_id);index;null" json:"questionID"`
	Votes       int    `orm:"column(votes);null" json:"votes"`
	ContentEnUS string `orm:"column(content_en_us);null;type(text)" json:"contentEnUS"`
	ContentZhCN string `orm:"column(content_zh_cn);null;type(text)" json:"contentZhCN"`
	Path        string `orm:"column(path);null;type(text)" json:"path"`
	Source      int    `orm:"column(source);null;index" json:"source"`
	SourceID    string `orm:"unique;column(source_id);null;size(255)" sql:"index"`
	SourceURL   string `orm:"column(source_url);null;size(255);index" json:"sourceURL"`
	AuthorName  string `orm:"column(author_name);null" json:"authorName"`
	AuthorURL   string `orm:"column(author_url);null;size(255)" json:"authorURL"`

	//QuestionID  uint64 `sql:"index" json:"questionID"`
	//Votes       int    `json:"votes"`
	//ContentEnUS string `gorm:"type:mediumtext" json:"contentEnUS"`
	//ContentZhCN string `gorm:"type:mediumtext" json:"contentZhCN"`
	//Path        string `gorm:"type:text" json:"path"`
	//Source      int    `sql:"index" json:"source"`
	//SourceID    string `gorm:"size:255" sql:"index"`
	//SourceURL   string `gorm:"size:255" sql:"index" json:"sourceURL"`
	//AuthorName  string `json:"authorName"`
	//AuthorURL   string `gorm:"size:255" json:"authorURL"`
}

func (m *Answer) TableName() string {
	return TNAnswer()
}

func NewAnswer() *Answer {
	return &Answer{}
}

func (m *Answer) GetAnswers(questionID uint64) (ret *[]Answer) {
	o := orm.NewOrm()
	query := o.QueryTable(TNAnswer())
	var qa []Answer
	_, err := query.Filter("question_id", questionID).Limit(50).All(&qa)
	ret = &qa
	if nil != err {
		return
	}
	return
	//if err := db.Model(&model.Answer{}).Where("`question_id` = ?", questionID).Find(&ret).Error; nil != err {
	//	logger.Errorf("get answers of question [id=%d] failed: %s", questionID, err)
	//
	//	return
	//}

}

func (m *Answer) AddAnswers(ans []*Answer, qu *Question) {
	o := orm.NewOrm()
	for _, an := range ans {
		var aa Answer
		o.QueryTable(TNAnswer()).Filter("SourceID", an.SourceID).One(&aa)
		if aa.ID == 0 {
			an.ID = uint64(time.Now().UnixNano())
			an.QuestionID = qu.ID
			o.Insert(an)
		}
	}
}

func (m *Answer) GetUntranslatedAnswers(p int) (ans []*Answer) {
	sql := "select * from " + TNAnswer() + " where content_zh_cn = '' or content_zh_cn is null limit " + strconv.Itoa(p)
	o := orm.NewOrm()
	o.Raw(sql).QueryRows(&ans)
	return
}

//func (m *Answer) AddAns(qnas []*QnA) {
//
//	for _, qna := range qnas {
//		for _, answer := range qna.Answers {
//			answer.QuestionID = qna.Question.ID
//			answer.Insertorupdate()
//		}
//	}
//	//m.AddAnswers(qna)
//	//if err = srv.add(tx, qna); nil != err {
//	//	return
//	//}
//	return
//}
func (m *Answer) Insertorupdate() (err error) {
	//SourceID
	o := orm.NewOrm()
	var aa Answer
	o.QueryTable(TNAnswer()).Filter("SourceID", m.SourceID).One(&aa)
	if aa.ID > 0 {
		aa.Votes = m.Votes
		o.Update(&aa)
	} else {
		m.ID = uint64(time.Now().UnixNano())
		o.Insert(m)
	}
	return
}

func (m *Answer) Update() (err error) {
	o := orm.NewOrm()
	_, err = o.Update(m)
	return
}
