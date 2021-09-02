//package main

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-redis/redis"
	"github.com/gomarkho/sas-rvm-provapi/auth"
	"github.com/gomarkho/sas-rvm-provapi/data/service"
	"github.com/gomarkho/sas-rvm-provapi/model"
	apirouterv1 "github.com/gomarkho/sas-rvm-provapi/router/v1"
	"github.com/gomarkho/sas-rvm-provapi/utils/cmlutils"
	"github.com/gomarkho/sas-rvm-provapi/utils/logger"
	"github.com/jasonlvhit/gocron"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	// check if all env variables are valid
	hasValidEnvVariables()

	// initialize logger
	logger.InitLogger()

	// general configurations
	configureGeneral()

	// configure new db session
	dbSession := getSession()
	defer dbSession.Close()
	redisSession := getRedisSession()
	defer redisSession.Close()

	redisSessiontwo := getRedisSessiontwo()
	defer redisSession.Close()

	index := mgo.Index{
		Key:    []string{"telnyxID", "contactListID", "userid"},
		Unique: true,
	}
	if err := dbSession.DB(cmlutils.DefaultDatabase()).C("telnyxnumbersgroup").EnsureIndex(index); err != nil {
		logrus.Errorln(err.Error())
	}
	go func() {
		cr := cron.New()
		_, err := cr.AddFunc("@daily", func() {
			Job(dbSession)
		})
		if err != nil {
			panic(err)
		}
		cr.Run()
		defer cr.Stop()
	}()




	go func() {
		cr1 := cron.New()
		_, err1 := cr1.AddFunc("@daily", func() {
			Job1(dbSession)
		})
		if err1 != nil {
			logrus.Errorln(err1.Error())
		}
	cr1.Run()
		defer cr1.Stop()
	}()
	// configure to serve WebServices
	go func() {

		gocron.Every(1).Day().At("00:30").Do(CheckBalance, dbSession)
		<-gocron.Start()

	}()
	configureAsHttpServer(dbSession, redisSession, redisSessiontwo)

}

func configureGeneral() {
	rand.Seed(time.Now().UTC().UnixNano())
	logrus.Infoln(cmlutils.RandomString(10))
}

func CheckBalance(session *mgo.Session) {
	logrus.Infoln("Monthly Cron trigger")
	current_time := time.Now().UTC()
	//fmt.Println(current_time)
	//after := t.AddDate(0, -2, 0)
	//fromDate := after.UTC()
	var results []model.TelnyxNumber
	err2 := session.DB("coredb").C("numbers").Find(bson.M{}).All((&results))
	if err2 != nil {
		logrus.Infoln("Error :", err2.Error())

	}
	var user model.User
	var parentUser model.User
	for i, _ := range results {
		days := results[i].UpdateDate.Sub(current_time).Hours() / 24
		var diff int = int(days)

		if diff == -30 {
			err := session.DB("coredb").C("users").Find(bson.M{"_id": results[i].UserID}).One((&user))
			if err != nil {
				logrus.Infoln("Error :", err.Error())
			}
			err1 := session.DB("coredb").C("users").Find(bson.M{"_id": user.ParentID}).One((&parentUser))
			if err1 != nil {

			}

			monthlyrate := parentUser.Defaults.MonthlyPurchaseRate
			f := float32(monthlyrate)

			err3 := ProcessDeductBalance(results[i].UserID.Hex(), f)
			if err3 != nil {
				logrus.Infoln("Error :", err3.Error())
			}
			err4 := session.DB("coredb").C("numbers").Update(bson.M{"userID": results[i].UserID}, bson.M{"$set": bson.M{"updateDate": current_time}})
			if err4 != nil {

				logrus.Infoln("Error :", err4.Error())
			}
		}
		//fmt.Println(v)
	}
}

func ProcessDeductBalance(userId string, amount float32) error {

	var userIdHex bson.ObjectId = bson.ObjectIdHex(userId)

	balanceObj := model.DeductBalanceMessage{Amount: amount, UserID: userIdHex}

	BalanceDeductQueue := os.Getenv("BALANCE_DEDUCT_QUEUE")

	b, err := json.Marshal(balanceObj)
	if err != nil {
		return err
	}
	//now := time.Now().Add(1 * time.Minute).UTC()
	//s := rand.NewSource(now.UnixNano())
	//rnd := rand.New(s)
	//rn := rnd.Int63n(8) + 12
	bm := service.NewBillingManager()
	params := &sqs.SendMessageInput{
		MessageBody:            aws.String(string(b)),
		QueueUrl:               aws.String(BalanceDeductQueue),
		MessageDeduplicationId: aws.String(bson.NewObjectId().Hex()),
		MessageGroupId:         aws.String(userId),
		//DelaySeconds: aws.Int64(rn),
	}
	if _, err := bm.SVC.SendMessage(params); err != nil {
		logrus.Errorln("Error : ", err.Error())
		return err
	}

	return nil
}

