package models

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"html/template"
	"mbook/conf"
	"mbook/utils"

	"strconv"
	"time"
)

type tag struct {
	Title string
}

type contributor struct {
	Name           string
	Avatar         string
	ContriCount    int
	ContriDistance int
}

type question struct {
	ID           uint64
	Path         string
	Title        string
	Description  string
	Tags         []*tag
	Content      template.HTML
	SourceURL    string
	Contributors []*contributor
}

// Question model.
type Question struct {
	Model

	TitleEnUS   string `orm:"column(title_en_us);null;type(text)" json:"titleEnUS"`
	TitleZhCN   string `orm:"column(title_zh_cn);null;type(text)" json:"titleZhCN"`
	Tags        string `orm:"column(tags);null;type(text)" json:"tags"`
	Votes       int    `orm:"column(votes);null" json:"votes"`
	Views       int    `orm:"column(views);null" json:"views"`
	ContentEnUS string `orm:"column(content_en_us);null;type(text)" json:"contentEnUS"`
	ContentZhCN string `orm:"column(content_zh_cn);null;type(text)" json:"contentZhCN"`
	Path        string `orm:"column(path);null;type(text)" json:"path"`
	Source      int    `orm:"column(source);null;index" json:"source"`
	SourceID    string `orm:"unique;column(source_id);null;size(255);index" json:"sourceID"`
	SourceURL   string `orm:"column(source_url);null;size(255);index" json:"sourceURL"`
	AuthorName  string `orm:"column(author_name);null" json:"authorName"`
	AuthorURL   string `orm:"column(author_url);null;size(255)" json:"authorURL"`
}

// Sources.
const (
	SourceStackOverflow = iota
)

// Data types
const (
	DataTypeQuestion = iota
	DataTypeAnswer
)

func (m *Question) TableName() string {
	return TNQuestion()
}

func NewQuestion() *Question {
	return &Question{}
}

func (srv *Question) GetQuestions(page int) (ret []*Question, pagination *utils.Pagination) {
	offset := (page - 1) * conf.OSPageSize
	o := orm.NewOrm()

	sqlFmt := "select id ,source_id,created_at , title_zh_cn,tags,path FROM " + TNQuestion() + " where title_zh_cn !='' and content_zh_cn !='' ORDER BY updated_at,votes ASC LIMIT %v OFFSET %v"
	sql := fmt.Sprintf(sqlFmt, conf.OSPageSize, offset)
	o.Raw(sql).QueryRows(&ret);


	sqlCount := "select count(id) cnt FROM " + TNQuestion() + " where title_zh_cn !='' and content_zh_cn !=''"
	var params []orm.Params
	totalCount := 0
	if _, err := o.Raw(sqlCount).Values(&params); err == nil {
		if len(params) > 0 {
			totalCount, _ = strconv.Atoi(params[0]["cnt"].(string))
		}
	}

	pagination = utils.NewPagination(page, int(totalCount))

	return
}

func (srv *Question) GetQuestionByQuestionID(qa_id string) (ret *Question) {
	o := orm.NewOrm()
	query := o.QueryTable(TNQuestion())
	var qa Question
	_ = query.Filter("SourceID", qa_id).One(&qa)
	ret = &qa
	return
}

