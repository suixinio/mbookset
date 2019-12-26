package sysinit

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	_ "mbook/models"
)

//db_init()
func dbinit(aliases ...string) {
	isDev := ("dev" == beego.AppConfig.String("runmode"))

	if len(aliases) > 0 {
		for _, alias := range aliases {
			registrDatabase(alias)
			if "w" == alias {
				orm.RunSyncdb("default", false, isDev)
			}
		}
	} else {
		registrDatabase("w")
		orm.RunSyncdb("default", false, isDev)
	}
	
	if isDev {
		orm.Debug = isDev
	}
}

func registrDatabase(alias string) {
	if len(alias) <= 0 {
		return
	}

	dbAlias := alias //default
	if ("w" == alias || "default" == alias) {
		dbAlias = "default"
		alias = "w"
	}

	//数据库名称
	dbName := beego.AppConfig.String("db_" + alias + "_database")
	//数据库用户名
	dbUser := beego.AppConfig.String("db_" + alias + "_username")
	//数据库密码
	dbPwd := beego.AppConfig.String("db_" + alias + "_password")
	//数据库IP
	dbHost := beego.AppConfig.String("db_" + alias + "_host")
	//数据库端口
	dbPost := beego.AppConfig.String("db_" + alias + "_port")
	//root:123456@tcp(127.0.0.1:3306)/mbook?charset=utf8
	orm.RegisterDataBase(dbAlias, "mysql", dbUser+":"+
		dbPwd+"@tcp("+dbHost+":"+dbPost+")/"+dbName+"?charset=utf8mb4", 30)

}