func Job(session *mgo.Session) {

	logrus.Infoln("cron trigger")
	t := time.Now().UTC()
	after := t.AddDate(0, 0, -7)
	fromDate := after.UTC()

	_, err2 := session.DB("coredb").C("cdr").RemoveAll(bson.M{"createDate": bson.M{"$lt": fromDate}})
	if err2 != nil {
		logrus.Infoln("Error :", err2.Error())

	}



}

func Job1(session *mgo.Session) {

	logrus.Infoln("cron trigger")
	t := time.Now().UTC()
	after := t.AddDate(0, 0, -7)
	fromDate := after.UTC()


	_, err3 := session.DB("coredb").C("cdr_raw").RemoveAll(bson.M{"createdDate": bson.M{"$lt": fromDate}})
	if err3 != nil {
		logrus.Infoln("Error :", err3.Error())

	}

}
func configureAsHttpServer(dbSession *mgo.Session, redisSession *redis.Client, redisSessiontwo *redis.Client) {

	e := echo.New()

	// add validation
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(middleware.Recover())
	//e.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{Skipper: skipper, Handler: bodyDumpHandler}))

	//CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{echo.GET, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		ExposeHeaders: []string{"x_auth_token"},
	}))

	// perform routing for v1 version of web apis
	performV1APIRouting(e, dbSession, redisSession, redisSessiontwo)

	if os.Getenv("ENV") == "prod" {
		// Server
		e.Logger.Fatal(e.StartTLS(":"+os.Getenv("PORT"), os.Getenv("CERT_PATH"), os.Getenv("CERT_KEY_PATH")))
	} else {
		// Server
		e.Logger.Fatal(e.Start(":" + os.Getenv("PORT")))
	}
}

func performV1APIRouting(echo *echo.Echo, dbSession *mgo.Session, redisSession *redis.Client, redisSessiontwo *redis.Client) {

	// accessible web services will fall in this group
	acc := echo.Group("/v1")

	// restricted web services will fall in this group
	res := echo.Group("/v1")

	// Configure middleware with the custom claims type
	config := middleware.JWTConfig{
		Claims:     &auth.JwtUserClaim{},
		SigningKey: []byte(auth.JwtSecret),
	}

	// add middleware for restricted services
	res.Use(middleware.JWTWithConfig(config))
	res.Use(auth.MiddlewareRes)

	// route user apis
	apirouterv1.RouteUserApis(acc, res, dbSession, redisSession, redisSessiontwo)

	// route user detail apis
	apirouterv1.RouteUserDetailApis(acc, res, dbSession)

	apirouterv1.RouteSystemMessagesApis(acc, res, dbSession)
	// route root apis
	apirouterv1.RouteRootApis(acc, res, dbSession)

	// route sound file apis
	apirouterv1.RouteSoundFileInfoApis(acc, res, dbSession)

	// route contact list apis
	apirouterv1.RouteContactListApis(acc, res, dbSession, redisSession)

	// route campaign apis
	apirouterv1.RouteCampaignApis(acc, res, dbSession, redisSession, redisSessiontwo)

	// route campaign history apis
	apirouterv1.RouteCampaignHistoryApi(acc, res)

	// route cdr apis
	apirouterv1.RouteCdrApis(acc, res, dbSession)

	//route telnyx apis
	apirouterv1.RouteNumberApis(acc, res, dbSession, redisSession)

	apirouterv1.RouteNumbersCarrierGroupApis(acc, res, dbSession, redisSession)

	//route telnyx cid apis
	//apirouterv1.RouteTelnyxCidApis(acc, res, dbSession)
	// route webhooks
	apirouterv1.RouteWebhookApis(acc, res, dbSession)

	// route billing history
	apirouterv1.RouteBillingApis(acc, res, dbSession)

	//route Dnc
	apirouterv1.RouteDncApis(acc, res, dbSession)

	apirouterv1.RouteSystemTTSApis(acc, res, dbSession, redisSession)

	apirouterv1.RouteCustomFieldApis(acc, res, dbSession)
}

