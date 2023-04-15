package db

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strconv"
)

var (
	SqlDb    *gorm.DB
	RedisCli *redis.Client
)

const (
	defaultSQLAdd          = "www.sunsnasserver.top:3306"
	defaultSQLUser         = "root"
	defaultSQLPswd         = "abcd1234"
	defaultSQLDataBaseName = "go_r5"
	defaultRedisAdd        = "www.sunsnasserver.top:6379"
	defaultRedisPswd       = "PuA7s^4nP*V$Ri"
	defaultRedisDB         = 0
)

func init() {
	for {
		connSentinel := 0
		initMySQL(&connSentinel)
		initRedis(&connSentinel)

		if connSentinel != 0 {
			fmt.Printf("End - 0; Retry - 1: \n")
			var command string
			_, _ = fmt.Scanln(&command)
			if command == "0" {
				os.Exit(1) //todo: UNSAFE! defer first?
			}
		} else {
			break
		}
	}

}

func initRedis(connSentinel *int) {
	redisAddr := os.Getenv("DB_REDIS_ADDR")
	if redisAddr == "" {
		log.Printf("empty DB_REDIS_ADDR? Using defaultRedisAdd \n")
		redisAddr = defaultRedisAdd
	}
	redisPswd := os.Getenv("DB_REDIS_PSWD")
	redisDB, err := strconv.Atoi(os.Getenv("DB_REDIS_DB"))
	if err != nil {
		log.Printf("Redis conn err, err: %v, using empty \n", err)
		redisDB = defaultRedisDB
	}
	if redisPswd == "" {
		log.Printf("empty DB_REDIS_DB? Using defaultRedisDB \n")
		redisPswd = defaultRedisPswd
	}
	RedisCli = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPswd,
		DB:       redisDB,
	})
	pong, err := RedisCli.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Err conn redsi! %v", err)
		*connSentinel++
	} else {
		log.Printf("Redis conn success! pong: %v", pong)
	}
}

func initMySQL(connSentinel *int) {
	sqlAddr := os.Getenv("DB_SQL_ADDR")
	sqlUser := os.Getenv("DB_SQL_USER")
	sqlPswd := os.Getenv("DB_SQL_PSWD")
	sqlDataBaseName := os.Getenv("DB_SQL_DB_NAME")
	if sqlAddr == "" {
		log.Printf("empty DB_SQL_ADDR? Using defaultSQLAdd \n")
		sqlAddr = defaultSQLAdd
	}
	if sqlUser == "" {
		log.Printf("empty DB_USER? Using defaultSQLUser \n")
		sqlUser = defaultSQLUser
	}
	if sqlPswd == "" {
		log.Printf("empty DB_PSWD? Using defaultSQLPswd \n")
		sqlPswd = defaultSQLPswd
	}
	if sqlDataBaseName == "" {
		log.Printf("empty DB_SQL_DB_NAME? Using defaultSQLDataBaseName \n")
		sqlDataBaseName = defaultSQLDataBaseName
	}
	dsn := fmt.Sprintf("%v:%v@tcp(%v)/%v?charset=utf8mb3&parseTime=True&loc=Local", sqlUser, sqlPswd, sqlAddr, sqlDataBaseName)
	log.Printf("Connecting to MySQL server: dsn = %v\n", dsn)
	var err error
	SqlDb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		*connSentinel++
		log.Fatalf("Error occoured when connecting to the data base! error: %e", err)
	}
}
