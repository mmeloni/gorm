package tests_test

import (
	"os"
	"testing"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

func TestUpdateMany2ManyAssociations(t *testing.T) {
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
	var user = *GetUser("update-many2many", Config{})

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	user.Languages = []Language{{Code: "zh-CN", Name: "Chinese"}, {Code: "en", Name: "English"}}
	for _, lang := range user.Languages {
		db.Create(&lang)
	}
	user.Friends = []*User{{Name: "friend-1"}, {Name: "friend-2"}}

	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user2 User
	db.Preload("Languages").Preload("Friends").Find(&user2, "id = ?", user.ID)
	CheckUser(t, user2, user, db)

	for idx := range user.Friends {
		user.Friends[idx].Name += "new"
	}

	for idx := range user.Languages {
		user.Languages[idx].Name += "new"
	}

	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user3 User
	db.Preload("Languages").Preload("Friends").Find(&user3, "id = ?", user.ID)
	CheckUser(t, user2, user3, db)

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user4 User
	db.Preload("Languages").Preload("Friends").Find(&user4, "id = ?", user.ID)
	CheckUser(t, user4, user, db)
}
