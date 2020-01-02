package controllers

import (
	"github.com/astaxie/beego"
	"mbook/models"
)

type HomeController struct {
	BaseController
}

func (c *HomeController) Index() {
	cid, _ := c.GetInt("cid")
	if cid > 0 {
		c.Redirect(beego.URLFor("HomeController.Index")+c.Ctx.Request.RequestURI, 302)
	}
	c.List()
	//if cates, err := new(models.Category).GetCates(-1, 1); err == nil {
	//	c.Data["Cates"] = cates
	//	fmt.Println(cates)
	//} else {
	//	beego.Error(err.Error())
	//}
}

//分类
func (this *HomeController) List() {
	if cates, err := new(models.Category).GetCates(-1, 1); err == nil {
		this.Data["Cates"] = cates
	} else {
		beego.Error(err.Error())
	}
	this.GetSeoByPage("cate", map[string]string{
		"title":       "书籍分类",
		"keywords":    "文档托管,在线创作,文档在线管理,在线知识管理,文档托管平台,在线写书,文档在线转换,在线编辑,在线阅读,开发手册,api手册,文档在线学习,技术文档,在线编辑",
		"description": this.Sitename + "专注于文档在线写作、协作、分享、阅读与托管，让每个人更方便地发布、分享和获得知识。",
	})
	this.Data["IsCate"] = true
	this.Data["Recommends"], _, _ = models.NewBook().HomeData(1, 12, models.OrderLatestRecommend, 0)
	//this.TplName = "cates/list.html"
	this.TplName = "home/list.html"

}
