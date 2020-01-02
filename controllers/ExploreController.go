package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"math"
	"mbook/models"
	"mbook/utils"
	"strconv"
)

type ExploreController struct {
	BaseController
}

func (c *ExploreController) Index() {
	fmt.Println(c)
	var (
		cid       int // 分类id
		cate      models.Category
		urlPrefix = beego.URLFor("ExploreController.Index")
		tabName   = map[string]string{"recommend": "站长推荐", "latest": "最新发布", "popular": "热门书籍"}
	)
	if cid, _ = c.GetInt("cid"); cid >= 0 {
		cateModel := new(models.Category)
		cate = cateModel.Find(cid)
		c.Data["Cate"] = cate
	}
	c.Data["Cid"] = cid
	c.TplName = "explore/index.html"
	pageIndex, _ := c.GetInt("page", 1)
	pageSize := 24
	books, totalCount, err := new(models.Book).HomeData(pageIndex, pageSize,models.BookOrder("latest"), cid)
	if err != nil {
		beego.Error(err)
		c.Abort("404")
	}
	if totalCount > 0 {
		urlSuffix := ""
		if cid > 0 {
			urlPrefix = urlPrefix + "&cid=" + strconv.Itoa(cid)
		}
		html := utils.NewPaginations(4, totalCount, pageSize, pageIndex, urlPrefix, urlSuffix)
		c.Data["PageHtml"] = html
	} else {
		c.Data["PageHtml"] = ""
	}
	c.Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	c.Data["Lists"] = books

	//this.Data["TotalPages"] = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	//this.Data["Lists"] = books
	//this.Data["Tab"] = tab
	//this.Data["Lang"] = lang
	title := c.Sitename
	if cid > 0 {
		title = "[发现] " + cate.Title + " - " + tabName["latest"] + " - " + title
	} else {
		title = "探索，发现新世界，畅想新知识 - " + c.Sitename
	}
	c.GetSeoByPage("explore", map[string]string{
		"title":       title,
		"keywords":    "文档托管,在线创作,文档在线管理,在线知识管理,文档托管平台,在线写书,文档在线转换,在线编辑,在线阅读,开发手册,api手册,文档在线学习,技术文档,在线编辑",
		"description": c.Sitename + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
}
