package sysinit

import (
	"github.com/astaxie/beego"
	"mbook/models"
	"mbook/utils"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

func sysinit() {
	uploads := filepath.Join("./", "uploads")
	beego.BConfig.WebConfig.StaticDir["/uploads"] = uploads
	//注册前端使用函数
	registerFuncionts()
}

func registerFuncionts() {
	beego.AddFuncMap("cdnjs", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdnjs", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		return cdn + p
	})
	beego.AddFuncMap("cdncss", func(p string) string {
		cdn := beego.AppConfig.DefaultString("cdncss", "")
		if strings.HasPrefix(p, "/") && strings.HasSuffix(cdn, "/") {
			return cdn + string(p[1:])
		}
		if !strings.HasPrefix(p, "/") && !strings.HasSuffix(cdn, "/") {
			return cdn + "/" + p
		}
		return cdn + p
	})
	beego.AddFuncMap("getUsernameByUid", func(id interface{}) string {
		return new(models.Member).GetUsernameByUid(id)
	})
	beego.AddFuncMap("getNicknameByUid", func(id interface{}) string {
		return new(models.Member).GetNicknameByUid(id)
	})
	beego.AddFuncMap("inMap", utils.InMap)

	//	//用户是否收藏了文档
	beego.AddFuncMap("doesCollection", new(models.Collection).DoesCollection)
	beego.AddFuncMap("scoreFloat", utils.ScoreFloat)
	beego.AddFuncMap("showImg", utils.ShowImg)
	beego.AddFuncMap("IsFollow", new(models.Fans).Relation)
	beego.AddFuncMap("isubstr", utils.Substr)
	beego.AddFuncMap("date_format", func(t time.Time, format string) string {
		return t.Local().Format(format)
	})
	beego.AddFuncMap("encodeURIComponent", func(str string) string {
		r := url.QueryEscape(str)
		return r
	})
}
