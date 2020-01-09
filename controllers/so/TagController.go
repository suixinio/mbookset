package so

import (
	"github.com/astaxie/beego/logs"
	"mbook/common"
	"mbook/controllers"
	models "mbook/models/so"
)

type TagController struct {
	controllers.BaseController
}

func (c *TagController) ShowTagAction() {
	tag := c.Ctx.Input.Param(":splat")
	logs.Info("标签：%s", tag)
	page, _ := c.GetInt("p", 1)
	c.TplName = "so/tag.html"

	// 获取title的数据库数据
	tagModel := models.NewTag().GetTagByTitle(tag)
	if nil == tagModel {
		// 如果没有找到
		return
	}
	qModels, pagination := models.NewQuestion().GetTagTranslatedQuestions(tagModel.ID, page)
	questions := questionsVos(qModels)
	c.Data["Questions"] = questions
	c.Data["Pagination"] = pagination

	c.Data["MetaKeywords"] = common.MetaKeywords + "," + tagModel.Title
	c.Data["Title"] = "关于 " + tagModel.Title + " 的问题 - BookSet"

}
