package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/russross/blackfriday.v2"

	"html/template"
	"io/ioutil"
	"mbook/common"
	"mbook/conf"
	"mbook/models"
	"mbook/utils"
	"mbook/utils/filetil"
	"mbook/utils/graphics"
	"mbook/utils/html2md"
	"mbook/utils/mdtil"
	"mbook/utils/store"
	"mbook/utils/ziptil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type BookController struct {
	BaseController
}

//我的图书页面
func (c *BookController) Index() {
	pageIndex, _ := c.GetInt("page", 1)
	private, _ := c.GetInt("private", 1) //默认私有
	books, totalCount, err := models.NewBook().SelectPage(pageIndex, common.PageSize, c.Member.MemberId, private)
	if err != nil {
		logs.Error("BookController.Index => ", err)
		c.Abort("404")
	}
	if totalCount > 0 {
		c.Data["PageHtml"] = utils.NewPaginations(common.RollPage, totalCount, common.PageSize, pageIndex, beego.URLFor("BookController.Index"), fmt.Sprintf("&private=%v", private))
	} else {
		c.Data["PageHtml"] = ""
	}
	//封面图片
	for idx, book := range books {
		//book.Cover = utils.ShowImg(book.Cover, "cover")
		books[idx] = book
	}
	b, err := json.Marshal(books)
	if err != nil || len(books) <= 0 {
		c.Data["Result"] = template.JS("[]")
	} else {
		c.Data["Result"] = template.JS(string(b))
	}
	c.Data["Private"] = private
	c.TplName = "book/index.html"
}

// 设置图书页面
func (c *BookController) Setting() {

	key := c.Ctx.Input.Param(":key")

	if key == "" {
		c.Abort("404")
	}

	book, err := models.NewBookData().SelectByIdentify(key, c.Member.MemberId)
	if err != nil && err != orm.ErrNoRows {
		c.Abort("404")
	}

	//需管理员以上权限
	if book.RoleId != common.BookFounder && book.RoleId != common.BookAdmin {
		c.Abort("404")
	}

	if book.PrivateToken != "" {
		book.PrivateToken = c.BaseUrl() + beego.URLFor("DocumentController.Index", ":key", book.Identify, "token", book.PrivateToken)
	}

	//查询图书分类
	if selectedCates, rows, _ := new(models.BookCategory).SelectByBookId(book.BookId); rows > 0 {
		var maps = make(map[int]bool)
		for _, cate := range selectedCates {
			maps[cate.Id] = true
		}
		c.Data["Maps"] = maps
	}

	c.Data["Cates"], _ = new(models.Category).GetCates(-1, 1)
	c.Data["Model"] = book
	c.TplName = "book/setting.html"
}

//保存图书信息
func (c *BookController) SaveBook() {

	bookResult, err := c.isPermission()
	if err != nil {
		c.JsonResult(1, err.Error())
	}

	book, err := models.NewBook().Select("book_id", bookResult.BookId)
	if err != nil {
		logs.Error("SaveBook => ", err)
		c.JsonResult(1, err.Error())
	}

	bookName := strings.TrimSpace(c.GetString("book_name"))
	description := strings.TrimSpace(c.GetString("description", ""))
	editor := strings.TrimSpace(c.GetString("editor"))

	if strings.Count(description, "") > 500 {
		c.JsonResult(1, "描述需小于500字")
	}

	if editor != "markdown" && editor != "html" {
		editor = "markdown"
	}

	book.BookName = bookName
	book.Description = description
	book.Editor = editor
	book.Author = c.GetString("author")
	book.AuthorURL = c.GetString("author_url")

	if err := book.Update(); err != nil {
		c.JsonResult(1, "保存失败")
	}
	bookResult.BookName = bookName
	bookResult.Description = description

	//Update分类
	if cids, ok := c.Ctx.Request.Form["cid"]; ok {
		new(models.BookCategory).SetBookCates(book.BookId, cids)
	}

	c.JsonResult(0, "ok", bookResult)
}

