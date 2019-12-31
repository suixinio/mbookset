package models

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"mbook/utils"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
)

//图书章节内容
type Document struct {
	DocumentId   int           `orm:"pk;auto;column(document_id)" json:"doc_id"`
	DocumentName string        `orm:"column(document_name);size(500)" json:"doc_name"`
	Identify     string        `orm:"column(identify);size(100);index;null;default(null)" json:"identify"`
	BookId       int           `orm:"column(book_id);type(int)" json:"book_id"`
	ParentId     int           `orm:"column(parent_id);type(int);default(0)" json:"parent_id"`
	OrderSort    int           `orm:"column(order_sort);default(0);type(int)" json:"order_sort"`
	Release      string        `orm:"column(release);type(text);null" json:"release"`
	CreateTime   time.Time     `orm:"column(create_time);type(datetime);auto_now_add" json:"create_time"`
	MemberId     int           `orm:"column(member_id);type(int)" json:"member_id"`
	ModifyTime   time.Time     `orm:"column(modify_time);type(datetime);default(null);auto_now" json:"modify_time"`
	ModifyAt     int           `orm:"column(modify_at);type(int)" json:"-"`
	Version      int64         `orm:"type(bigint);column(version)" json:"version"`
	AttachList   []*Attachment `orm:"-" json:"attach"`
	Vcnt         int           `orm:"column(vcnt);default(0)" json:"vcnt"`
	Markdown     string        `orm:"-" json:"markdown"`
}

func (m *Document) TableName() string {
	return TNDocuments()
}

func NewDocument() *Document {
	return &Document{
		Version: time.Now().Unix(),
	}
}

//根据文档ID查询指定文档
func (m *Document) SelectByDocId(id int) (doc *Document, err error) {
	if id <= 0 {
		return m, errors.New("Invalid parameter")
	}

	o := orm.NewOrm()
	err = o.QueryTable(m.TableName()).Filter("document_id", id).One(m)
	if err == orm.ErrNoRows {
		return m, errors.New("数据不存在")
	}

	return m, nil
}

//根据指定字段查询一条文档
func (m *Document) SelectByIdentify(BookId, Identify interface{}) (*Document, error) {
	err := orm.NewOrm().QueryTable(m.TableName()).Filter("BookId", BookId).Filter("Identify", Identify).One(m)
	return m, err
}

//插入和更新文档
func (m *Document) InsertOrUpdate(cols ...string) (id int64, err error) {
	o := orm.NewOrm()
	id = int64(m.DocumentId)
	m.ModifyTime = time.Now()
	m.DocumentName = strings.TrimSpace(m.DocumentName)
	if m.DocumentId > 0 { //文档id存在，则更新
		_, err = o.Update(m, cols...)
		return
	}

	var selectDocument Document
	//直接查询一个字段
	o.QueryTable(TNDocuments()).Filter("identify", m.Identify).Filter("book_id", m.BookId).One(&selectDocument, "document_id")
	if selectDocument.DocumentId == 0 {
		m.CreateTime = time.Now()
		id, err = o.Insert(m)
		NewBook().RefreshDocumentCount(m.BookId)
	} else { //identify存在，则执行更新
		_, err = o.Update(m)
		id = int64(selectDocument.DocumentId)
	}
	return
}

//删除文档及其子文档
func (m *Document) Delete(docId int) error {

	o := orm.NewOrm()
	modelStore := new(DocumentStore)

	if doc, err := m.SelectByDocId(docId); err == nil {
		o.Delete(doc)
		modelStore.Delete(docId)
	}

	var docs []*Document

	_, err := o.QueryTable(m.TableName()).Filter("parent_id", docId).All(&docs)
	if err != nil {
		return err
	}

	for _, item := range docs {
		docId := item.DocumentId
		o.QueryTable(m.TableName()).Filter("document_id", docId).Delete()
		//删除document_store表对应的文档
		modelStore.Delete(docId)
		m.Delete(docId)
	}
	return nil
}

//发布文档内容
func (m *Document) ReleaseContent(bookId int, baseUrl string) {
	// 防止多处重复发布 ,Lock
	utils.BooksRelease.Set(bookId)
	defer utils.BooksRelease.Delete(bookId)

	o := orm.NewOrm()
	var book Book
	querySeter := o.QueryTable(TNBook()).Filter("book_id", bookId)
	querySeter.One(&book)

	//重新发布
	var documents []*Document
	_, err := o.QueryTable(TNDocuments()).Filter("book_id", bookId).Limit(5000).All(&documents, "document_id")
	if err != nil {
		return
	}

	documentStore := new(DocumentStore)
	for _, doc := range documents {
		content := strings.TrimSpace(documentStore.SelectField(doc.DocumentId, "content"))
		doc.Release = content
		attachList, err := NewAttachment().SelectByDocumentId(doc.DocumentId)
		if err == nil && len(attachList) > 0 {
			content := bytes.NewBufferString("<div class=\"attach-list\"><strong>附件</strong><ul>")
			for _, attach := range attachList {
				li := fmt.Sprintf("<li><a href=\"%s\" target=\"_blank\" title=\"%s\">%s</a></li>", attach.HttpPath, attach.Name, attach.Name)
				content.WriteString(li)
			}
			content.WriteString("</ul></div>")
			doc.Release += content.String()
		}
		o.Update(doc, "release")
	}

	//更新时间戳
	if _, err = querySeter.Update(orm.Params{
		"release_time": time.Now(),
	}); err != nil {
		beego.Error(err.Error())
	}
}

