package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strings"
	"time"
)

// Tag model.
type Tag struct {
	Model

	Title         string `orm:"column(title);null;size(128)" json:"title"`
	QuestionCount int    `orm:"column(question_count);null" json:"questionCount"`
}

func (m *Tag) TableName() string {
	return TNTag()
}

func NewTag() *Tag {
	return &Tag{}
}

func (m *Tag) GetTagByTitle(title string) *Tag {
	o := orm.NewOrm()
	query := o.QueryTable(TNTag())
	tg := NewTag()
	err := query.Filter("title", title).One(tg)
	if nil != err {
		return nil
	}
	return tg
}

func (m *Tag) tagQuestion(o orm.Ormer, question *Question) (err error) {
	tags := strings.Split(question.Tags, ",")
	for _, tagTitle := range tags {
		tag := &Tag{}
		o.QueryTable(TNTag()).Filter("title", tagTitle).One(tag)
		if "" == tag.Title {
			tag.ID = uint64(time.Now().UnixNano())
			tag.Title = tagTitle
			tag.QuestionCount = 1
			_, err = o.Insert(tag)
		} else {
			tag.QuestionCount += 1
			_, err = o.Update(tag)
		}
		rel := &Correlation{}
		o.QueryTable(TNCorrelation()).Filter("id1", question.ID).Filter("id2", tag.ID).Filter("type", CorrelationQuestionTag).One(rel)
		if rel.ID == 0 {
			rel.ID = uint64(time.Now().UnixNano())
			rel.ID1 = question.ID
			rel.ID2 = tag.ID
			rel.Type = CorrelationQuestionTag
			o.Insert(rel)
		}
		logs.Info("标签 - 问题：%s - %s", tag.Title, question.TitleEnUS)
		//rel := Correlation{ID1: question.ID, ID2: tag.ID, Type: model.CorrelationQuestionTag}
	}
	//for _, tagTitle := range tags {
	//	tag := &Tag{}
	//	tx.Where("`title` = ?", tagTitle).First(tag)
	//	if "" == tag.Title {
	//		tag.Title = tagTitle
	//		tag.QuestionCount = 1
	//		if err := tx.Create(tag).Error; nil != err {
	//			return err
	//		}
	//	} else {
	//		tag.QuestionCount += 1
	//		if err := tx.Model(tag).Updates(tag).Error; nil != err {
	//			return err
	//		}
	//	}
	//
	//	rel := &model.Correlation{ID1: question.ID, ID2: tag.ID, Type: model.CorrelationQuestionTag}
	//	if err := tx.Create(rel).Error; nil != err {
	//		return err
	//	}
	//}

	return
}
