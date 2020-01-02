package controllers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"io"
	"mbook/common"
	"mbook/models"
	"mbook/utils"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type BaseController struct {
	beego.Controller
	Member          *models.Member    //用户
	Option          map[string]string //全局设置
	EnableAnonymous bool              //开启匿名访问
	Sitename        string
}
type CookieRemember struct {
	MemberId int
	Account  string
	Time     time.Time
}

//每个子类Controller公用方法调用前，都执行一下Prepare方法
func (c *BaseController) Prepare() {
	c.Member = models.NewMember() //初始化
	c.EnableAnonymous = false
	//从session中获取用户信息
	if member, ok := c.GetSession(common.SessionName).(models.Member); ok && member.MemberId > 0 {
		c.Member = &member
	} else {
		//如果Cookie中存在登录信息，从cookie中获取用户信息
		if cookie, ok := c.GetSecureCookie(common.AppKey(), "login"); ok {
			var remember CookieRemember
			err := utils.Decode(cookie, &remember)
			if err == nil {
				member, err := models.NewMember().Find(remember.MemberId)
				if err == nil {
					c.SetMember(*member)
					c.Member = member
				}
			}
		}
	}
	if c.Member.RoleName == "" {
		c.Member.RoleName = common.Role(c.Member.MemberId)
	}
	c.Data["Member"] = c.Member
	c.Data["BaseUrl"] = c.BaseUrl()
	c.Sitename = "BookSet"
	c.Data["SITE_NAME"] = c.Sitename
	//设置全局配置
	c.Option = make(map[string]string)
	c.Option["ENABLED_CAPTCHA"] = "false"
	controller, action := c.GetControllerAndAction()

	c.Data["ActionName"] = action
	c.Data["ControllerName"] = controller

}

// Ajax接口返回Json
func (c *BaseController) JsonResult(errCode int, errMsg string, data ...interface{}) {
	jsonData := make(map[string]interface{}, 3)
	jsonData["errcode"] = errCode
	jsonData["message"] = errMsg

	if len(data) > 0 && data[0] != nil {
		jsonData["data"] = data[0]
	}
	returnJSON, err := json.Marshal(jsonData)
	if err != nil {
		beego.Error(err)
	}
	c.Ctx.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")
	//启用gzip压缩
	if strings.Contains(strings.ToLower(c.Ctx.Request.Header.Get("Accept-Encoding")), "gzip") {
		c.Ctx.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		w := gzip.NewWriter(c.Ctx.ResponseWriter)
		defer w.Close()
		w.Write(returnJSON)
		w.Flush()
	} else {
		io.WriteString(c.Ctx.ResponseWriter, string(returnJSON))
	}
	c.StopRun()
}

func (c *BaseController) BaseUrl() string {
	host := beego.AppConfig.String("sitemap_host")
	if len(host) > 0 {
		if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
			return host
		}
		return c.Ctx.Input.Scheme() + "://" + host
	}
	return c.Ctx.Input.Scheme() + "://" + c.Ctx.Request.Host
}

// 设置登录用户信息
func (c *BaseController) SetMember(member models.Member) {
	if member.MemberId <= 0 {
		c.DelSession(common.SessionName)
		c.DelSession("uid")
		c.DestroySession()
	} else {
		c.SetSession(common.SessionName, member)
		c.SetSession("uid", member.MemberId)
	}
}

//关注或取消关注
func (c *BaseController) SetFollow() {
	if c.Member.MemberId == 0 {
		c.JsonResult(1, "请先登录")
	}
	uid, _ := c.GetInt(":uid")
	if uid == c.Member.MemberId {
		c.JsonResult(1, "不能关注自己")
	}
	cancel, _ := new(models.Fans).FollowOrCancel(uid, c.Member.MemberId)
	if cancel {
		c.JsonResult(0, "已成功取消关注")
	}
	c.JsonResult(0, "已成功关注")
}

