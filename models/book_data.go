package models

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"mbook/common"
	"mbook/conf"
	"time"

	"github.com/astaxie/beego/orm"
)

//拼接返回到接口的图书信息
type BookData struct {
	BookId         int       `json:"book_id"`
	BookName       string    `json:"book_name"`
	Identify       string    `json:"identify"`
	OrderIndex     int       `json:"order_index"`
	Description    string    `json:"description"`
	PrivatelyOwned int       `json:"privately_owned"`
	PrivateToken   string    `json:"private_token"`
	DocCount       int       `json:"doc_count"`
	CommentCount   int       `json:"comment_count"`
	CreateTime     time.Time `json:"create_time"`
	CreateName     string    `json:"create_name"`
	ModifyTime     time.Time `json:"modify_time"`
	Cover          string    `json:"cover"`
	MemberId       int       `json:"member_id"`
	Username       int       `json:"user_name"`
	Editor         string    `json:"editor"`
	RelationshipId int       `json:"relationship_id"`
	RoleId         int       `json:"role_id"`
	RoleName       string    `json:"role_name"`
	Status         int
	Vcnt           int    `json:"vcnt"`
	Collection     int    `json:"star"`
	Score          int    `json:"score"`
	CntComment     int    `json:"cnt_comment"`
	CntScore       int    `json:"cnt_score"`
	ScoreFloat     string `json:"score_float"`
	LastModifyText string `json:"last_modify_text"`
	Author         string `json:"author"`
	AuthorURL      string `json:"author_url"`
}

func NewBookData() *BookData {
	return &BookData{}
}

func (m *BookData) SelectByIdentify(identify string, memberId int) (result *BookData, err error) {
	if identify == "" || memberId <= 0 {
		return result, errors.New("Invalid parameter")
	}

	book := NewBook()
	o := orm.NewOrm()
	err = o.QueryTable(TNBook()).Filter("identify", identify).One(book)
	if err != nil {
		return
	}

	//查看权限
	relationship := NewRelationship()
	err = o.QueryTable(TNRelationship()).Filter("book_id", book.BookId).Filter("role_id", 0).One(relationship)
	if err != nil {
		return result, errors.New("Permission denied")
	}
	member, err := NewMember().Find(relationship.MemberId)
	if err != nil {
		return result, err
	}

	err = o.QueryTable(TNRelationship()).Filter("book_id", book.BookId).Filter("member_id", memberId).One(relationship)
	if err != nil {
		return
	}

	result = book.ToBookData()
	result.CreateName = member.Account
	result.MemberId = relationship.MemberId
	result.RoleId = relationship.RoleId
	result.RoleName = common.BookRole(result.RoleId)
	result.RelationshipId = relationship.RelationshipId

	document := NewDocument()
	err = o.QueryTable(TNDocuments()).Filter("book_id", book.BookId).OrderBy("modify_time").One(document)
	return
}

// 根据项目标识查询项目以及指定用户权限的信息.
func (m *BookData) FindByIdentify(identify string, memberId int) (result *BookData, err error) {
	if identify == "" || memberId <= 0 {
		beego.Error("FindByIdentify ->", err)
		return result, err
	}
	o := orm.NewOrm()

	book := NewBook()

	err = o.QueryTable(book.TableName()).Filter("identify", identify).One(book)
	if err != nil {
		return
	}

	relationship := NewRelationship()

	//err = o.QueryTable(relationship.TableNameWithPrefix()).Filter("book_id", book.BookId).Filter("member_id", memberId).One(relationship)
	err = o.QueryTable(relationship.TableName()).Filter("book_id", book.BookId).Filter("member_id", memberId).One(relationship)
	if err != nil {
		return
	}

	var relationship2 Relationship

	//err = o.QueryTable(relationship.TableNameWithPrefix()).Filter("book_id", book.BookId).Filter("role_id", conf.BookFounder).One(&relationship2)
	err = o.QueryTable(relationship.TableName()).Filter("book_id", book.BookId).Filter("role_id", conf.BookFounder).One(&relationship2)
	if err != nil {
		logs.Error("根据项目标识查询项目以及指定用户权限的信息 => ", err)
		return result, err
	}

	member, err := NewMember().Find(relationship2.MemberId)
	if err != nil {
		return result, err
	}

	result = book.ToBookResult()

	result.CreateName = member.Account
	result.MemberId = relationship.MemberId
	result.RoleId = relationship.RoleId
	result.RelationshipId = relationship.RelationshipId

	switch result.RoleId {
	case conf.BookFounder:
		result.RoleName = "创始人"
	case conf.BookAdmin:
		result.RoleName = "管理员"
	case conf.BookEditor:
		result.RoleName = "编辑者"
	case conf.BookObserver:
		result.RoleName = "观察者"
	}

	doc := NewDocument()

	err = o.QueryTable(doc.TableName()).Filter("book_id", book.BookId).OrderBy("modify_time").One(doc)
	if err == nil {
		member2 := NewMember()
		member2.Find(doc.ModifyAt)
		result.LastModifyText = member2.Account + " 于 " + doc.ModifyTime.Format("2006-01-02 15:04:05")
	}
	return
}
