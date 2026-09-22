package database

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/teslaXvip/go/go_frame/post/util"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	PostDB *gorm.DB
)

func ConnectPostDB(confDir, confFile, fileType, logDir string) {
	viper := util.InitViper(confDir, confFile, fileType)
	user := viper.GetString("post.user")
	pass := viper.GetString("post.pass")
	host := viper.GetString("post.host")
	port := viper.GetInt("post.port")
	dbname := "post"
	logFileName := viper.GetString("post.log")

	// DSN连接字符串
	DataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, host, port, dbname)

	//日志控制，打开日志文件
	logFile, _ := os.OpenFile(path.Join(logDir, logFileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)

	// 自定义GORM日志，输出到日志文件
	newLogger := logger.New(
		log.New(logFile, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond, //慢SQL阈值
			LogLevel:                  logger.Info,            //打印SQL日志级别
			IgnoreRecordNotFoundError: true,                   //忽略查询不到记录的错误
			Colorful:                  false,                  //文件输出关闭彩色
		},
	)

	// 打开mysql连接
	db, err := gorm.Open(mysql.Open(DataSourceName), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败:%v", err))
	}

	// 获取底层sqlDB，设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("获取sqlDB失败:%v", err))
	}
	// 连接池参数，可以从viper读取配置，这里写默认值
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	// 赋值全局变量
	PostDB = db
}

func PingPostDB() {
	if PostDB != nil {
		sqlDB, _ := PostDB.DB()
		sqlDB.Ping()
		slog.Info("ping post db")
	}
}

func ClosePostDb() {
	if PostDB != nil {
		sqlDb, _ := PostDB.DB()
		sqlDb.Close()
	}
}