func configCron(session *mgo.Session, session2 *redis.Client) {

}

// getSession creates a new mongo session and panics if connection error occurs
func getSession() *mgo.Session {
	info := &mgo.DialInfo{
		Addrs:    []string{os.Getenv("DB_HOST_1")},
		Timeout:  360 * time.Second,
		Database: cmlutils.DefaultDatabase(),
		Username: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Source:   os.Getenv("DB_AUTH_DB"),
	}

	session, err := mgo.DialWithInfo(info)
	logrus.Infoln(info)
	if err != nil {
		panic(err)
	} else {
		logrus.Infoln("----------------------------------- \n")
		logrus.Infoln("*** Database Session created ***\n")
		logrus.Infoln("----------------------------------- \n")
	}

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	return session
}

//redis connection details
func getRedisSession() *redis.Client {
	//getting redis connection details
	reHost1 := os.Getenv("REDIS_HOST_NUM_INFO")
	rePort := os.Getenv("REDIS_PORT_NUM_INFO")
	db := os.Getenv("REDIS_DB_NUM_INFO")

	dbi, err := strconv.Atoi(db)
	if err != nil {
		panic("Invalid REDIS_DB_NUM_INFO value")
	}
	reAddr1 := reHost1 + ":" + rePort

	//connecting to redis(cache database)
	logrus.Debugln("Redis connection establishing .....")

	var client *redis.Client

	if os.Getenv("ENV") == "local" {
		dbi = 15
	}

	client = redis.NewClient(&redis.Options{
		Addr:     reAddr1,
		Password: "CFF52C3F9E3C5A422325AD44DBAD8CA8FCE8FE387423FD726336DCBCD2CCF0A0", // no password set
		DB:       dbi,                                                                // use default DB
	})

	//checking redis health
	pong, err := client.Ping().Result()
	if err != nil {
		logrus.Errorln("Redis connection error : " + err.Error())
		logrus.Debugln("Redis initialization failed")
		logrus.Debugln("Panic generated")
		panic(err)
	}
	logrus.Infoln("Redis connection established")
	logrus.Debugln(pong)
	return client
}

//redis connection details
func getRedisSessiontwo() *redis.Client {
	//getting redis connection details
	reHost1 := os.Getenv("REDIS_HOST_CORE")
	rePort := os.Getenv("REDIS_PORT_CORE")
	db := os.Getenv("REDIS_DB_CORE")

	dbi, err := strconv.Atoi(db)
	if err != nil {
		panic("Invalid REDIS_DB_CORE value")
	}
	reAddr1 := reHost1 + ":" + rePort

	//connecting to redis(cache database)
	logrus.Infoln("Redis connection establishing .....")

	var client *redis.Client

	if os.Getenv("ENV") == "local" {
		dbi = 0
	}

	client = redis.NewClient(&redis.Options{
		Addr:     reAddr1,
		Password: "CFF52C3F9E3C5A422325AD44DBAD8CA8FCE8FE387423FD726336DCBCD2CCF0A0", // no password set
		DB:       dbi,                                                                // use default DB
	})

	//checking redis health
	pong, err := client.Ping().Result()
	if err != nil {
		logrus.Errorln("Redis connection error : " + err.Error())
		logrus.Debugln("Redis initialization failed")
		logrus.Debugln("Panic generated")
		panic(err)
	}
	logrus.Infoln("Redis connection established")
	logrus.Debugln(pong)
	return client
}

// middle ware handler
func bodyDumpHandler(c echo.Context, reqBody, resBody []byte) {

	logrus.Debugln("================================")

	printRequest(c.Request())

	logrus.Debugln("--------request body-------")
	printBody(reqBody)
	logrus.Debugln("---------------------------")

	logrus.Debugln("-------- response body --------")
	printBody(resBody)
	logrus.Debugln("-------------------------------")
	logrus.Debugln("=================================")
}

