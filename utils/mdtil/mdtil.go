package mdtil

import "gopkg.in/russross/blackfriday.v2"

//将markdown内容转成html内容
//@param            MarkdownContent     markdown文本内容
//@return           html                转化后的html
func Md2html(MarkdownContent string) (html string) {
	//out := blackfriday.MarkdownCommon([]byte(MarkdownContent))
	out := blackfriday.Run([]byte(MarkdownContent))
	return string(out)
}
