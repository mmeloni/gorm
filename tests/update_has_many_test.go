package tests_test

import (
	"os"
	"testing"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

func TestUpdateHasManyAssociations(t *testing.T) {
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
	var user = *GetUser("update-has-many", Config{})

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	user.Pets = []*Pet{{Name: "pet1"}, {Name: "pet2"}}
	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user2 User
	db.Preload("Pets").Find(&user2, "id = ?", user.ID)
	CheckUser(t, user2, user, db)

	for _, pet := range user.Pets {
		pet.Name += "new"
	}

	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user3 User
	db.Preload("Pets").Find(&user3, "id = ?", user.ID)
	CheckUser(t, user2, user3, db)

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user4 User
	db.Preload("Pets").Find(&user4, "id = ?", user.ID)
	CheckUser(t, user4, user, db)

	t.Run("Polymorphic", func(t *testing.T) {
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
		var user = *GetUser("update-has-many", Config{})

		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("errors happened when create: %v", err)
		}

		user.Toys = []Toy{{Name: "toy1"}, {Name: "toy2"}}
		if err := db.Save(&user).Error; err != nil {
			t.Fatalf("errors happened when update: %v", err)
		}

		var user2 User
		db.Preload("Toys").Find(&user2, "id = ?", user.ID)
		CheckUser(t, user2, user, db)

		for idx := range user.Toys {
			user.Toys[idx].Name += "new"
		}

		if err := db.Save(&user).Error; err != nil {
			t.Fatalf("errors happened when update: %v", err)
		}

		var user3 User
		db.Preload("Toys").Find(&user3, "id = ?", user.ID)
		CheckUser(t, user2, user3, db)

		if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&user).Error; err != nil {
			t.Fatalf("errors happened when update: %v", err)
		}

		var user4 User
		db.Preload("Toys").Find(&user4, "id = ?", user.ID)
		CheckUser(t, user4, user, db)
	})
}
