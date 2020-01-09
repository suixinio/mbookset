package so

import (
	"mbook/controllers"
	models "mbook/models/so"
)

type IndexController struct {
	controllers.BaseController
}

func (c *IndexController) Index() {
	page, _ := c.GetInt("p", 1)
	qModels, pagination := models.NewQuestion().GetQuestions(page)
	questions := questionsVos(qModels)

	c.Data["Pagination"] = pagination
	c.Data["Questions"] = questions

	c.Data["MetaKeywords"] = "程序员,编程,代码,问答,javascript,git,python,java,c#,html"
	c.Data["Title"] = "专注编程问答汉化"

	c.TplName = "so/index.html"
}