//上传封面.
func (c *BookController) UploadCover() {
	bookResult, err := c.isPermission()
	if err != nil {
		c.JsonResult(1, err.Error())
	}

	book, err := models.NewBook().Select("book_id", bookResult.BookId)
	if err != nil {
		c.JsonResult(1, err.Error())
	}

	file, moreFile, err := c.GetFile("image-file")
	if err != nil {
		logs.Error("", err.Error())
		c.JsonResult(1, "读取文件异常")
	}

	defer file.Close()

	ext := filepath.Ext(moreFile.Filename)

	if !strings.EqualFold(ext, ".png") && !strings.EqualFold(ext, ".jpg") && !strings.EqualFold(ext, ".gif") && !strings.EqualFold(ext, ".jpeg") {
		c.JsonResult(1, "不支持图片格式")
	}

	x1, _ := strconv.ParseFloat(c.GetString("x"), 10)
	y1, _ := strconv.ParseFloat(c.GetString("y"), 10)
	w1, _ := strconv.ParseFloat(c.GetString("width"), 10)
	h1, _ := strconv.ParseFloat(c.GetString("height"), 10)

	x := int(x1)
	y := int(y1)
	width := int(w1)
	height := int(h1)

	fileName := strconv.FormatInt(time.Now().UnixNano(), 16)

	filePath := filepath.Join("uploads", time.Now().Format("200601"), fileName+ext)

	path := filepath.Dir(filePath)

	os.MkdirAll(path, os.ModePerm)

	err = c.SaveToFile("image-file", filePath)

	if err != nil {
		logs.Error("", err)
		c.JsonResult(1, "保存图片失败")
	}

	//剪切图片
	subImg, err := graphics.ImageCopyFromFile(filePath, x, y, width, height)
	if err != nil {
		c.JsonResult(1, "图片剪切")
	}

	filePath = filepath.Join(common.WorkingDirectory, "uploads", time.Now().Format("200601"), fileName+ext)

	//生成缩略图
	err = graphics.ImageResizeSaveFile(subImg, 175, 230, filePath)
	if err != nil {
		c.JsonResult(1, "保存图片失败")
	}

	url := "/" + strings.Replace(strings.TrimPrefix(filePath, common.WorkingDirectory), "\\", "/", -1)
	if strings.HasPrefix(url, "//") {
		url = string(url[1:])
	}
	book.Cover = url

	if err := book.Update(); err != nil {
		c.JsonResult(1, "保存图片失败")
	}

	save := book.Cover
	if err := store.SaveToLocal("."+url, save); err != nil {
		beego.Error(err.Error())
	} else {
		url = book.Cover
	}
	c.JsonResult(0, "ok", url)
}

//创建图书
func (c *BookController) Create() {
	identify := strings.TrimSpace(c.GetString("identify", ""))
	bookName := strings.TrimSpace(c.GetString("book_name", ""))
	author := strings.TrimSpace(c.GetString("author", ""))
	authorURL := strings.TrimSpace(c.GetString("author_url", ""))
	privatelyOwned, _ := c.GetInt("privately_owned")
	description := strings.TrimSpace(c.GetString("description", ""))

	/*
	* 约束条件判断
	 */
	if identify == "" || strings.Count(identify, "") > 50 {
		c.JsonResult(1, "请正确填写图书标识，不能超过50字")
	}
	if bookName == "" {
		c.JsonResult(1, "请填图书名称")
	}

	if strings.Count(description, "") > 500 {
		c.JsonResult(1, "图书描述需小于500字")
	}

	if privatelyOwned != 0 && privatelyOwned != 1 {
		privatelyOwned = 1
	}

	book := models.NewBook()
	if book, _ := book.Select("identify", identify); book.BookId > 0 {
		c.JsonResult(1, "identify冲突")
	}

	book.BookName = bookName
	book.Identify = identify
	book.Description = description
	book.CommentCount = 0
	book.PrivatelyOwned = privatelyOwned
	book.Cover = common.DefaultCover()
	book.DocCount = 0
	book.MemberId = c.Member.MemberId
	book.CommentCount = 0
	book.Editor = "markdown"
	book.ReleaseTime = time.Now()
	book.Score = 40 //评分
	book.Author = author
	book.AuthorURL = authorURL

	if err := book.Insert(); err != nil {
		c.JsonResult(1, "数据库错误")
	}

	bookResult, err := models.NewBookData().SelectByIdentify(book.Identify, c.Member.MemberId)
	if err != nil {
		beego.Error(err)
	}

	c.JsonResult(0, "ok", bookResult)
}

