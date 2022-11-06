package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"github.com/spf13/viper"

	_activityHttpDelivery "github.com/bxcodec/go-clean-arch/activity/delivery/http"
	_activityHttpDeliveryMiddleware "github.com/bxcodec/go-clean-arch/activity/delivery/http/middleware"
	_activityRepo "github.com/bxcodec/go-clean-arch/activity/repository/mysql"
	_activityUcase "github.com/bxcodec/go-clean-arch/activity/usecase"
	_todoHttpDelivery "github.com/bxcodec/go-clean-arch/todo/delivery/http"
	_todoHttpDeliveryMiddleware "github.com/bxcodec/go-clean-arch/todo/delivery/http/middleware"
	_todoRepo "github.com/bxcodec/go-clean-arch/todo/repository/mysql"
	_todoUcase "github.com/bxcodec/go-clean-arch/todo/usecase"
)

func init() {
	viper.SetConfigFile(`config.json`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	if viper.GetBool(`debug`) {
		log.Println("Service RUN on DEBUG mode")
	}
}

func main() {
	dbHost := viper.GetString(`database.host`)
	dbPort := viper.GetString(`database.port`)
	dbUser := viper.GetString(`database.user`)
	dbPass := viper.GetString(`database.pass`)
	dbName := viper.GetString(`database.name`)
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	dbConn, err := sql.Open(`mysql`, dsn)

	if err != nil {
		log.Fatal(err)
	}
	err = dbConn.Ping()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	e := echo.New()
	middL := _activityHttpDeliveryMiddleware.InitMiddleware()
	toMiddl := _todoHttpDeliveryMiddleware.InitMiddleware()
	e.Use(middL.CORS)
	e.Use(toMiddl.CORS)
	todo := _todoRepo.NewMysqlTodoRepository(dbConn)
	ar := _activityRepo.NewMysqlActivityRepository(dbConn)

	timeoutContext := time.Duration(viper.GetInt("context.timeout")) * time.Second
	td := _todoUcase.NewTodoUsecase(ar, todo, timeoutContext)
	au := _activityUcase.NewArticleUsecase(ar, timeoutContext)
	_activityHttpDelivery.NewArticleHandler(e, au)
	_todoHttpDelivery.NewTodoHandler(e, td)

	log.Fatal(e.Start(viper.GetString("server.address")))
}
