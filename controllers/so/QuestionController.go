package so

import (
	"mbook/controllers"
	models "mbook/models/so"
)

type QuestionController struct {
	controllers.BaseController
}

func (c *QuestionController) ShowQuestionAction() {
	questionID := c.Ctx.Input.Param(":questionID")
	// 通过questionID获取问题
	qModel := models.NewQuestion().GetQuestionByQuestionID(questionID)
	if qModel.ID == 0{
		c.Abort("404")
	}
	question := questionVo(qModel)
	// 获取问题的回答
	aModels := models.NewAnswer().GetAnswers(qModel.ID)

	c.Data["Answers"] = answersVos(*aModels)
	c.Data["Question"] = question

	// SEO
	c.Data["Title"] = question.Title + " - " + "BookSet"
	c.Data["MetaKeywords"] = qModel.Tags
	c.Data["MetaDescription"] = question.Description
	c.TplName = "so/question.html"
}