func (src *Question) GetTagTranslatedQuestions(tagID uint64, page int) (ret []*Question, pagination *utils.Pagination) {
	//var rels []*Correlation
	o := orm.NewOrm()

	offset := (page - 1) * conf.PageSize
	sql := "select %s from " + TNQuestion() + " where (title_zh_cn !='' and content_zh_cn != '') and id in (select id1 from " + TNCorrelation() + " where id2=%v and type=%v)"
	sqlCount := fmt.Sprintf(sql, "count(id) cnt", tagID, CorrelationQuestionTag)
	var params []orm.Params
	totalCount := 0
	if _, err := o.Raw(sqlCount).Values(&params); err == nil {
		if len(params) > 0 {
			totalCount, _ = strconv.Atoi(params[0]["cnt"].(string))
		}
	}
	sql = "select %s from " + TNQuestion() + " where (title_zh_cn !='' and content_zh_cn != '') and id in (select id1 from " + TNCorrelation() + " where id2=%v and type=%v) ORDER BY created_at desc limit %v OFFSET %v"
	sqlFmt := fmt.Sprintf(sql, "*", tagID, CorrelationQuestionTag, conf.PageSize, offset)
	o.Raw(sqlFmt).QueryRows(&ret)

	//sqlCount := fmt.Sprintf(sqlFmt, "count(book_id) cnt")
	//var params []orm.Params
	//if _, err := o.Raw(sqlCount).Values(&params); err == nil {
	//	if len(params) > 0 {
	//		totalCount, _ = strconv.Atoi(params[0]["cnt"].(string))
	//	}
	//}
	//_, err := o.QueryTable(TNCorrelation()).Filter("id2", tagID).Filter("type", CorrelationQuestionTag).All(&rels)
	//if nil != err {
	//	return
	//}
	//var questionIDs []uint64
	//for _, questionTagRel := range rels {
	//	questionIDs = append(questionIDs, questionTagRel.ID1)
	//}

	//count, err := o.QueryTable(TNQuestion()).Filter("id__in", questionIDs).Count()

	pagination = utils.NewPagination(page, totalCount)
	//_, err = o.QueryTable(TNQuestion()).Filter("id__in", questionIDs).OrderBy("-created_at").Offset(offset).Limit(conf.PageSize).All(&ret)

	return

}

func (src *Question) AddQuestions(qus []*Question) (err error) {
	o := orm.NewOrm()
	for _, qu := range qus {
		// 开启事务
		o.Begin()
		var qa Question
		o.QueryTable(TNQuestion()).Filter("SourceID", qu.SourceID).One(&qa)
		if qa.ID > 0 {
			//qu.ID = qa.ID
			qa.Votes = qu.Votes
			qa.Views = qu.Views
			_, err = o.Update(&qa)
			qu = &qa
			//if nil != err {
			//	logs.Error("回滚1 %s %s", err, qa)
			//	o.Rollback()
			//
			//}
		} else {
			qu.ID = uint64(time.Now().UnixNano())
			_, err = o.Insert(qu)
		}
		NewTag().tagQuestion(o, qu)
		if nil != err {
			logs.Error("回滚1%s %s", err, qu)
			o.Rollback()
		}
		err = o.Commit()
		if nil != err {
			logs.Error("回滚2%s %s", err, qu)
			o.Rollback()
		}
	}

	return
}

func (stc *Question) GetUntranslatedQuestions() (qus []*Question) {
	sql := "select * from " + TNQuestion() + " where title_zh_cn ='' or title_zh_cn is null or content_zh_cn = '' or content_zh_cn is null"
	o := orm.NewOrm()
	o.Raw(sql).QueryRows(&qus)
	return
}

func (m *Question) GetUnansweredQuestions(p int) (qus []*Question) {
	sql := "SELECT * from " + TNQuestion() + " WHERE id not in (SELECT question_id from " + TNAnswer() + ") ORDER BY votes desc limit " + strconv.Itoa(p)
	o := orm.NewOrm()
	o.Raw(sql).QueryRows(&qus)
	return
}

func (m *Question) GetansweredUntranslatedQuestions(p int) (qus []*Question) {
	sql := "SELECT * from " + TNQuestion() + " WHERE id in (SELECT question_id from " + TNAnswer() + ") and (title_zh_cn ='' or title_zh_cn is null or content_zh_cn = '' or content_zh_cn is null) ORDER BY votes desc limit " + strconv.Itoa(p)
	o := orm.NewOrm()
	o.Raw(sql).QueryRows(&qus)
	return
}

func (str *Question) Update() (err error) {
	o := orm.NewOrm()
	_, err = o.Update(str)
	return
}