//内容采集
func (this *BaseController) Crawl() {
	if this.Member.MemberId > 0 {
		if val, ok := this.GetSession("crawl").(string); ok && val == "1" {
			this.JsonResult(1, "您提交的上一次采集未完成，请稍后再提交新的内容采集")
		}
		this.SetSession("crawl", "1")
		defer this.DelSession("crawl")
		urlStr := this.GetString("url")
		force, _ := this.GetBool("force")              //是否是强力采集，强力采集，使用Chrome
		intelligence, _ := this.GetInt("intelligence") //是否是强力采集，强力采集，使用Chrome
		contType, _ := this.GetInt("type")
		diySel := this.GetString("diy")
		bookId := this.GetString("bookId")

		book, err := models.NewBook().Select("book_id", bookId)
		// todo 判断一下是否有权限操作
		headers := map[string]string{}
		headers["project"] = book.Identify

		content, err := utils.CrawlHtml2Markdown(urlStr, contType, force, intelligence, diySel, []string{}, nil, headers)
		if err != nil {
			this.JsonResult(1, "采集失败："+err.Error())
		}
		this.JsonResult(0, "采集成功", content)
	}
	this.JsonResult(1, "请先登录再操作")
}

//在markdown头部加上<bookstack></bookstack>或者<bookstack/>，即解析markdown中的ul>li>a链接作为目录
func (this *BaseController) sortBySummary(bookIdentify, htmlStr string, bookId int) {
	debug := beego.AppConfig.String("runmod") != "prod"
	o := orm.NewOrm()
	qs := o.QueryTable("md_documents").Filter("book_id", bookId)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlStr))
	if err != nil {
		beego.Error(err)
	}
	idx := 1
	if debug {
		beego.Info("根据summary文件进行排序")
	}

	//查找ul>li下的所有a标签，并提取text和href，查询数据库，如果标识不存在，则把这些新的数据录入数据库
	var hrefs = make(map[string]string)
	var hrefSlice []interface{}
	var docs []models.Document
	doc.Find("li>a").Each(func(i int, selection *goquery.Selection) {
		if href, ok := selection.Attr("href"); ok && strings.HasPrefix(href, "$") {
			href = strings.TrimLeft(strings.Replace(href, "/", "-", -1), "$")
			if utf8.RuneCountInString(href) <= 100 {
				hrefs[href] = selection.Text()
				hrefSlice = append(hrefSlice, href)
			}
		}
	})
	if debug {
		beego.Info(hrefs)
	}
	if len(hrefSlice) > 0 {
		if _, err := qs.Filter("identify__in", hrefSlice...).Limit(len(hrefSlice)).All(&docs, "identify"); err != nil {
			beego.Error(err.Error())
		} else {
			for _, doc := range docs {
				//删除存在的标识
				delete(hrefs, doc.Identify)
			}
		}
	}
	if len(hrefs) > 0 { //存在未创建的文档，先创建
		ModelStore := new(models.DocumentStore)
		for identify, docName := range hrefs {
			doc := models.Document{
				BookId:       bookId,
				Identify:     identify,
				DocumentName: docName,
				CreateTime:   time.Now(),
				ModifyTime:   time.Now(),
			}
			// 如果文档标识超过了规定长度（100），则进行忽略
			if utf8.RuneCountInString(identify) <= 100 {
				if docId, err := doc.InsertOrUpdate(); err == nil {
					ModelStore.DocumentId = int(docId)
					ModelStore.Markdown = "[TOC]\n\r\n\r"

					if err = ModelStore.InsertOrUpdate(); err != nil {
						beego.Error(err.Error())
					}
				}
			}
		}

	}

	// 重置所有之前的文档排序
	_, _ = qs.Update(orm.Params{"order_sort": 100000})

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		docName := selection.Text()
		pid := 0
		if docId, exist := selection.Attr("data-pid"); exist {
			did, _ := strconv.Atoi(docId)
			eleParent := selection.Parent().Parent().Parent()
			if eleParent.Is("li") {
				fst := eleParent.Find("a").First()
				pidstr, _ := fst.Attr("data-pid")
				//如果这里的pid为0，表示数据库还没存在这个标识，需要创建
				pid, _ = strconv.Atoi(pidstr)
			}
			if did > 0 {
				qs.Filter("document_id", did).Update(orm.Params{
					"parent_id": pid, "document_name": docName,
					"order_sort": idx, "modify_time": time.Now(),
				})
			}
		} else if href, ok := selection.Attr("href"); ok && strings.HasPrefix(href, "$") {
			identify := strings.TrimPrefix(href, "$") //文档标识
			eleParent := selection.Parent().Parent().Parent()
			if eleParent.Is("li") {
				if parentHref, ok := eleParent.Find("a").First().Attr("href"); ok {
					var one models.Document
					qs.Filter("identify", strings.Split(strings.TrimPrefix(parentHref, "$"), "#")[0]).One(&one, "document_id")
					pid = one.DocumentId
				}
			}

			if _, err := qs.Filter("identify", identify).Update(orm.Params{
				"parent_id": pid, "document_name": docName,
				"order_sort": idx, "modify_time": time.Now(),
			}); err != nil {
				beego.Error(err)
			}
		}
		idx++
	})

	if len(hrefs) > 0 { //如果有新创建的文档，则再调用一遍，用于处理排序
		this.replaceLinks(bookIdentify, htmlStr, true)
	}
}

