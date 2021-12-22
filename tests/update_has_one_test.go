package tests_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

func TestUpdateHasOne(t *testing.T) {
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
	var user = *GetUser("update-has-one", Config{})

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	user.Account = Account{Number: "account-has-one-association"}

	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user2 User
	db.Preload("Account").Find(&user2, "id = ?", user.ID)
	CheckUser(t, user2, user, db)

	user.Account.Number += "new"
	if err := db.Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user3 User
	db.Preload("Account").Find(&user3, "id = ?", user.ID)

	CheckUser(t, user2, user3, db)
	var lastUpdatedAt = user2.Account.UpdatedAt
	time.Sleep(time.Second)

	if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&user).Error; err != nil {
		t.Fatalf("errors happened when update: %v", err)
	}

	var user4 User
	db.Preload("Account").Find(&user4, "id = ?", user.ID)

	if lastUpdatedAt.Format(time.RFC3339) == user4.Account.UpdatedAt.Format(time.RFC3339) {
		t.Fatalf("updated at should be updated, but not, old: %v, new %v", lastUpdatedAt.Format(time.RFC3339), user3.Account.UpdatedAt.Format(time.RFC3339))
	} else {
		user.Account.UpdatedAt = user4.Account.UpdatedAt
		CheckUser(t, user4, user, db)
	}

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
		var pet = Pet{Name: "create"}

		if err := db.Create(&pet).Error; err != nil {
			t.Fatalf("errors happened when create: %v", err)
		}

		pet.Toy = Toy{Name: "Update-HasOneAssociation-Polymorphic"}

		if err := db.Save(&pet).Error; err != nil {
			t.Fatalf("errors happened when create: %v", err)
		}

		var pet2 Pet
		db.Preload("Toy").Find(&pet2, "id = ?", pet.ID)
		CheckPet(t, pet2, pet, db)

		pet.Toy.Name += "new"
		if err := db.Save(&pet).Error; err != nil {
			t.Fatalf("errors happened when update: %v", err)
		}

		var pet3 Pet
		db.Preload("Toy").Find(&pet3, "id = ?", pet.ID)
		CheckPet(t, pet2, pet3, db)

		if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&pet).Error; err != nil {
			t.Fatalf("errors happened when update: %v", err)
		}

		var pet4 Pet
		db.Preload("Toy").Find(&pet4, "id = ?", pet.ID)
		CheckPet(t, pet4, pet, db)
	})

	t.Run("Restriction", func(t *testing.T) {
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
		type CustomizeAccount struct {
			gorm.Model
			UserID sql.NullInt64
			Number string `gorm:"<-:create"`
		}

		type CustomizeUser struct {
			gorm.Model
			Name    string
			Account CustomizeAccount `gorm:"foreignkey:UserID"`
		}

		db.Migrator().DropTable(&CustomizeUser{})
		db.Migrator().DropTable(&CustomizeAccount{})

		if err := db.AutoMigrate(&CustomizeUser{}); err != nil {
			t.Fatalf("failed to migrate, got error: %v", err)
		}
		if err := db.AutoMigrate(&CustomizeAccount{}); err != nil {
			t.Fatalf("failed to migrate, got error: %v", err)
		}

		number := "number-has-one-associations"
		cusUser := CustomizeUser{
			Name: "update-has-one-associations",
			Account: CustomizeAccount{
				Number: number,
			},
		}

		if err := db.Create(&cusUser).Error; err != nil {
			t.Fatalf("errors happened when create: %v", err)
		}
		cusUser.Account.Number += "-update"
		if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&cusUser).Error; err != nil {
			t.Fatalf("errors happened when create: %v", err)
		}

		var account2 CustomizeAccount
		db.Find(&account2, "user_id = ?", cusUser.ID)
		AssertEqual(t, account2.Number, number)
	})
}
