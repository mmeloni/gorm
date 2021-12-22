package tests_test

import (
	"gorm.io/gorm"
	"os"
	"testing"

	. "gorm.io/gorm/utils/tests"
)

type Hamster struct {
	Id           int
	Name         string
	PreferredToy Toy `gorm:"polymorphic:Owner;polymorphicValue:hamster_preferred"`
	OtherToy     Toy `gorm:"polymorphic:Owner;polymorphicValue:hamster_other"`
}

func _TestNamedPolymorphic(t *testing.T) {
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
	db.Migrator().DropTable(&Hamster{})
	db.AutoMigrate(&Hamster{})

	hamster := Hamster{Name: "Mr. Hammond", PreferredToy: Toy{Name: "bike"}, OtherToy: Toy{Name: "treadmill"}}
	db.Save(&hamster)

	hamster2 := Hamster{}
	db.Preload("PreferredToy").Preload("OtherToy").Find(&hamster2, hamster.Id)

	if hamster2.PreferredToy.ID != hamster.PreferredToy.ID || hamster2.PreferredToy.Name != hamster.PreferredToy.Name {
		t.Errorf("Hamster's preferred toy failed to preload")
	}

	if hamster2.OtherToy.ID != hamster.OtherToy.ID || hamster2.OtherToy.Name != hamster.OtherToy.Name {
		t.Errorf("Hamster's other toy failed to preload")
	}

	// clear to omit Toy.ID in count
	hamster2.PreferredToy = Toy{}
	hamster2.OtherToy = Toy{}

	if db.Model(&hamster2).Association("PreferredToy").Count() != 1 {
		t.Errorf("Hamster's preferred toy count should be 1")
	}

	if db.Model(&hamster2).Association("OtherToy").Count() != 1 {
		t.Errorf("Hamster's other toy count should be 1")
	}

	// Query
	hamsterToy := Toy{}
	db.Model(&hamster).Association("PreferredToy").Find(&hamsterToy)
	if hamsterToy.Name != hamster.PreferredToy.Name {
		t.Errorf("Should find has one polymorphic association")
	}

	hamsterToy = Toy{}
	db.Model(&hamster).Association("OtherToy").Find(&hamsterToy)
	if hamsterToy.Name != hamster.OtherToy.Name {
		t.Errorf("Should find has one polymorphic association")
	}

	// Append
	db.Model(&hamster).Association("PreferredToy").Append(&Toy{
		Name: "bike 2",
	})

	db.Model(&hamster).Association("OtherToy").Append(&Toy{
		Name: "treadmill 2",
	})

	hamsterToy = Toy{}
	db.Model(&hamster).Association("PreferredToy").Find(&hamsterToy)
	if hamsterToy.Name != "bike 2" {
		t.Errorf("Should update has one polymorphic association with Append")
	}

	hamsterToy = Toy{}
	db.Model(&hamster).Association("OtherToy").Find(&hamsterToy)
	if hamsterToy.Name != "treadmill 2" {
		t.Errorf("Should update has one polymorphic association with Append")
	}

	if db.Model(&hamster2).Association("PreferredToy").Count() != 1 {
		t.Errorf("Hamster's toys count should be 1 after Append")
	}

	if db.Model(&hamster2).Association("OtherToy").Count() != 1 {
		t.Errorf("Hamster's toys count should be 1 after Append")
	}

	// Replace
	db.Model(&hamster).Association("PreferredToy").Replace(&Toy{
		Name: "bike 3",
	})

	db.Model(&hamster).Association("OtherToy").Replace(&Toy{
		Name: "treadmill 3",
	})

	hamsterToy = Toy{}
	db.Model(&hamster).Association("PreferredToy").Find(&hamsterToy)
	if hamsterToy.Name != "bike 3" {
		t.Errorf("Should update has one polymorphic association with Replace")
	}

	hamsterToy = Toy{}
	db.Model(&hamster).Association("OtherToy").Find(&hamsterToy)
	if hamsterToy.Name != "treadmill 3" {
		t.Errorf("Should update has one polymorphic association with Replace")
	}

	if db.Model(&hamster2).Association("PreferredToy").Count() != 1 {
		t.Errorf("hamster's toys count should be 1 after Replace")
	}

	if db.Model(&hamster2).Association("OtherToy").Count() != 1 {
		t.Errorf("hamster's toys count should be 1 after Replace")
	}

	// Clear
	db.Model(&hamster).Association("PreferredToy").Append(&Toy{
		Name: "bike 2",
	})
	db.Model(&hamster).Association("OtherToy").Append(&Toy{
		Name: "treadmill 2",
	})

	if db.Model(&hamster).Association("PreferredToy").Count() != 1 {
		t.Errorf("Hamster's toys should be added with Append")
	}

	if db.Model(&hamster).Association("OtherToy").Count() != 1 {
		t.Errorf("Hamster's toys should be added with Append")
	}

	db.Model(&hamster).Association("PreferredToy").Clear()

	if db.Model(&hamster2).Association("PreferredToy").Count() != 0 {
		t.Errorf("Hamster's preferred toy should be cleared with Clear")
	}

	if db.Model(&hamster2).Association("OtherToy").Count() != 1 {
		t.Errorf("Hamster's other toy should be still available")
	}

	db.Model(&hamster).Association("OtherToy").Clear()
	if db.Model(&hamster).Association("OtherToy").Count() != 0 {
		t.Errorf("Hamster's other toy should be cleared with Clear")
	}
}