//排序
type Sort struct {
	Id        int
	Pid       int
	SortOrder int
	Identify  string
}

//替换链接
//如果是summary，则根据这个进行排序调整
func (this *BaseController) replaceLinks(bookIdentify string, docHtml string, isSummary ...bool) string {
	var (
		book models.Book
		docs []models.Document
		o    = orm.NewOrm()
	)

	o.QueryTable("md_books").Filter("identify", bookIdentify).One(&book, "book_id")
	if book.BookId > 0 {
		o.QueryTable("md_documents").Filter("book_id", book.BookId).Limit(5000).All(&docs, "identify", "document_id")
		if len(docs) > 0 {
			Links := make(map[string]string)
			for _, doc := range docs {
				idStr := strconv.Itoa(doc.DocumentId)
				if len(doc.Identify) > 0 {
					Links["$"+strings.ToLower(doc.Identify)] = beego.URLFor("DocumentController.Read", ":key", bookIdentify, ":id", doc.Identify) + "||" + idStr
				}
				if doc.DocumentId > 0 {
					Links["$"+strconv.Itoa(doc.DocumentId)] = beego.URLFor("DocumentController.Read", ":key", bookIdentify, ":id", doc.DocumentId) + "||" + idStr
				}
			}

			//替换文档内容中的链接
			if gq, err := goquery.NewDocumentFromReader(strings.NewReader(docHtml)); err == nil {
				gq.Find("a").Each(func(i int, selection *goquery.Selection) {
					if href, ok := selection.Attr("href"); ok && strings.HasPrefix(href, "$") {
						if slice := strings.Split(href, "#"); len(slice) > 1 {
							if newHref, ok := Links[strings.ToLower(slice[0])]; ok {
								arr := strings.Split(newHref, "||") //整理的arr数组长度，肯定为2，所以不做数组长度判断
								selection.SetAttr("href", arr[0]+"#"+strings.Join(slice[1:], "#"))
								selection.SetAttr("data-pid", arr[1])
							}
						} else {
							if newHref, ok := Links[strings.ToLower(href)]; ok {
								arr := strings.Split(newHref, "||") //整理的arr数组长度，肯定为2，所以不做数组长度判断
								selection.SetAttr("href", arr[0])
								selection.SetAttr("data-pid", arr[1])
							}
						}
					}
				})

				if newHtml, err := gq.Find("body").Html(); err == nil {
					docHtml = newHtml
					if len(isSummary) > 0 && isSummary[0] == true { //更新排序
						this.sortBySummary(bookIdentify, docHtml, book.BookId) //更新排序
					}
				}
			} else {
				beego.Error(err.Error())
			}
		}
	}

	return docHtml
}

//根据页面获取seo
//@param			page			页面标识
//@param			defSeo			默认的seo的map，必须有title、keywords和description字段
func (this *BaseController) GetSeoByPage(page string, defSeo map[string]string) {
	var seo models.Seo

	orm.NewOrm().QueryTable(models.TableSeo).Filter("Page", page).One(&seo)
	defSeo["sitename"] = this.Sitename
	if seo.Id > 0 {
		for k, v := range defSeo {
			seo.Title = strings.Replace(seo.Title, fmt.Sprintf("{%v}", k), v, -1)
			seo.Keywords = strings.Replace(seo.Keywords, fmt.Sprintf("{%v}", k), v, -1)
			seo.Description = strings.Replace(seo.Description, fmt.Sprintf("{%v}", k), v, -1)
		}
	}
	this.Data["SeoTitle"] = seo.Title
	this.Data["SeoKeywords"] = seo.Keywords
	this.Data["SeoDescription"] = seo.Description
}
