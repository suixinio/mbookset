package so

import (
	"github.com/astaxie/beego/logs"
	"mbook/controllers"
	models "mbook/models/so"
	"mbook/utils/spider"
	//"so-translate/spider"

	//"so-translate/models"
	//"so-translate/spider"
)

type SOController struct {
	controllers.BaseController
}

var importing = false

func (c *SOController) SpiderOSQuestions() {
	// 获取爬取的页数
	page, _ := c.GetInt("p", 1)
	if importing {
		return
	}
	importing = true
	go func() {
		defer func() { importing = false }()
		for i := 1; i <= page; {
			qnas := spider.StackOverflow.ParseQuestionsByVotes(i, 100)
			if nil == qnas {
				continue
			}
			models.NewQuestion().AddQuestions(qnas)

			i++
		}
	}()
	c.JsonResult(0, "OK")
}

func (c *SOController) SpiderOSAnswers() {
	// 获取需要爬取的问题qu的数量
	size, _ := c.GetInt("s", 1)
	q_id := c.GetString("q", "")
	if "" != q_id {
		// 针对特定问题爬取
		q := models.NewQuestion().GetQuestionByQuestionID(q_id)
		if q.ID == 0 {
			c.JsonResult(0, "1")
		}
		ans := spider.StackOverflow.ParseAnswers(q_id)
		models.NewAnswer().AddAnswers(ans, q)
	} else {
		// 批量爬取
		go func() {
			//Get all the unanswered questions
			qs := models.NewQuestion().GetUnansweredQuestions(size)
			qnas := spider.StackOverflow.ParseQuestionsAndAnswers(qs)
			for _, qna := range qnas {
				for _, answer := range qna.Answers {
					answer.QuestionID = qna.Question.ID
					answer.Insertorupdate()
				}
			}
		}()
	}
	c.JsonResult(0, "OK")
}

func (c *SOController) GoogleTranslate() {
	translate_flag := c.GetString("flag", "")
	size, _ := c.GetInt("s", 0)
	qCnt, aCnt := 0, 0

	if translate_flag == "q" && size != 0 {
		questions := models.NewQuestion().GetansweredUntranslatedQuestions(size)
		for _, q := range questions {
			if "" == q.TitleZhCN {
				q.TitleZhCN = spider.Translation.Translate(q.TitleEnUS, "text")
			}
			if "" == q.ContentZhCN {
				q.ContentZhCN = spider.Translation.Translate(q.ContentEnUS, "html")
			}
			if err := q.Update(); nil != err {
				logs.Error("update question failed: " + err.Error())
			}
			logs.Info("translated a question [" + q.Path + "]")
			qCnt++
		}

	}

	if size == 0 && (translate_flag == "qa" || translate_flag == "q") {
		questions := models.NewQuestion().GetUntranslatedQuestions()
		for _, q := range questions {
			if "" == q.TitleZhCN {
				q.TitleZhCN = spider.Translation.Translate(q.TitleEnUS, "text")
			}
			if "" == q.ContentZhCN {
				q.ContentZhCN = spider.Translation.Translate(q.ContentEnUS, "html")
			}
			if err := q.Update(); nil != err {
				logs.Error("update question failed: " + err.Error())
			}
			logs.Info("translated a question [" + q.Path + "]")
			qCnt++
		}

	}

	if size == 0 && (translate_flag == "qa" || translate_flag == "a") {
		answers := models.NewAnswer().GetUntranslatedAnswers()
		for _, an := range answers {
			if "" == an.ContentZhCN {
				an.ContentZhCN = spider.Translation.Translate(an.ContentEnUS, "html")
			}
			if err := an.Update(); nil != err {
				logs.Error("update answer failed: " + err.Error())
			}
			logs.Info("translated an answer [%d]", an.ID)
			aCnt++
		}
	}

	logs.Info("translated questions [%d], answers [%d]", qCnt, aCnt)
	c.JsonResult(0, "1")
}