//发布图书.
func (c *BookController) Release() {
	identify := c.GetString("identify")
	bookId := 0
	if c.Member.IsAdministrator() {
		book, err := models.NewBook().Select("identify", identify)
		if err != nil {
			beego.Error(err)
		}
		bookId = book.BookId
	} else {
		book, err := models.NewBookData().SelectByIdentify(identify, c.Member.MemberId)
		if err != nil {
			c.JsonResult(1, "未知错误")
		}
		if book.RoleId != common.BookAdmin && book.RoleId != common.BookFounder && book.RoleId != common.BookEditor {
			c.JsonResult(1, "权限不足")
		}
		bookId = book.BookId
	}

	if exist := utils.BooksRelease.Exist(bookId); exist {
		c.JsonResult(1, "正在发布中，请稍后操作")
	}

	go func(identify string) {
		models.NewDocument().ReleaseContent(bookId, c.BaseUrl())
	}(identify)

	c.JsonResult(0, "已发布")
}

func (c *BookController) isPermission() (*models.BookData, error) {

	identify := c.GetString("identify")

	book, err := models.NewBookData().SelectByIdentify(identify, c.Member.MemberId)
	if err != nil {
		return book, err
	}

	if book.RoleId != common.BookAdmin && book.RoleId != common.BookFounder {
		return book, errors.New("权限不足")
	}
	return book, nil
}

//收藏
func (c *BookController) Collection() {
	uid := c.BaseController.Member.MemberId
	if uid <= 0 {
		c.JsonResult(1, "收藏失败，请先登录")
	}

	id, _ := c.GetInt(":id")
	if id <= 0 {
		c.JsonResult(1, "收藏失败，图书不存在")
	}

	cancel, err := new(models.Collection).Collection(uid, id)
	data := map[string]bool{"IsCancel": cancel}
	if err != nil {
		beego.Error(err.Error())
		if cancel {
			c.JsonResult(1, "取消收藏失败", data)
		}
		c.JsonResult(1, "添加收藏失败", data)
	}

	if cancel {
		c.JsonResult(0, "取消收藏成功", data)
	}
	c.JsonResult(0, "添加收藏成功", data)
}

//打分
func (c *BookController) Score() {
	bookId, _ := c.GetInt(":id")
	if bookId == 0 {
		c.JsonResult(1, "文档不存在")
	}

	score, _ := c.GetInt("score")
	if uid := c.Member.MemberId; uid > 0 {
		if err := new(models.Score).AddScore(uid, bookId, score); err != nil {
			c.JsonResult(1, err.Error())
		}
		c.JsonResult(0, "感谢您给当前文档打分")
	}
	c.JsonResult(1, "给文档打分失败，请先登录再操作")
}

//评论
func (c *BookController) Comment() {
	if c.Member.MemberId == 0 {
		c.JsonResult(1, "请先登录在评论")
	}
	content := c.GetString("content")
	if l := len(content); l < 5 || l > 512 {
		c.JsonResult(1, "评论内容允许5-512个字符")
	}
	bookId, _ := c.GetInt(":id")
	if bookId > 0 {
		// todo 这个地方应该使用事务，其中任何一个发生错误都需要回滚
		if err := new(models.Comments).AddComments(c.Member.MemberId, bookId, content); err != nil {
			c.JsonResult(1, err.Error())
		}
		// 评论数+1
		if err := models.IncOrDec(models.TNBook(), "cnt_comment", fmt.Sprintf("book_id=%v", bookId), true, 1); err != nil {
			c.JsonResult(1, err.Error())
		}

		c.JsonResult(0, "评论成功")
	}
	c.JsonResult(1, "文档图书不存在")
}

//私有图书创建访问Token
func (c *BookController) CreateToken() {
	action := c.GetString("action")
	bookResult, err := c.isPermission()
	if err != nil {
		c.JsonResult(1, err.Error())
	}

	fmt.Println(bookResult.BookId)

	book := models.NewBook()
	if _, err := book.Select("book_id", bookResult.BookId); err != nil {
		c.JsonResult(1, "图书不存在")
	}

	if action == "create" {
		if bookResult.PrivatelyOwned == 0 {
			c.JsonResult(1, "公开图书不能创建令牌")
		}

		book.PrivateToken = string(utils.Krand(12, utils.KC_RAND_KIND_ALL))
		if err := book.Update(); err != nil {
			c.JsonResult(1, "生成阅读失败")
		}
		c.JsonResult(0, "ok", c.BaseUrl()+beego.URLFor("DocumentController.Index", ":key", book.Identify, "token", book.PrivateToken))
	}

	book.PrivateToken = ""
	if err := book.Update(); err != nil {
		c.JsonResult(1, "删除令牌失败")
	}
	c.JsonResult(0, "ok", "")
}

