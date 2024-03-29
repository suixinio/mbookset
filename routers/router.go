package routers

import (
	"github.com/astaxie/beego"
	"mbook/controllers"
	"mbook/controllers/so"
)

func init() {
	//首页&分类&详情
	beego.Router("/", &controllers.HomeController{}, "get:Index")
	beego.Router("/explore", &controllers.ExploreController{}, "*:Index")
	beego.Router("/book/:key", &controllers.DocumentController{}, "*:Index")

	//读书
	beego.Router("/read/:key/:id", &controllers.DocumentController{}, "*:Read")
	beego.Router("/read/:key/search", &controllers.DocumentController{}, "post:Search")

	//搜索
	beego.Router("/search", &controllers.SearchController{}, "get:Search")
	beego.Router("/search/result", &controllers.SearchController{}, "get:Result")

	//login
	beego.Router("/login", &controllers.AccountController{}, "*:Login")
	beego.Router("/regist", &controllers.AccountController{}, "*:Regist")
	beego.Router("/logout", &controllers.AccountController{}, "*:Logout")
	beego.Router("/doregist", &controllers.AccountController{}, "post:DoRegist")

	//编辑
	beego.Router("/api/:key/edit/?:id", &controllers.DocumentController{}, "*:Edit")
	beego.Router("/api/:key/content/?:id", &controllers.DocumentController{}, "*:Content")
	beego.Router("/api/upload", &controllers.DocumentController{}, "post:Upload")
	beego.Router("/api/:key/create", &controllers.DocumentController{}, "post:Create")
	beego.Router("/api/:key/delete", &controllers.DocumentController{}, "post:Delete")

	//用户图书管理
	beego.Router("/books", &controllers.BookController{}, "*:Index")                         //我的图书
	beego.Router("/book/create", &controllers.BookController{}, "post:Create")               //创建图书
	beego.Router("/book/:key/setting", &controllers.BookController{}, "*:Setting")           //图书设置
	beego.Router("/book/:key/sort", &controllers.BookController{}, "post:SaveSort")          //目录排序
	beego.Router("/book/:key/replace", &controllers.BookController{}, "get,post:Replace")    //全文替换
	beego.Router("/book/setting/upload", &controllers.BookController{}, "post:UploadCover")  //图书封面
	beego.Router("/book/setting/open", &controllers.BookController{}, "post:PrivatelyOwned")
	beego.Router("/book/star/:id", &controllers.BookController{}, "*:Collection")            //收藏图书
	beego.Router("/book/setting/save", &controllers.BookController{}, "post:SaveBook")       //保存
	beego.Router("/book/:key/release", &controllers.BookController{}, "post:Release")        //发布
	beego.Router("/book/setting/token", &controllers.BookController{}, "post:CreateToken")   //创建Token
	beego.Router("/book/uploadProject", &controllers.BookController{}, "post:UploadProject") //用户上传
	beego.Router("/book/setting/delete", &controllers.BookController{}, "post:Delete")       //删除book
	//个人中心
	beego.Router("/user/:username", &controllers.UserController{}, "get:Index")                 //分享
	beego.Router("/user/:username/collection", &controllers.UserController{}, "get:Collection") //收藏
	beego.Router("/user/:username/follow", &controllers.UserController{}, "get:Follow")         //关注
	beego.Router("/user/:username/fans", &controllers.UserController{}, "get:Fans")             //粉丝
	beego.Router("/follow/:uid", &controllers.BaseController{}, "get:SetFollow")                //关注或取消关注
	beego.Router("/book/score/:id", &controllers.BookController{}, "*:Score")                   //评分
	beego.Router("/book/comment/:id", &controllers.BookController{}, "post:Comment")            //评论

	//个人设置
	beego.Router("/setting", &controllers.SettingController{}, "*:Index")
	beego.Router("/setting/upload", &controllers.SettingController{}, "*:Upload")
	//管理后台
	beego.Router("/manager/category", &controllers.ManagerController{}, "post,get:Category")
	beego.Router("/manager/update-cate", &controllers.ManagerController{}, "get:UpdateCate")
	beego.Router("/manager/del-cate", &controllers.ManagerController{}, "get:DelCate")
	beego.Router("/manager/icon-cate", &controllers.ManagerController{}, "post:UpdateCateIcon")

	//文章

	//管理文章的路由
	beego.Router("/manage/blogs", &controllers.BlogController{},"*:ManageList")
	//beego.Router("/manage/blogs/setting/?:id", &controllers.BlogController{}, "*:ManageSetting")
	//beego.Router("/manage/blogs/edit/?:id",&controllers.BlogController{}, "*:ManageEdit")
	//beego.Router("/manage/blogs/delete",&controllers.BlogController{}, "post:ManageDelete")
	//beego.Router("/manage/blogs/upload",&controllers.BlogController{}, "post:Upload")
	//beego.Router("/manage/blogs/attach/:id",&controllers.BlogController{}, "post:RemoveAttachment")


	//读文章的路由
	beego.Router("/blogs", &controllers.BlogController{}, "*:List")
	//beego.Router("/blog-attach/:id:int/:attach_id:int", &controllers.BlogController{},"get:Download")
	beego.Router("/blog-:id([0-9]+).html", &controllers.BlogController{}, "*:Index")
	beego.Router("/crawl", &controllers.BaseController{}, "post:Crawl") //爬取

	// stackoverflow 汉化
	beego.Router("/zh-so", &so.IndexController{}, "get:Index")
	beego.Router("/questions/:questionID", &so.QuestionController{}, "get:ShowQuestionAction")
	beego.Router("/tags/*", &so.TagController{}, "get:ShowTagAction")
	beego.Router("/so/questions", &so.SOController{}, "get:SpiderOSQuestions")
	beego.Router("/so/answers", &so.SOController{}, "get:SpiderOSAnswers")
	beego.Router("/so/translate", &so.SOController{}, "get:GoogleTranslate")
}
