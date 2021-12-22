package tests_test

import (
	"context"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/server"
	immudb "github.com/codenotary/immugorm"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	. "gorm.io/gorm/utils/tests"
)


func SetUp() (*gorm.DB, func(), error){
	dir, err := ioutil.TempDir(os.TempDir(), "immuDB_test")
	if err != nil {
		return nil, nil, err
	}
	/*sessOptions := &sessions.Options{
		SessionGuardCheckInterval: time.Second * 1,
		MaxSessionInactivityTime:  time.Second * 2000,
		MaxSessionAgeTime:         time.Second * 4000,
		Timeout:                   time.Second * 2000,
	}*/
	options := server.DefaultOptions().
		WithMetricsServer(false).
		WithWebServer(false).
		WithPgsqlServer(false).
		WithPort(0).
	    WithDir(dir)//.WithSessionOptions(sessOptions)
	srv := server.DefaultServer().WithOptions(options).(*server.ImmuServer)
	err = srv.Initialize()
	if err != nil {
		return nil, nil, err
	}
	go func() {
		err := srv.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	log.Println("testing immuDB...")

	opts := client.DefaultOptions().WithPort(srv.Listener.Addr().(*net.TCPAddr).Port)

	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"

	gormDB, err := gorm.Open(immudb.Open(opts, &immudb.ImmuGormConfig{Verify: false}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	closeDB := func() {
		sqlDB, er := gormDB.DB()
		if er != nil {
			panic(er)
		}
		conn, er := sqlDB.Conn(context.TODO())
		if er != nil && er.Error() != "sql: database is closed" {
			panic(er)
		}
		if conn != nil {
			er = conn.Close()
			if er != nil {
				panic(er)
			}
		}
		if sqlDB != nil {
			er = sqlDB.Close()
			if er != nil {
				panic(er)
			}
		}
		gormDB = nil
		srv.Stop()
		srv = nil
		os.RemoveAll(options.Dir)
		os.Remove(".state-")
		runtime.GC()

	}

	allModels := []interface{}{&User{}, &Account{}, &Pet{}, &Company{}, &Toy{}, &Language{}, &Coupon{}/*, &CouponProduct{}*/, &Order{}}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allModels), func(i, j int) { allModels[i], allModels[j] = allModels[j], allModels[i] })

	if err = gormDB.AutoMigrate(allModels...); err != nil {
		log.Printf("Failed to auto migrate, but got error %v\n", err)
		os.Exit(1)
	}

	/*for _, m := range allModels {
		if !gorm.Migrator().HasTable(m) {
			log.Printf("Failed to create table for %#v\n", m)
			os.Exit(1)
		}
	}*/

	if debug := os.Getenv("DEBUG"); debug == "true" {
		gormDB.Logger = gormDB.Logger.LogMode(logger.Info)
	} else if debug == "false" {
		gormDB.Logger = gormDB.Logger.LogMode(logger.Silent)
	}

	return gormDB, closeDB, err
}

var DB *gorm.DB

func init() {
	if os.Getenv("GORM_DIALECT") == "immuDB"{
		return
	}
	var err error
	if DB, err = OpenTestConnection(); err != nil {
		log.Printf("failed to connect database, got error %v", err)
		os.Exit(1)
	} else {
		sqlDB, err := DB.DB()
		if err == nil {
			err = sqlDB.Ping()
		}

		if err != nil {
			log.Printf("failed to connect database, got error %v", err)
		}

		RunMigrations()
		if DB.Dialector.Name() == "sqlite" {
			DB.Exec("PRAGMA foreign_keys = ON")
		}
	}
}

func OpenTestConnection() (DB *gorm.DB, err error) {
	DBDSN := os.Getenv("GORM_DSN")
	switch os.Getenv("GORM_DIALECT") {
	case "mysql":
		log.Println("testing mysql...")
		if DBDSN == "" {
			DBDSN = "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local"
		}
		DB, err = gorm.Open(mysql.Open(DBDSN), &gorm.Config{})
	case "postgres":
		log.Println("testing postgres...")
		if DBDSN == "" {
			DBDSN = "user=gorm password=gorm DBname=gorm host=localhost port=5432 sslmode=disable TimeZone=Asia/Shanghai"
		}
		DB, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  DBDSN,
			PreferSimpleProtocol: true,
		}), &gorm.Config{})
	case "sqlserver":
		// CREATE LOGIN gorm WITH PASSWORD = 'LoremIpsum86';
		// CREATE DATABASE gorm;
		// USE gorm;
		// CREATE USER gorm FROM LOGIN gorm;
		// sp_changeDBowner 'gorm';
		// npm install -g sql-cli
		// mssql -u gorm -p LoremIpsum86 -d gorm -o 9930
		log.Println("testing sqlserver...")
		if DBDSN == "" {
			DBDSN = "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
		}
		DB, err = gorm.Open(sqlserver.Open(DBDSN), &gorm.Config{})
	default:
		log.Println("testing sqlite3...")
		DB, err = gorm.Open(sqlite.Open(filepath.Join(os.TempDir(), "gorm.DB")), &gorm.Config{})
	}

	if debug := os.Getenv("DEBUG"); debug == "true" {
		DB.Logger = DB.Logger.LogMode(logger.Info)
	} else if debug == "false" {
		DB.Logger = DB.Logger.LogMode(logger.Silent)
	}

	return
}

func RunMigrations() {
	var err error
	allModels := []interface{}{&User{}, &Account{}, &Pet{}, &Company{}, &Toy{}, &Language{}, &Coupon{}, &CouponProduct{}, &Order{}}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allModels), func(i, j int) { allModels[i], allModels[j] = allModels[j], allModels[i] })

	DB.Migrator().DropTable("user_friends", "user_speaks")

	if err = DB.Migrator().DropTable(allModels...); err != nil {
		log.Printf("Failed to drop table, got error %v\n", err)
		os.Exit(1)
	}

	if err = DB.AutoMigrate(allModels...); err != nil {
		log.Printf("Failed to auto migrate, but got error %v\n", err)
		os.Exit(1)
	}

	for _, m := range allModels {
		if !DB.Migrator().HasTable(m) {
			log.Printf("Failed to create table for %#v\n", m)
			os.Exit(1)
		}
	}
}
