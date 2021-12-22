package tests_test

import (
	"os"
	"testing"

	"gorm.io/gorm"
)

func TestDefaultValue(t *testing.T) {
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
	type Harumph struct {
		gorm.Model
		Email   string `gorm:"not null;index:,unique"`
		Name    string `gorm:"notNull;default:foo"`
		Name2   string `gorm:"size:233;not null;default:'foo'"`
		Name3   string `gorm:"size:233;notNull;default:''"`
		Age     int    `gorm:"default:18"`
		Enabled bool   `gorm:"default:true"`
	}

	db.Migrator().DropTable(&Harumph{})

	if err := db.AutoMigrate(&Harumph{}); err != nil {
		t.Fatalf("Failed to migrate with default value, got error: %v", err)
	}

	var harumph = Harumph{Email: "hello@gorm.io"}
	if err := db.Create(&harumph).Error; err != nil {
		t.Fatalf("Failed to create data with default value, got error: %v", err)
	} else if harumph.Name != "foo" || harumph.Name2 != "foo" || harumph.Name3 != "" || harumph.Age != 18 || !harumph.Enabled {
		t.Fatalf("Failed to create data with default value, got: %+v", harumph)
	}

	var result Harumph
	if err := db.First(&result, "email = ?", "hello@gorm.io").Error; err != nil {
		t.Fatalf("Failed to find created data, got error: %v", err)
	} else if result.Name != "foo" || result.Name2 != "foo" || result.Name3 != "" || result.Age != 18 || !result.Enabled {
		t.Fatalf("Failed to find created data with default data, got %+v", result)
	}
}