//图书目录
func (m *Document) GetMenuTop(bookId int) (docs []*Document, err error) {
	var docsAll []*Document
	o := orm.NewOrm()
	cols := []string{"document_id", "document_name", "member_id", "parent_id", "book_id", "identify"}
	fmt.Println("---------------start")
	_, err = o.QueryTable(m.TableName()).Filter("book_id", bookId).Filter("parent_id", 0).OrderBy("order_sort", "document_id").Limit(5000).All(&docsAll, cols...)
	fmt.Println("---------------end")
	for _, doc := range docsAll {
		docs = append(docs, doc)
	}
	return
}

//自动生成下一级的内容
func (m *Document) BookStackAuto(bookId, docId int) (md, cont string) {
	//自动生成文档内容
	var docs []Document
	orm.NewOrm().QueryTable("md_documents").Filter("book_id", bookId).Filter("parent_id", docId).OrderBy("order_sort").All(&docs, "document_id", "document_name", "identify")
	var newCont []string //新HTML内容
	var newMd []string   //新markdown内容
	for _, doc := range docs {
		newMd = append(newMd, fmt.Sprintf(`- [%v]($%v)`, doc.DocumentName, doc.Identify))
		newCont = append(newCont, fmt.Sprintf(`<li><a href="$%v">%v</a></li>`, doc.Identify, doc.DocumentName))
	}
	md = strings.Join(newMd, "\n")
	cont = "<ul>" + strings.Join(newCont, "") + "</ul>"
	return
}

//爬虫批量采集
//@param		html				html
//@param		md					markdown内容
//@return		content,markdown	把链接替换为标识后的内容
func (m *Document) BookStackCrawl(html, md string, bookId, uid int) (content, markdown string, err error) {
	var gq *goquery.Document
	content = html
	markdown = md
	project := ""
	if book, err := NewBook().Find(bookId); err == nil {
		project = book.Identify
	}
	//执行采集
	if gq, err = goquery.NewDocumentFromReader(strings.NewReader(content)); err == nil {
		//采集模式mode
		CrawlByChrome := false
		if strings.ToLower(gq.Find("mode").Text()) == "chrome" {
			CrawlByChrome = true
		}
		//内容选择器selector
		selector := ""
		//if selector = strings.TrimSpace(gq.Find("selector").Text()); selector == "" {
		//	err = errors.New("内容选择器不能为空")
		//	return
		//}

		// 截屏选择器
		if screenshot := strings.TrimSpace(gq.Find("screenshot").Text()); screenshot != "" {
			utils.ScreenShotProjects.Store(project, screenshot)
			defer utils.DeleteScreenShot(project)
		}

		//排除的选择器
		var exclude []string
		if excludeStr := strings.TrimSpace(gq.Find("exclude").Text()); excludeStr != "" {
			slice := strings.Split(excludeStr, ",")
			for _, item := range slice {
				exclude = append(exclude, strings.TrimSpace(item))
			}
		}

		var links = make(map[string]string) //map[url]identify

		gq.Find("a").Each(func(i int, selection *goquery.Selection) {
			if href, ok := selection.Attr("href"); ok {
				if !strings.HasPrefix(href, "$") {
					hrefTrim := strings.TrimRight(href, "/")
					identify := utils.MD5Sub16(hrefTrim) + ".md"
					links[hrefTrim] = identify
					links[href] = identify
				}
			}
		})

		gq.Find("a").Each(func(i int, selection *goquery.Selection) {
			if href, ok := selection.Attr("href"); ok {
				hrefLower := strings.ToLower(href)
				//以http或者https开头
				if strings.HasPrefix(hrefLower, "http://") || strings.HasPrefix(hrefLower, "https://") {
					//采集文章内容成功，创建文档，填充内容，替换链接为标识
					if retMD, err := utils.CrawlHtml2Markdown(href, 0, CrawlByChrome, 2, selector, exclude, links, map[string]string{"project": project}); err == nil {
						var doc Document
						identify := utils.MD5Sub16(strings.TrimRight(href, "/")) + ".md"
						doc.Identify = identify
						doc.BookId = bookId
						doc.Version = time.Now().Unix()
						doc.ModifyAt = int(time.Now().Unix())
						doc.DocumentName = selection.Text()
						doc.MemberId = uid

						if docId, err := doc.InsertOrUpdate(); err != nil {
							beego.Error("InsertOrUpdate => ", err)
						} else {
							var ds DocumentStore
							ds.DocumentId = int(docId)
							ds.Markdown = "[TOC]\n\r\n\r" + retMD
							if err := ds.InsertOrUpdate("markdown", "content"); err != nil {
								beego.Error(err)
							}
						}
						selection = selection.SetAttr("href", "$"+identify)
						if _, ok := links[href]; ok {
							markdown = strings.Replace(markdown, "("+href+")", "($"+identify+")", -1)
						}
					} else {
						beego.Error(err.Error())
					}
				}
			}
		})
		content, _ = gq.Find("body").Html()
	}
	return
}