func hasValidEnvVariables() {
	if os.Getenv("PORT") == "" {
		panic(errors.New("Please provide valid PORT"))
	}

	if os.Getenv("ENV") == "" {
		panic(errors.New("Please provide valid ENV"))
	}

	if os.Getenv("CERT_KEY_PATH") == "" {
		panic(errors.New("Please provide valid CERT_KEY_PATH"))
	}

	if os.Getenv("CERT_PATH") == "" {
		panic(errors.New("Please provide valid CERT_PATH"))
	}

	if os.Getenv("VOIP_API") == "" {
		panic(errors.New("Please provide valid VOIP_API"))
	}

	if os.Getenv("CDR_API") == "" {
		panic(errors.New("Please provide valid CDR_API"))
	}

	if os.Getenv("BUCKET_NAME") == "" {
		panic(errors.New("Please provide valid BUCKET_NAME"))
	}

	if os.Getenv("BUCKET_PATH") == "" {
		panic(errors.New("Please provide valid BUCKET_PATH"))
	}

	if os.Getenv("DB_HOST_1") == "" {
		panic(errors.New("Please provide valid DB_HOST_1"))
	}

	if os.Getenv("DB_DEFAULT") == "" {
		panic(errors.New("Please provide valid DB_DEFAULT"))
	}

	if os.Getenv("DB_USERNAME") == "" {
		panic(errors.New("Please provide valid DB_USERNAME"))
	}

	if os.Getenv("DB_PASSWORD") == "" {
		panic(errors.New("Please provide valid DB_PASSWORD"))
	}

	if os.Getenv("DB_AUTH_DB") == "" {
		panic(errors.New("Please provide valid DB_AUTH_DB"))
	}

	if os.Getenv("BALANCE_ADD_QUEUE") == "" {
		panic(errors.New("Please provide valid BALANCE_ADD_QUEUE"))
	}

	if os.Getenv("BUCKET_PROXY_PATH") == "" {
		panic(errors.New("Please provide valid BUCKET_PROXY_PATH"))
	}
	if os.Getenv("REDIS_HOST_NUM_INFO") == "" {
		panic(errors.New("Please provide valid REDIS_HOST_NUM_INFO"))
	}
	if os.Getenv("BULK_VS_USERNAME") == "" {
		panic(errors.New("Please provide valid BULK_VS_USERNAME"))
	}

	if os.Getenv("BULK_VS_PASSWORD") == "" {
		panic(errors.New("Please provide valid BULK_VS_PASSWORD"))
	}

	if os.Getenv("REDIS_PORT_NUM_INFO") == "" {
		panic(errors.New("Please provide valid REDIS_PORT_NUM_INFO"))
	}

	if os.Getenv("REDIS_DB_NUM_INFO") == "" {
		panic(errors.New("Please provide valid REDIS_DB_NUM_INFO"))
	}
	if os.Getenv("REDIS_HOST_CORE") == "" {
		panic(errors.New("Please provide valid REDIS_HOST_CORE"))
	}

	if os.Getenv("REDIS_PORT_CORE") == "" {
		panic(errors.New("Please provide valid REDIS_PORT_CORE"))
	}

	if os.Getenv("REDIS_DB_CORE") == "" {
		panic(errors.New("Please provide valid REDIS_DB_CORE"))
	}
	if os.Getenv("LOG_LEVEL") == "" {
		panic(errors.New("Please provide valid LOG_LEVEL"))
	}

	if os.Getenv("SIGNUP_NOTIFICATION_FROM_EMAIL") == "" {
		panic(errors.New("Please provide valid SIGNUP_NOTIFICATION_FROM_EMAIL"))
	}
	if os.Getenv("SIGNUP_NOTIFICATION_TO_EMAIL") == "" {
		panic(errors.New("Please provide valid SIGNUP_NOTIFICATION_TO_EMAIL"))
	}

	if os.Getenv("TCPA_USERNAME") == "" {
		panic(errors.New("Please provide valid TCPA_USERNAME"))

	}
	if os.Getenv("BULK_VS_PASSWORD") == "" {
		panic(errors.New("Please provide valid BULK_VS_PASSWORD"))

	}
	if os.Getenv("BULK_VS_USERNAME") == "" {
		panic(errors.New("Please provide valid BULK_VS_USERNAME"))

	}



}

// body dump skipper
func skipper(c echo.Context) bool {

	if c.Request().Method == "POST" {
		logrus.Debugln(c.Request().URL.Path)
		return true
	}

	return false
}

// private method
func printBody(obj []byte) {

	logrus.WithFields(logrus.Fields{
		"json": string(obj),
	}).Info("Payload")
}

func printRequest(request *http.Request) {

	logrus.WithFields(logrus.Fields{
		"HOST":       request.Host,
		"PATH":       request.URL.Path,
		"METHOD":     request.Method,
		"QueryParam": request.URL.Query().Encode(),
	}).Info("Request")
}