//上传项目
func (this *BookController) UploadProject() {
	//处理步骤
	//1、接受上传上来的zip文件，并存放到store/temp目录下
	//2、解压zip到当前目录，然后移除非图片文件
	//3、将文件夹移动到uploads目录下

	identify := this.GetString("identify")

	if !models.NewBook().HasProjectAccess(identify, this.Member.MemberId, conf.BookEditor) {
		this.JsonResult(1, "无操作权限")
	}

	book, _ := models.NewBookData().FindByIdentify(identify, this.Member.MemberId)
	if book.BookId == 0 {
		this.JsonResult(1, "项目不存在")
	}

	f, h, err := this.GetFile("zipfile")
	if err != nil {
		this.JsonResult(1, err.Error())
	}
	defer f.Close()
	if strings.ToLower(filepath.Ext(h.Filename)) != ".zip" && strings.ToLower(filepath.Ext(h.Filename)) != ".epub" {
		this.JsonResult(1, "请上传指定格式文件")
	}
	tmpFile := "store/" + identify + ".zip" //保存的文件名
	if err := this.SaveToFile("zipfile", tmpFile); err == nil {
		go this.unzipToData(book.BookId, identify, tmpFile, h.Filename)
	} else {
		beego.Error(err.Error())
	}
	this.JsonResult(0, "上传成功")
}

