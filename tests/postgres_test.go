package tests_test

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func TestPostgres(t *testing.T) {
	var cl = func() {}
	var err error
	var db *gorm.DB
	if os.Getenv("GORM_DIALECT") == "immudb"{
		db, cl, err = SetUp()
		defer cl()
		if err != nil {
			t.Error(err)
		}
	}
	if db.Dialector.Name() != "postgres" {
		t.Skip()
	}

	type Harumph struct {
		gorm.Model
		Name   string         `gorm:"check:name_checker,name <> ''"`
		Test   uuid.UUID      `gorm:"type:uuid;not null;default:gen_random_uuid()"`
		Things pq.StringArray `gorm:"type:text[]"`
	}

	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto;").Error; err != nil {
		t.Errorf("Failed to create extension pgcrypto, got error %v", err)
	}

	db.Migrator().DropTable(&Harumph{})

	if err := db.AutoMigrate(&Harumph{}); err != nil {
		t.Fatalf("Failed to migrate for uuid default value, got error: %v", err)
	}

	harumph := Harumph{}
	if err := db.Create(&harumph).Error; err == nil {
		t.Fatalf("should failed to create data, name can't be blank")
	}

	harumph = Harumph{Name: "jinzhu"}
	if err := db.Create(&harumph).Error; err != nil {
		t.Fatalf("should be able to create data, but got %v", err)
	}

	var result Harumph
	if err := db.First(&result, "id = ?", harumph.ID).Error; err != nil || harumph.Name != "jinzhu" {
		t.Errorf("No error should happen, but got %v", err)
	}

	if err := db.Where("id = $1", harumph.ID).First(&Harumph{}).Error; err != nil || harumph.Name != "jinzhu" {
		t.Errorf("No error should happen, but got %v", err)
	}
}

type Post struct {
	ID         uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4();autoincrement"`
	Title      string
	Categories []*Category `gorm:"Many2Many:post_categories"`
}

type Category struct {
	ID    uuid.UUID `gorm:"primary_key;type:uuid;default:uuid_generate_v4();autoincrement"`
	Title string
	Posts []*Post `gorm:"Many2Many:post_categories"`
}

func TestMany2ManyWithDefaultValueUUID(t *testing.T) {
	var cl = func() {}
	var err error
	var db *gorm.DB
	if os.Getenv("GORM_DIALECT") == "immudb"{
		db, cl, err = SetUp()
		defer cl()
		if err != nil {
			t.Error(err)
		}
	}
	if db.Dialector.Name() != "postgres" {
		t.Skip()
	}

	if err := db.Exec(`create extension if not exists "uuid-ossp"`).Error; err != nil {
		t.Fatalf("Failed to create 'uuid-ossp' extension, but got error %v", err)
	}

	db.Migrator().DropTable(&Post{}, &Category{}, "post_categories")
	db.AutoMigrate(&Post{}, &Category{})

	post := Post{
		Title: "Hello World",
		Categories: []*Category{
			{Title: "Coding"},
			{Title: "Golang"},
		},
	}

	if err := db.Create(&post).Error; err != nil {
		t.Errorf("Failed, got error: %v", err)
	}
}
