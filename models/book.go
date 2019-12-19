package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"strconv"
	"strings"
)

type Book struct {
	BookId      int    `orm:"pk;auto"json:"book_id"`
	BookName    string `orm:"size(500)"json:"book_name"`
	Identify    string `orm:"size(100);unique"json:"identify"`
	OrderIndex  int    `orm:"default(0)"json:"order_index"`
	Description string `orm:"size(1000)"json:"description"`
}

func (m *Book) TableName() string {
	return TNBook()
}

func (m *Book) HomeData(pageIndex, pageSize int, cid int, fields ...string) (books []Book, totalCount int, err error) {
	if len(fields) == 0 {
		fields = append(fields, "book_id", "book_name", "identify", "conver", "order_index")
	}
	fieldStr := "b." + strings.Join(fields, ",b.")

	//select * from md_books b LEFT JOIN md_book_category c on b.book_id=c.	book_id WHERE c.category_id = 1;
	//todo 数据库联合查询的点
	sqlFmt := "select %v from " + TNBook() + " b LEFT JOIN " + TNBookCategory() + " c on b.book_id=c.book_id WHERE c.category_id = " + strconv.Itoa(cid)

	sql := fmt.Sprintf(sqlFmt, fieldStr)
	sqlCount := fmt.Sprintf(sqlFmt, "count(*) cnt")
	o := orm.NewOrm()
	var params []orm.Params
	if _, err := o.Raw(sqlCount).Values(&params); err == nil {
		if len(params) > 0 {
			totalCount, _ = strconv.Atoi(params[0]["cnt"].(string))
		}
	}
	_, err = o.Raw(sql).QueryRows(&books)
	return
}
