package main

import (
	"flag"
	"github.com/gofiber/fiber/v2"
	"github.com/mkorman9/tiny"
	"github.com/mkorman9/tiny/tinyhttp"
	"github.com/mkorman9/tiny/tinysqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"net/http"
	"os"
)

type Client struct {
	ID          int          `json:"id" gorm:"column:id; type:int; primaryKey; autoIncrement"`
	FullName    string       `json:"fullName" gorm:"column:full_name; type:text; not null"`
	Age         int          `json:"age" gorm:"column:age; type:int; not null"`
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

type ClientAddForm struct {
	FullName string `json:"fullName" validate:"required"`
	Age      int    `json:"age" validate:"required,gte=18"`
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

	tiny.Init()

	db, err := tinysqlite.Open(dsn)
	if err != nil {
		log.Error().Err(err).Msg("failed to open sqlite")
		os.Exit(1)
	}

	migrate(db)
	insertTestData(db)

	server := tinyhttp.NewServer(address)

	server.Get("/", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).
			JSON(fiber.Map{
				"Hello": "World",
			})
	})

	server.Get("/age/avg", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).
			JSON(queryAverageAge(db))
	})

	server.Post("/", func(c *fiber.Ctx) error {
		var form ClientAddForm
		if errs := tinyhttp.BindBody(c, &form); errs != nil {
			return c.Status(http.StatusBadRequest).
				JSON(errs)
		}

		return c.Status(http.StatusOK).
			JSON(insertClient(db, form.FullName, form.Age, nil))
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

	err := db.Transaction(func(tx *gorm.DB) error {
		if result := tx.Create(clientToInsert); result.Error != nil {
			log.Error().Err(result.Error).Msg("failed to insert client")
			return result.Error
		}

		if credentials != nil {
			credentialsToInsert := &Credentials{
				ClientID: clientToInsert.ID,
				Email:    credentials.Email,
				Password: credentials.Password,
			}
			if result := tx.Create(credentialsToInsert); result.Error != nil {
				log.Error().Err(result.Error).Msg("failed to insert credentials")
				return result.Error
			}

			clientToInsert.Credentials = credentialsToInsert
		}

		return nil
	})
	if err != nil {
		return nil
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
	// if tx := db.Raw("SELECT AVG(age) FROM clients").Scan(&avgAge); tx.Error != nil {
	if tx := db.Select("AVG(age)").Table("clients").Scan(&avgAge); tx.Error != nil {
		log.Error().Err(tx.Error).Msg("failed to query average age")
		return 0
	}

	return avgAge
}
