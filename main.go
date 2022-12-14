package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/mkorman9/tiny"
	"github.com/mkorman9/tiny/tinyhttp"
	"github.com/mkorman9/tiny/tinysqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
)

type Client struct {
	ID          int          `json:"id" gorm:"column:id; type:int; primaryKey; autoIncrement"`
	FullName    string       `json:"fullName" gorm:"column:full_name; type:text; not null"`
	Age         int          `json:"age" gorm:"column:age; type: int; not null"`
	Credentials *Credentials `json:"credentials,omitempty" gorm:"foreignKey:ClientID; constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}

func (Client) TableName() string {
	return "clients"
}

type Credentials struct {
	ID       int    `json:"-" gorm:"column:id; type:int; primaryKey; autoIncrement"`
	ClientID int    `json:"-" gorm:"column:client_id; type:int; unique"`
	Email    string `json:"email" gorm:"column:email; type:text; not null"`
	Password string `json:"password" gorm:"column:password; type:text; not null"`
}

func (Credentials) TableName() string {
	return "client_credentials"
}

type BasicCredentials struct {
	Email    string
	Password string
}

func main() {
	var dsn string
	var address string

	flag.StringVar(&dsn, "dsn", ":memory:", "database DSN")
	flag.StringVar(&address, "address", "0.0.0.0:8080", "HTTP server address")
	flag.Parse()

	_ = tiny.LoadConfig()
	tiny.SetupLogger()

	db, err := tinysqlite.Open(dsn)
	if err != nil {
		log.Error().Err(err).Msg("failed to open sqlite")
		os.Exit(1)
	}
	defer func() {
		_ = db.Close()
	}()

	migrate(db.DB)
	insertTestData(db.DB)

	server := tinyhttp.NewServer(address)

	server.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, query(db.DB))
	})

	server.GET("/age/avg", func(c *gin.Context) {
		c.JSON(http.StatusOK, queryAverageAge(db.DB))
	})

	server.POST("/", func(c *gin.Context) {
		fullName := c.Query("fullName")
		if fullName == "" {
			fullName = "John Doe"
		}

		age, err := strconv.Atoi(c.Query("age"))
		if err != nil || age < 18 {
			age = 21
		}

		c.JSON(http.StatusOK, insertClient(db.DB, fullName, age, nil))
	})

	tiny.StartAndBlock(server)
}

func migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&Client{},
		&Credentials{},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to migrate schema")
		os.Exit(1)
	}
}

func insertTestData(db *gorm.DB) {
	insertClient(db, "John Doe", 31, &BasicCredentials{Email: "john.doe@example.com", Password: "12345"})
	insertClient(db, "Amy Kruger", 25, nil)
	insertClient(db, "Donald Trump", 60, &BasicCredentials{Email: "donald.trump@example.com", Password: "china"})
}

func insertClient(db *gorm.DB, fullName string, age int, credentials *BasicCredentials) *Client {
	clientToInsert := &Client{
		FullName: fullName,
		Age:      age,
	}
	if tx := db.Create(clientToInsert); tx.Error != nil {
		log.Error().Err(tx.Error).Msg("failed to insert client")
		return nil
	}

	if credentials != nil {
		credentialsToInsert := &Credentials{
			ClientID: clientToInsert.ID,
			Email:    credentials.Email,
			Password: credentials.Password,
		}
		if tx := db.Create(credentialsToInsert); tx.Error != nil {
			log.Error().Err(tx.Error).Msg("failed to insert credentials")
			return nil
		}

		clientToInsert.Credentials = credentialsToInsert
	}

	return clientToInsert
}

func query(db *gorm.DB) []*Client {
	var clients []*Client
	if tx := db.Joins("Credentials").Find(&clients); tx.Error != nil {
		log.Error().Err(tx.Error).Msg("failed to query records")
	}

	return clients
}

func queryAverageAge(db *gorm.DB) float64 {
	var avgAge float64
	if tx := db.Raw("SELECT AVG(age) FROM clients").Scan(&avgAge); tx.Error != nil {
		log.Error().Err(tx.Error).Msg("failed to query average age")
		return 0
	}

	return avgAge
}
