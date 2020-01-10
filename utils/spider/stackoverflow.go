// 协慌网 - 专注编程问答汉化 https://routinepanic.com
// Copyright (C) 2018-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package spider

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/astaxie/beego/logs"
	"github.com/parnurzeal/gorequest"
	"html"
	"mbook/common"
	models "mbook/models/so"
	"strconv"
	"strings"
	"time"
)

var StackOverflow = &stackOverflow{}

type stackOverflow struct{}

// Logger

const stackExchangeAPI = "https://api.stackexchange.com"

// QnA represents a question and its answers.
type QnA struct {
	Question *models.Question
	Answers  []*models.Answer
}

func (s *stackOverflow) ParseQuestionsAndAnswersByVotes(page, pageSize int) (ret []*QnA) {
	questsions := s.ParseQuestionsByVotes(page, pageSize)
	for _, qi := range questsions {
		qId := qi.SourceID

		answers := s.ParseAnswers(qId)
		qna := &QnA{Question: qi, Answers: answers}
		ret = append(ret, qna)
	}
	return
}

func (s *stackOverflow) ParseQuestionsAndAnswers(questsions []*models.Question) (ret []*QnA) {
	//questsions := s.ParseQuestionsByVotes(page, pageSize)
	for _, qi := range questsions {
		qId := qi.SourceID
	LoopNil:
		answers := s.ParseAnswers(qId)
		if nil == answers {
			goto LoopNil
		}
		qna := &QnA{Question: qi, Answers: answers}
		ret = append(ret, qna)
	}
	return
}

func (s *stackOverflow) GetProxy() (qs string) {
	request := gorequest.New()
	data := map[string]interface{}{}
	url := "http://182.92.105.252:5010/get/"
	response, body, errs := request.Set("User-Agent", common.UserAgent).Get(url).Timeout(30*time.Second).Retry(3, 5*time.Second).EndStruct(&data)
	if nil != response &&  200 != response.StatusCode {
		logs.Error("get [%s] status code is [%d], response body is [%s]", url, response.StatusCode, body)
		return
		//return nil
	}
	if nil != errs {
		logs.Error("get [%s] failed: %s", url, errs)
		return
		//return nil
	}
	qs = data["proxy"].(string)
	qs = "http://" + qs
	logs.Info("代理IP:" + qs)
	return

}

func (s *stackOverflow) ParseQuestionsByVotes(page, pageSize int) (ret []*models.Question) {
	logs.Info("questions requesting [page=" + strconv.Itoa(page) + ", pageSize=" + strconv.Itoa(pageSize) + "]")
	request := gorequest.New().Proxy(s.GetProxy())
	var url = stackExchangeAPI + "/2.2/questions?page=" + strconv.Itoa(page) + "&pagesize=" + strconv.Itoa(pageSize) + "&order=desc&sort=votes&site=stackoverflow&filter=!9Z(-wwYGT"
	data := map[string]interface{}{}
	response, body, errs := request.Set("User-Agent", common.UserAgent).Get(url).Timeout(30*time.Second).Retry(3, 5*time.Second).EndStruct(&data)
	logs.Info("questions requested [page=" + strconv.Itoa(page) + ", pageSize=" + strconv.Itoa(pageSize) + "]")
	if nil != errs {
		logs.Error("get [%s] failed: %s", url, errs)

		return nil
	}
	if 200 != response.StatusCode {
		logs.Error("get [%s] status code is [%d], response body is [%s]", url, response.StatusCode, body)

		return nil
	}

	qs := data["items"].([]interface{})
	for _, qi := range qs {
		q := qi.(map[string]interface{})
		question := &models.Question{}
		title := q["title"].(string)
		title = html.UnescapeString(title)
		question.TitleEnUS = title
		tis := q["tags"].([]interface{})
		var tags []string
		for _, ti := range tis {
			tags = append(tags, ti.(string))
		}
		question.Tags = strings.Join(tags, ",")
		question.Votes = int(q["score"].(float64))
		question.Views = int(q["view_count"].(float64))
		content := q["body"].(string)
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(content))
		doc.Find("pre,code").Each(func(i int, s *goquery.Selection) {
			s.SetAttr("translate", "no")
		})
		question.ContentEnUS, _ = doc.Find("body").Html()
		link := q["link"].(string)
		qId := strconv.Itoa(int(q["question_id"].(float64)))
		question.Path = strings.Split(link, qId+"/")[1]
		question.Source = models.SourceStackOverflow
		question.SourceID = qId
		question.SourceURL = link
		owner := q["owner"].(map[string]interface{})
		authorName := ""
		if nil == owner["display_name"] {
			authorName = "someone"
		} else {
			authorName = owner["display_name"].(string)
		}
		question.AuthorName = authorName
		l := owner["link"]
		if nil != l {
			question.AuthorURL = l.(string)
		}

		//answers := s.ParseAnswers(qId)
		//qna := &QnA{Question: question, Answers: answers}
		ret = append(ret, question)

		logs.Info("parsed voted question [id=%s]", question.SourceID)
	}

	logs.Info("parsed voted questions [page=%d]", page)

	return
}

