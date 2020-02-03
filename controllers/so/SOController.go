package so

import (
	"fmt"
	"github.com/astaxie/beego/logs"
	"mbook/controllers"
	models "mbook/models/so"
	"mbook/utils/spider"
	"time"
)

type SOController struct {
	controllers.BaseController
}

var importing = false

func (c *SOController) SpiderOSQuestions() {
	// 获取爬取的页数
	token := c.GetString("token", "")
	if token != "jiacheng" {
		c.JsonResult(0, "OK")
	}
	page, _ := c.GetInt("p", 1)
	startPage, _ := c.GetInt("s", 1)
	if importing {
		c.JsonResult(0, "error")
	}
	importing = true
	go func() {
		defer func() { importing = false }()
		for i := startPage; i <= page; {
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
	token := c.GetString("token", "")
	if token != "jiacheng" {
		c.JsonResult(0, "OK")
	}
	// 获取需要爬取的问题qu的数量
	size, _ := c.GetInt("size", 1)
	loopCount, _ := c.GetInt("loop", 1)
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
			for i := 1; i <= loopCount; i++ {
				//Get all the unanswered questions
				qs := models.NewQuestion().GetUnansweredQuestions(size)
				qnas := spider.StackOverflow.ParseQuestionsAndAnswers(qs)
				for _, qna := range qnas {
					for _, answer := range qna.Answers {
						answer.QuestionID = qna.Question.ID
						answer.Insertorupdate()
					}
				}
			}
		}()
	}
	c.JsonResult(0, "OK")
}

// 这个模块需要进行一次更新
// todo 针对每天的访问进行一次更新，交给宝塔每天进行更新
// todo 拆分翻译部分为一个单独的模块进程
func (c *SOController) GoogleTranslate() {
	token := c.GetString("token", "")
	if token != "jiacheng" {
		c.JsonResult(0, "OK")
	}
	translate_flag := c.GetString("flag", "")
	size, _ := c.GetInt("size", 0)
	qCnt, aCnt := 0, 0
	fmt.Println(translate_flag, size)
	//翻译有answer的question
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
			time.Sleep(time.Second)
		}
	}

	//if size == 0 && (translate_flag == "qa" || translate_flag == "q") {
	//	//	questions := models.NewQuestion().GetUntranslatedQuestions()
	//	//	for _, q := range questions {
	//	//		if "" == q.TitleZhCN {
	//	//			q.TitleZhCN = spider.Translation.Translate(q.TitleEnUS, "text")
	//	//		}
	//	//		if "" == q.ContentZhCN {
	//	//			q.ContentZhCN = spider.Translation.Translate(q.ContentEnUS, "html")
	//	//		}
	//	//		if err := q.Update(); nil != err {
	//	//			logs.Error("update question failed: " + err.Error())
	//	//		}
	//	//		logs.Info("translated a question [" + q.Path + "]")
	//	//		qCnt++
	//	//		time.Sleep(time.Second)
	//	//	}
	//	//}

	// 翻译所有没有翻译的回答
	if translate_flag == "a" && size != 0 {
		answers := models.NewAnswer().GetUntranslatedAnswers(size)
		for _, an := range answers {
			if "" == an.ContentZhCN {
				an.ContentZhCN = spider.Translation.Translate(an.ContentEnUS, "html")
			}
			if err := an.Update(); nil != err {
				logs.Error("update answer failed: " + err.Error())
			}
			logs.Info("translated an answer [%d]", an.ID)
			aCnt++
			time.Sleep(time.Second)
		}
	}

	logs.Info("translated questions [%d], answers [%d]", qCnt, aCnt)
	c.JsonResult(0, "1")
}
