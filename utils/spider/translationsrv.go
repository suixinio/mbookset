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
	"context"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strings"

	//"cloud.google.com/go/translate"
	translate "cloud.google.com/go/translate/apiv3"
	translatepb "google.golang.org/genproto/googleapis/cloud/translate/v3"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/proxy"
)

// Translation service.
var Translation = &translationService{}

type translationService struct {
}

func (srv *translationService) Translate(text string, format string) string {

	//if is_proxy, err := beego.AppConfig.Bool("is_proxy"); nil != err && is_proxy {
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:1080", nil, proxy.Direct)
	if err != nil {
		logs.Error("can't connect to the proxy: " + err.Error())
	}

	httpTransport := &http.Transport{Dial: dialer.Dial}
	http.DefaultClient.Transport = httpTransport

	//}

	//ctx := context.Background()
	ctx := context.Background()

	//client, err := translate.NewClient(ctx)
	client, err := translate.NewTranslationClient(ctx)
	if err != nil {
		logs.Error("create translate client failed: " + err.Error() + "123123123")

		return ""
	}

	ret := ""

	//translations, err := client.Translate(ctx, []string{text}, language.Chinese,
	//	&translate.Options{Source: language.English, Format: translate.Format(format), Model: "nmt"})
	//
	modelID := beego.AppConfig.DefaultString("modelID", "general/nmt")
	projectID := beego.AppConfig.DefaultString("projectID", "")
	location := beego.AppConfig.DefaultString("location", "global")

	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/global", projectID),
		SourceLanguageCode: "en",
		TargetLanguageCode: "zh",
		MimeType:           "text/html", // Mime types: "text/plain", "text/html"
		Contents:           []string{text},
		Model: fmt.Sprintf("projects/%s/locations/%s/models/%s", projectID, location, modelID),

	}

	translations, err := client.TranslateText(ctx, req)

	if nil == err {
		translation := translations.GetTranslations()[0]
		ret = translation.GetTranslatedText()
		//for _, translation := range translations.GetTranslations() {
		//	fmt.Fprintf(w, "Translated text: %v\n", translation.GetTranslatedText())
		//}
		//ret = translations[0].Text
		//ret = resp.GetTranslations
		//ret = resp
	}

	if "" == ret {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(text))
		if nil != err {
			logs.Error("parse text to HTML doc failed: " + err.Error())

			return ""
		}

		fragment := ""
		pCount := 0
		doc.Find("body").Children().Each(func(i int, s *goquery.Selection) {
			nodeName := goquery.NodeName(s)
			html, _ := s.Html()
			if "pre" == nodeName || "code" == nodeName {
				ret += translateFragment(client, ctx, fragment)
				ret += "<" + nodeName + ">" + html + "</" + nodeName + ">"
				fragment = ""
				pCount = 0

				return
			}

			if "" == html {
				fragment += "<" + nodeName + ">"
			} else {
				fragment += "<" + nodeName + ">" + html + "</" + nodeName + ">"
			}

			if "p" == nodeName {
				pCount++
			}

			if 3 < pCount {
				ret += translateFragment(client, ctx, fragment)
				fragment = ""
				pCount = 0
			}
		})

		if "" != fragment {
			ret += translateFragment(client, ctx, fragment)
		}
	}

	return ret
}

func translateFragment(client *translate.TranslationClient, ctx context.Context, fragment string) string {
	//translations, err := client.Translate(ctx, []string{fragment}, language.Chinese,
	modelID := beego.AppConfig.DefaultString("modelID", "general/nmt")
	projectID := beego.AppConfig.DefaultString("projectID", "")
	location := beego.AppConfig.DefaultString("location", "global")

	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/global", projectID),
		SourceLanguageCode: "en",
		TargetLanguageCode: "zh",
		MimeType:           "text/html", // Mime types: "text/plain", "text/html"
		Contents:           []string{fragment},
		Model: fmt.Sprintf("projects/%s/locations/%s/models/%s", projectID, location, modelID),

	}

	translations, err := client.TranslateText(ctx, req)

	//translations, err := client.Translate(ctx, []string{fragment}, language.Chinese,
	//	&translate.Options{Source: language.English, Format: translate.HTML, Model: "nmt"})

	if nil != err {
		logs.Error("translate failed: " + err.Error())
		return ""
		//return fragment
	}

	translated := translations.GetTranslations()[0].GetTranslatedText()
	if "" != translated {
		return translated
	}

	return fragment
}