//将zip压缩文件解压并录入数据库
//@param            book_id             项目id(其实有想不标识了可以不要这个的，但是这里的项目标识只做目录)
//@param            identify            项目标识
//@param            zipfile             压缩文件
//@param            originFilename      上传文件的原始文件名
func (this *BookController) unzipToData(bookId int, identify, zipFile, originFilename string) {

	//说明：
	//OSS中的图片存储规则为"projects/$identify/项目中图片原路径"
	//本地存储规则为"uploads/projects/$identify/项目中图片原路径"

	projectRoot := "" //项目根目录

	//解压目录
	unzipPath := "store/" + identify

	//如果存在相同目录，则率先移除
	if err := os.RemoveAll(unzipPath); err != nil {
		beego.Error(err.Error())
	}
	os.MkdirAll(unzipPath, os.ModePerm)

	imgMap := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".svg": true, ".webp": true}

	defer func() {
		os.Remove(zipFile)      //最后删除上传的临时文件
		os.RemoveAll(unzipPath) //删除解压后的文件夹
	}()

	//注意：这里的prefix必须是判断是否是GitHub之前的prefix
	if err := ziptil.Unzip(zipFile, unzipPath); err != nil {
		beego.Error("解压失败", zipFile, err.Error())
		return
	}

	//从获取SUMMARY
	summary := mdtil.SummaryToMap(unzipPath)
	//fmt.Println(summary)

	//读取文件，把图片文档录入oss
	if files, err := filetil.ScanFiles(unzipPath); err == nil {
		projectRoot = this.getProjectRoot(files)
		// 兼容
		projectRoot = strings.Replace(projectRoot, "\\", "/", -1)

		// 替换[img] [a]
		this.replaceToAbs(projectRoot, identify)

		//ModelStore := new(models.DocumentStore)
		//文档对应的标识
		for _, file := range files {
			if !file.IsDir {
				ext := strings.ToLower(filepath.Ext(file.Path))
				if ok, _ := imgMap[ext]; ok { //图片，录入oss
					switch utils.StoreType {
					//case utils.StoreOss:
					//	if err := store.ModelStoreOss.MoveToOss(file.Path, "projects/"+identify+strings.TrimPrefix(file.Path, projectRoot), false, false); err != nil {
					//		beego.Error(err)
					//	}
					case utils.StoreLocal:
						//if err := store.ModelStoreLocal.MoveToStore(file.Path, "uploads/projects/"+identify+strings.TrimPrefix(file.Path, projectRoot)); err != nil {
						if err := store.MoveToStore(file.Path, "uploads/projects/"+identify+strings.TrimPrefix(file.Path, projectRoot)); err != nil {
							beego.Error(err)
						}
					}
				} else if ext == ".md" || ext == ".markdown" || ext == ".html" { //markdown文档，提取文档内容，录入数据库
					tmpIdentify := strings.Replace(strings.Trim(strings.TrimPrefix(file.Path, projectRoot), "/"), "/", "-", -1)
					tmpIdentify = strings.Replace(tmpIdentify, ")", "", -1)

					//为后面重新上传就会自动更新提供依据，
					doc, _ := models.NewDocument().SelectByIdentify(bookId, tmpIdentify)

					var mdcont string
					var htmlStr string
					if b, err := ioutil.ReadFile(file.Path); err == nil {
						if ext == ".md" || ext == ".markdown" {
							mdcont = strings.TrimSpace(string(b))
							htmlStr = mdtil.Md2html(mdcont)
						} else {
							htmlStr = string(b)
							mdcont = html2md.Convert(htmlStr)
						}
						if !strings.HasPrefix(mdcont, "[TOC]") {
							mdcont = "[TOC]\r\n\r\n" + mdcont
						}
						// 页面上看到的内容
						doc.Release = htmlStr
						// 从summary中获取name，如果是获取不到就从文中
						DocumentName := summary[strings.Trim(strings.TrimPrefix(file.Path, projectRoot), "/")]
						if DocumentName == "" {
							doc.DocumentName = utils.ParseTitleFromMdHtml(htmlStr)
						} else {
							doc.DocumentName = DocumentName
						}

						doc.BookId = bookId
						//文档标识
						//doc.Identify = strings.Replace(strings.Trim(strings.TrimPrefix(file.Path, projectRoot), "/"), "/", "-", -1)
						//doc.Identify = strings.Replace(doc.Identify, ")", "", -1)
						doc.Identify = tmpIdentify

						doc.MemberId = this.Member.MemberId
						doc.OrderSort = 1
						if strings.HasSuffix(strings.ToLower(file.Name), "summary.md") {
							doc.OrderSort = 0
						}
						if strings.HasSuffix(strings.ToLower(file.Name), "summary.html") {
							mdcont += "<bookstack-summary></bookstack-summary>"
							// 生成带$的文档标识，阅读BaseController.go代码可知，
							// 要使用summary.md的排序功能，必须在链接中带上符号$
							mdcont = strings.Replace(mdcont, "](", "]($", -1)
							// 去掉可能存在的url编码的右括号，否则在url译码后会与markdown语法混淆
							mdcont = strings.Replace(mdcont, "%29", "", -1)
							mdcont, _ = url.QueryUnescape(mdcont)
							doc.OrderSort = 0
							doc.Identify = "summary.md"
						}
						if docId, err := doc.InsertOrUpdate("document_name", "release", "vcnt"); err == nil {
							// 写入Content，后面的上线需要从content中复制到release
							ds := models.DocumentStore{DocumentId: int(docId), Markdown: mdcont, Content: htmlStr}
							if err := ds.InsertOrUpdate("markdown", "content"); err != nil {
								//if err := ModelStore.InsertOrUpdate(models.DocumentStore{DocumentId: int(docId),Markdown:   mdcont,}, "markdown"); err != nil {
								//if err := ModelStore.InsertOrUpdate(models.DocumentStore{DocumentId: int(docId),Markdown:   mdcont,}, "markdown"); err != nil {
								beego.Error(err)
							}
						} else {
							beego.Error(err.Error())
						}
					} else {
						beego.Error("读取文档失败：", file.Path, "错误信息：", err)
					}

				}
			}
		}
	}
}

//获取文档项目的根目录
func (this *BookController) getProjectRoot(fl []filetil.FileList) (root string) {
	//获取项目的根目录(感觉这个函数封装的不是很好，有更好的方法，请通过issue告知我，谢谢。)
	i := 1000
	for _, f := range fl {
		if !f.IsDir {
			if cnt := strings.Count(f.Path, "/"); cnt < i {
				root = filepath.Dir(f.Path)
				i = cnt
			}
		}
	}
	return
}

