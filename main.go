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
)

type Client struct {
	ID       int    `json:"id" gorm:"column:id; type:int; primaryKey; autoIncrement"`
	FullName string `json:"fullName" gorm:"column:full_name; type:text; not null"`
}

func (Client) TableName() string {
	return "clients"
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

	server := tinyhttp.NewServer(address)

	server.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, query(db.DB))
	})

	server.POST("/", func(c *gin.Context) {
		fullName := c.Query("fullName")
		if fullName == "" {
			fullName = "John Doe"
		}

		c.JSON(http.StatusOK, insert(db.DB, fullName))
	})

	tiny.StartAndBlock(server)
}

func migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&Client{},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to migrate schema")
		os.Exit(1)
	}
}

func insert(db *gorm.DB, fullName string) *Client {
	clientToInsert := &Client{
		FullName: fullName,
	}
	if tx := db.Create(clientToInsert); tx.Error != nil {
		log.Error().Err(tx.Error).Msg("failed to insert record")
		return nil
	}

	return clientToInsert
}

func query(db *gorm.DB) []*Client {
	var clients []*Client
	if tx := db.Find(&clients); tx.Error != nil {
		log.Error().Err(tx.Error).Msg("failed to query records")
	}

	return clients
}
