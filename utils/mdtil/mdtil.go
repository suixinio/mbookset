package mdtil

import (
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"mbook/utils/filetil"
	"strings"
)

//将markdown内容转成html内容
//@param            MarkdownContent     markdown文本内容
//@return           html                转化后的html
func Md2html(MarkdownContent string) (html string) {
	//out := blackfriday.MarkdownCommon([]byte(MarkdownContent))
	out := blackfriday.Run([]byte(MarkdownContent))
	return string(out)
}

//查到summary，并将内容转换成map
func SummaryToMap(unzipPath string) (summary map[string]string) {
	summary = make(map[string]string)

	if files, err := filetil.ScanFiles(unzipPath); err == nil {
		for _, file := range files {
			if strings.HasSuffix(strings.ToLower(file.Name), "summary.md") || strings.HasSuffix(strings.ToLower(file.Name), "summary.html") {
				// 找到summary文件
				if b, err := ioutil.ReadFile(file.Path); err == nil {
					// summary文件内容
					//mdcont := strings.TrimSpace(string(b))
					output := blackfriday.Run(b)
					doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(output)))
					//a标签链接处理。要注意判断有锚点的情况
					doc.Find("a").Each(func(i int, selection *goquery.Selection) {
						if href, ok := selection.Attr("href"); ok && !strings.HasPrefix(strings.ToLower(href), "http") && !strings.HasPrefix(href, "#") {
							nameHref, err := selection.Html()
							if err != nil {
								nameHref = "blank"
							}
							summary[href] = nameHref
						}
					})
				}
			}
		}
	}
	return
}
