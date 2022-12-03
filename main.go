package main

import (
	"flag"
	"github.com/mkorman9/tiny"
	"github.com/mkorman9/tiny/tinysqlite"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"os"
)

type Client struct {
	ID       int    `gorm:"column:id; type:int; primaryKey; autoIncrement"`
	FullName string `gorm:"column:full_name; type:text; not null"`
}

func (Client) TableName() string {
	return "clients"
}

func main() {
	var dsn string
	flag.StringVar(&dsn, "dsn", ":memory:", "database DSN")
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
	insert(db.DB)
	query(db.DB)
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

func insert(db *gorm.DB) {
	clientToInsert := &Client{
		FullName: "John Doe",
	}
	if tx := db.Create(clientToInsert); tx.Error != nil {
		log.Error().Err(tx.Error).Msg("failed to insert record")
		os.Exit(1)
	}

	log.Info().Msgf("inserted: id = %d, full_name = %s", clientToInsert.ID, clientToInsert.FullName)
}

func query(db *gorm.DB) {
	var clients []*Client
	if tx := db.Find(&clients); tx.Error != nil {
		log.Error().Err(tx.Error).Msg("failed to query records")
		os.Exit(1)
	}

	for _, c := range clients {
		log.Info().Msgf("returned: id = %d, full_name = %s", c.ID, c.FullName)
	}
}
