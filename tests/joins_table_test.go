package tests_test

import (
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Person struct {
	ID        int
	Name      string
	Addresses []Address `gorm:"many2many:person_addresses;"`
	DeletedAt gorm.DeletedAt
}

type Address struct {
	ID   uint
	Name string
}

type PersonAddress struct {
	PersonID  int
	AddressID int
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func _TestOverrideJoinTable(t *testing.T) {
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
	db.Migrator().DropTable(&Person{}, &Address{}, &PersonAddress{})

	if err := db.SetupJoinTable(&Person{}, "Addresses", &PersonAddress{}); err != nil {
		t.Fatalf("Failed to setup join table for person, got error %v", err)
	}

	if err := db.AutoMigrate(&Person{}, &Address{}); err != nil {
		t.Fatalf("Failed to migrate, got %v", err)
	}

	address1 := Address{Name: "address 1"}
	address2 := Address{Name: "address 2"}
	person := Person{Name: "person", Addresses: []Address{address1, address2}}
	db.Create(&person)

	var addresses1 []Address
	if err := db.Model(&person).Association("Addresses").Find(&addresses1); err != nil || len(addresses1) != 2 {
		t.Fatalf("Failed to find address, got error %v, length: %v", err, len(addresses1))
	}

	if err := db.Model(&person).Association("Addresses").Delete(&person.Addresses[0]); err != nil {
		t.Fatalf("Failed to delete address, got error %v", err)
	}

	if len(person.Addresses) != 1 {
		t.Fatalf("Should have one address left")
	}

	if db.Find(&[]PersonAddress{}, "person_id = ?", person.ID).RowsAffected != 1 {
		t.Fatalf("Should found one address")
	}

	var addresses2 []Address
	if err := db.Model(&person).Association("Addresses").Find(&addresses2); err != nil || len(addresses2) != 1 {
		t.Fatalf("Failed to find address, got error %v, length: %v", err, len(addresses2))
	}

	if db.Model(&person).Association("Addresses").Count() != 1 {
		t.Fatalf("Should found one address")
	}

	var addresses3 []Address
	if err := db.Unscoped().Model(&person).Association("Addresses").Find(&addresses3); err != nil || len(addresses3) != 2 {
		t.Fatalf("Failed to find address, got error %v, length: %v", err, len(addresses3))
	}

	if db.Unscoped().Find(&[]PersonAddress{}, "person_id = ?", person.ID).RowsAffected != 2 {
		t.Fatalf("Should found soft deleted addresses with unscoped")
	}

	if db.Unscoped().Model(&person).Association("Addresses").Count() != 2 {
		t.Fatalf("Should found soft deleted addresses with unscoped")
	}

	db.Model(&person).Association("Addresses").Clear()

	if db.Model(&person).Association("Addresses").Count() != 0 {
		t.Fatalf("Should deleted all addresses")
	}

	if db.Unscoped().Model(&person).Association("Addresses").Count() != 2 {
		t.Fatalf("Should found soft deleted addresses with unscoped")
	}

	db.Unscoped().Model(&person).Association("Addresses").Clear()

	if db.Unscoped().Model(&person).Association("Addresses").Count() != 0 {
		t.Fatalf("address should be deleted when clear with unscoped")
	}

	address2_1 := Address{Name: "address 2-1"}
	address2_2 := Address{Name: "address 2-2"}
	person2 := Person{Name: "person_2", Addresses: []Address{address2_1, address2_2}}
	db.Create(&person2)
	if err := db.Select(clause.Associations).Delete(&person2).Error; err != nil {
		t.Fatalf("failed to delete person, got error: %v", err)
	}

	if count := db.Unscoped().Model(&person2).Association("Addresses").Count(); count != 2 {
		t.Errorf("person's addresses expects 2, got %v", count)
	}

	if count := db.Model(&person2).Association("Addresses").Count(); count != 0 {
		t.Errorf("person's addresses expects 2, got %v", count)
	}
}