func (s *stackOverflow) ParseAnswers(questionId string) (ret []*models.Answer) {
	logs.Info("answer requesting for question [id=" + questionId + "]")
	request := gorequest.New().Proxy(s.GetProxy())
	var url = stackExchangeAPI + "/2.2/questions/" + questionId + "/answers?pagesize=10&order=desc&sort=votes&site=stackoverflow&filter=!9Z(-wzu0T"
	data := map[string]interface{}{}
	response, _, errs := request.Set("User-Agent", common.UserAgent).Get(url).Timeout(30*time.Second).Retry(3, 5*time.Second).EndStruct(&data)
	logs.Info("answer requested [questionId=" + questionId + "]")
	if nil != errs {
		logs.Error("get [%s] failed: %s", url, errs)

		return nil
	}
	if 200 != response.StatusCode {
		logs.Error("get [%s] status code is [%d]", url, response.StatusCode)

		return nil
	}

	as := data["items"].([]interface{})
	for _, ai := range as {
		a := ai.(map[string]interface{})
		answer := &models.Answer{}
		answer.Votes = int(a["score"].(float64))
		content := a["body"].(string)
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(content))
		doc.Find("pre,code").Each(func(i int, s *goquery.Selection) {
			s.SetAttr("translate", "no")
		})
		answer.ContentEnUS, _ = doc.Find("body").Html()
		answer.Source = models.SourceStackOverflow
		answer.SourceID = strconv.Itoa(int(a["answer_id"].(float64)))
		//answer.QuestionID = strconv.Itoa(int(a["question_id"].(int64)))
		owner := a["owner"].(map[string]interface{})
		if nil != owner {
			n := owner["display_name"]
			if nil != n {
				answer.AuthorName = n.(string)
			}
			l := owner["link"]
			if nil != l {
				answer.AuthorURL = l.(string)
			}
		}

		ret = append(ret, answer)
	}

	logs.Info("parsed answers for question [id=" + questionId + "]")

	return
}

//func (s *stackOverflow) ParseQuestionsAndAnswersByVotes(page, pageSize int) (ret []*QnA) {
//
//	logs.Info("questions requesting [page=" + strconv.Itoa(page) + ", pageSize=" + strconv.Itoa(pageSize) + "]")
//	request := gorequest.New()
//	var url = stackExchangeAPI + "/2.2/questions?page=" + strconv.Itoa(page) + "&pagesize=" + strconv.Itoa(pageSize) + "&order=desc&sort=votes&site=stackoverflow&filter=!9Z(-wwYGT"
//	data := map[string]interface{}{}
//	response, body, errs := request.Set("User-Agent", conf.UserAgent).Get(url).Timeout(30*time.Second).Retry(3, 5*time.Second).EndStruct(&data)
//	logs.Info("questions requested [page=" + strconv.Itoa(page) + ", pageSize=" + strconv.Itoa(pageSize) + "]")
//	if nil != errs {
//		logs.Error("get [%s] failed: %s", url, errs)
//
//		return nil
//	}
//	if 200 != response.StatusCode {
//		logs.Error("get [%s] status code is [%d], response body is [%s]", url, response.StatusCode, body)
//
//		return nil
//	}
//
//	qs := data["items"].([]interface{})
//	for _, qi := range qs {
//		q := qi.(map[string]interface{})
//		question := &models.Question{}
//		title := q["title"].(string)
//		title = html.UnescapeString(title)
//		question.TitleEnUS = title
//		tis := q["tags"].([]interface{})
//		var tags []string
//		for _, ti := range tis {
//			tags = append(tags, ti.(string))
//		}
//		question.Tags = strings.Join(tags, ",")
//		question.Votes = int(q["score"].(float64))
//		question.Views = int(q["view_count"].(float64))
//		content := q["body"].(string)
//		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(content))
//		doc.Find("pre,code").Each(func(i int, s *goquery.Selection) {
//			s.SetAttr("translate", "no")
//		})
//		question.ContentEnUS, _ = doc.Find("body").Html()
//		link := q["link"].(string)
//		qId := strconv.Itoa(int(q["question_id"].(float64)))
//		question.Path = strings.Split(link, qId+"/")[1]
//		question.Source = models.SourceStackOverflow
//		question.SourceID = qId
//		question.SourceURL = link
//		owner := q["owner"].(map[string]interface{})
//		authorName := ""
//		if nil == owner["display_name"] {
//			authorName = "someone"
//		} else {
//			authorName = owner["display_name"].(string)
//		}
//		question.AuthorName = authorName
//		l := owner["link"]
//		if nil != l {
//			question.AuthorURL = l.(string)
//		}
//
//		answers := s.ParseAnswers(qId)
//		qna := &QnA{Question: question, Answers: answers}
//		ret = append(ret, qna)
//
//		logs.Info("parsed voted question [id=%s]", qna.Question.SourceID)
//	}
//
//	logs.Info("parsed voted questions [page=%d]", page)
//
//	return
//}