//查找并替换markdown文件中的路径，把图片链接替换成url的相对路径，把文档间的链接替换成【$+文档标识链接】
func (this *BookController) replaceToAbs(projectRoot string, identify string) {
	imgBaseUrl := "/uploads/projects/" + identify
	switch utils.StoreType {
	case utils.StoreLocal:
		// 添加store
		imgBaseUrl = "/uploads/projects/" + identify
	case utils.StoreOss:
		//imgBaseUrl = this.BaseController.OssDomain + "/projects/" + identify
		imgBaseUrl = "/projects/" + identify
	}
	files, _ := filetil.ScanFiles(projectRoot)
	for _, file := range files {
		if ext := strings.ToLower(filepath.Ext(file.Path)); ext == ".md" || ext == ".markdown" {
			//mdb ==> markdown byte
			mdb, _ := ioutil.ReadFile(file.Path)
			mdCont := string(mdb)
			basePath := filepath.Dir(file.Path)
			basePath = strings.Trim(strings.Replace(basePath, "\\", "/", -1), "/")
			basePathSlice := strings.Split(basePath, "/")
			l := len(basePathSlice)
			b, _ := ioutil.ReadFile(file.Path)
			output := blackfriday.Run(b)
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(string(output)))

			//图片链接处理
			doc.Find("img").Each(func(i int, selection *goquery.Selection) {
				//非http开头的图片地址，即是相对地址
				if src, ok := selection.Attr("src"); ok && !strings.HasPrefix(strings.ToLower(src), "http") {
					newSrc := src //默认为旧地址
					if cnt := strings.Count(src, "../"); cnt < l { //以或者"../"开头的路径
						newSrc = strings.Join(basePathSlice[0:l-cnt], "/") + "/" + strings.TrimLeft(src, "./")
					}
					newSrc = imgBaseUrl + "/" + strings.TrimLeft(strings.TrimPrefix(strings.TrimLeft(newSrc, "./"), projectRoot), "/")
					//newSrc = imgBaseUrl + strings.TrimLeft(strings.TrimPrefix(strings.TrimLeft(newSrc, "./"), projectRoot), "/")
					mdCont = strings.Replace(mdCont, src, newSrc, -1)
				}
			})

			//a标签链接处理。要注意判断有锚点的情况
			doc.Find("a").Each(func(i int, selection *goquery.Selection) {
				if href, ok := selection.Attr("href"); ok && !strings.HasPrefix(strings.ToLower(href), "http") && !strings.HasPrefix(href, "#") {
					newHref := href //默认
					if cnt := strings.Count(href, "../"); cnt < l {
						newHref = strings.Join(basePathSlice[0:l-cnt], "/") + "/" + strings.TrimLeft(href, "./")
					}
					newHref = strings.TrimPrefix(strings.Trim(newHref, "/"), projectRoot)
					if !strings.HasPrefix(href, "$") { //原链接不包含$符开头，否则表示已经替换过了。
						newHref = "$" + strings.Replace(strings.Trim(newHref, "/"), "/", "-", -1)
						slice := strings.Split(newHref, "$")
						if ll := len(slice); ll > 0 {
							newHref = "$" + slice[ll-1]
						}
						mdCont = strings.Replace(mdCont, "]("+href, "]("+newHref, -1)
					}
				}
			})
			ioutil.WriteFile(file.Path, []byte(mdCont), os.ModePerm)
		}
	}
}

// Delete 删除项目.
func (this *BookController) Delete() {

	bookResult, err := this.IsPermission()
	if err != nil {
		this.JsonResult(6001, err.Error())
	}

	if bookResult.RoleId != conf.BookFounder {
		this.JsonResult(6002, "只有创始人才能删除项目")
	}

	//用户密码
	pwd := this.GetString("password")
	if m, err := models.NewMember().Login(this.Member.Account, pwd); err != nil || m.MemberId == 0 {
		this.JsonResult(1, "项目删除失败，您的登录密码不正确")
	}

	err = models.NewBook().ThoroughDeleteBook(bookResult.BookId)
	if err == orm.ErrNoRows {
		this.JsonResult(6002, "项目不存在")
	}

	if err != nil {
		logs.Error("删除项目 => ", err)
		this.JsonResult(6003, "删除失败")
	}

	//go func() {
	//	client := models.NewElasticSearchClient()
	//	if errDel := client.DeleteIndex(bookResult.BookId, true); errDel != nil && client.On {
	//		beego.Error(errDel.Error())
	//	}
	//}()

	this.JsonResult(0, "ok")
}

// 判断是否具有管理员或管理员以上权限
func (this *BookController) IsPermission() (*models.BookData, error) {

	identify := this.GetString("identify")

	book, err := models.NewBookData().FindByIdentify(identify, this.Member.MemberId)

	if err != nil {
		if err == models.ErrPermissionDenied {
			return book, errors.New("权限不足")
		}
		if err == orm.ErrNoRows {
			return book, errors.New("项目不存在")
		}
		return book, err
	}

	if book.RoleId != conf.BookAdmin && book.RoleId != conf.BookFounder {
		return book, errors.New("权限不足")
	}
	return book, nil
}
