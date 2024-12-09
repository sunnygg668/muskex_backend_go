package utils

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

// var DomainPidMap = map[string]int{}
//
//	func LoadMerchantDbs(dbs []string) {
//		RawOrm.AutoMigrate(&go				{})
//		items := []*Merchant{}
//		//测试时先用dbs参数传入的数据库连接串
//		for i, db := range dbs {
//			if db != "" {
//				contextDbs[i] = ConnectDb(db)
//			}
//		}
//		DomainPidMap["localhost:8080"] = 1
//		DomainPidMap["127.0.0.1:8080"] = 1
//		err := RawOrm.Find(&items, "status=?", 1).Error
//		if err != nil {
//			panic(err)
//		}
//		for _, item := range items {
//			contextDbs[item.Pid] = ConnectDb(item.Dbstr)
//		}
//	}

/*
商户注册信息说明：
各商户使用相同的数据库账号密码
各商户使用的数据库名字，依次为db1 db2 db3 db4。一般程序不需要读取这个名字。可以直接在商户配置文件里写上不同的值。
*/
type Merchant struct {
	ID   int
	Name string
	//内部域名InternalDomain，依次为http://localhost:8080 http://localhost:8082 http://localhost:8084 ...需要按规则填充好，程序很可能会从库库里读取使用。可以允许订制化其它内部ip域名
	InternalDomain string
	//商户api header标识 值如 1 2 3 4 ;分别表示 1号盘 2号盘 3号盘 4号盘。
	Pid int
	//忽略数据库内可以不用填写。 会写在商户配置文件里. 公网域名。所有商户可以共用一个域名如：example.com，节省配制时间，但前端api请求需要带上pid Header参数，后端根据pid来区分商户; 也可以每个商户有自己的域名，如：x号盘.com
	PubDomain string
	//忽略数据库内可以不用填写。 dbHost dbUser DbName会写在商户配置文件里,使用相同的数据库账号密码，相同数据库名字规则，节省配置时间。
	DbName string
	//状态 1启用 2禁用
	Status    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

//// contextdb
//var contextDbs = map[int]*gorm.DB{}
//
//func Pdb(pid int) *gorm.DB {
//	if pid == 0 {
//		return Orm
//	}
//	return contextDbs[pid]
//}
//func Cdb(ctx context.Context) *gorm.DB {
//	pid, _ := ctx.Value("pid").(int)
//	return Pdb(pid)
//}

func ConnectDb(connstr string) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: connstr, DefaultStringSize: 256}), &gorm.Config{NamingStrategy: schema.NamingStrategy{TablePrefix: "ba_", SingularTable: true}})
	if err != nil {
		log.Println(err)
		panic(err)
	}
	return db.Debug()
}

func InitDb(connstr string) {
	//var connstr="root:emdata2015@tcp(192.168.9.100:3306)/dataExport"
	log.Println("connect:", connstr)
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: connstr, DefaultStringSize: 256}), &gorm.Config{NamingStrategy: schema.NamingStrategy{TablePrefix: "ba_", SingularTable: true}})
	if err != nil {
		log.Println(err)
		panic(err)
	}
	log.Println("InitGormBD,dbType", connstr)
	RawOrm = db
	Orm = db.Debug()
}

var Orm, RawOrm *gorm.DB
