package tests_test

import (
	"os"
	"testing"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

func TestUpdateBelongsTo(t *testing.T) {
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
	var user = *GetUser("update-belongs-to", Config{})

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	user.Company = Company{Name: "company-belongs-to-association"}
	user.Manager = &User{Name: "manager-belongs-to-association"}
	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user2 User
	db.Preload("Company").Preload("Manager").Find(&user2, "id = ?", user.ID)
	CheckUser(t, user2, user, db)

	user.Company.Name += "new"
	user.Manager.Name += "new"
	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user3 User
	db.Preload("Company").Preload("Manager").Find(&user3, "id = ?", user.ID)
	CheckUser(t, user2, user3, db)

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user4 User
	db.Preload("Company").Preload("Manager").Find(&user4, "id = ?", user.ID)
	CheckUser(t, user4, user, db)
}
