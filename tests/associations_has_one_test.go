package tests_test

import (
	"gorm.io/gorm"
	"os"
	"testing"

	. "gorm.io/gorm/utils/tests"
)

func TestHasOneAssociation(t *testing.T) {
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
	var user = *GetUser("hasone", Config{Account: true})

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	CheckUser(t, user, user, db)

	// Find
	var user2 User
	db.Find(&user2, "id = ?", user.ID)
	db.Model(&user2).Association("Account").Find(&user2.Account)
	CheckUser(t, user2, user, db)

	// Count
	AssertAssociationCount(t, user, "Account", 1, "", db)

	// Append
	var account = Account{Number: "account-has-one-append"}

	if err := db.Model(&user2).Association("Account").Append(&account); err != nil {
		t.Fatalf("Error happened when append account, got %v", err)
	}

	if account.ID == 0 {
		t.Fatalf("Account's ID should be created")
	}

	user.Account = account
	CheckUser(t, user2, user, db)

	AssertAssociationCount(t, user, "Account", 1, "AfterAppend", db)

	// Replace
	var account2 = Account{Number: "account-has-one-replace"}

	if err := db.Model(&user2).Association("Account").Replace(&account2); err != nil {
		t.Fatalf("Error happened when append Account, got %v", err)
	}

	if account2.ID == 0 {
		t.Fatalf("account2's ID should be created")
	}

	user.Account = account2
	CheckUser(t, user2, user, db)

	AssertAssociationCount(t, user2, "Account", 1, "AfterReplace", db)

	// Delete
	if err := db.Model(&user2).Association("Account").Delete(&Account{}); err != nil {
		t.Fatalf("Error happened when delete account, got %v", err)
	}
	AssertAssociationCount(t, user2, "Account", 1, "after delete non-existing data", db)

	if err := db.Model(&user2).Association("Account").Delete(&account2); err != nil {
		t.Fatalf("Error happened when delete Account, got %v", err)
	}
	AssertAssociationCount(t, user2, "Account", 0, "after delete", db)

	// Prepare Data for Clear
	account = Account{Number: "account-has-one-append"}
	if err := db.Model(&user2).Association("Account").Append(&account); err != nil {
		t.Fatalf("Error happened when append Account, got %v", err)
	}

	AssertAssociationCount(t, user2, "Account", 1, "after prepare data", db)

	// Clear
	if err := db.Model(&user2).Association("Account").Clear(); err != nil {
		t.Errorf("Error happened when clear Account, got %v", err)
	}

	AssertAssociationCount(t, user2, "Account", 0, "after clear", db)
}

func TestHasOneAssociationWithSelect(t *testing.T) {
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
	var user = *GetUser("hasone", Config{Account: true})

	db.Omit("Account.Number").Create(&user)

	AssertAssociationCount(t, user, "Account", 1, "", db)

	var account Account
	db.Model(&user).Association("Account").Find(&account)
	if account.Number != "" {
		t.Errorf("account's number should not be saved")
	}
}

func TestHasOneAssociationForSlice(t *testing.T) {
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
	var users = []User{
		*GetUser("slice-hasone-1", Config{Account: true}),
		*GetUser("slice-hasone-2", Config{Account: false}),
		*GetUser("slice-hasone-3", Config{Account: true}),
	}

	db.Create(&users)

	// Count
	AssertAssociationCount(t, users, "Account", 2, "", db)

	// Find
	var accounts []Account
	if db.Model(&users).Association("Account").Find(&accounts); len(accounts) != 2 {
		t.Errorf("accounts count should be %v, but got %v", 3, len(accounts))
	}

	// Append
	db.Model(&users).Association("Account").Append(
		&Account{Number: "account-slice-append-1"},
		&Account{Number: "account-slice-append-2"},
		&Account{Number: "account-slice-append-3"},
	)

	AssertAssociationCount(t, users, "Account", 3, "After Append", db)

	// Replace -> same as append

	// Delete
	if err := db.Model(&users).Association("Account").Delete(&users[0].Account); err != nil {
		t.Errorf("no error should happened when deleting account, but got %v", err)
	}

	AssertAssociationCount(t, users, "Account", 2, "after delete", db)

	// Clear
	db.Model(&users).Association("Account").Clear()
	AssertAssociationCount(t, users, "Account", 0, "After Clear", db)
}

func TestPolymorphicHasOneAssociation(t *testing.T) {
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
	var pet = Pet{Name: "hasone", Toy: Toy{Name: "toy-has-one"}}

	if err := db.Create(&pet).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	CheckPet(t, pet, pet, db)

	// Find
	var pet2 Pet
	db.Find(&pet2, "id = ?", pet.ID)
	db.Model(&pet2).Association("Toy").Find(&pet2.Toy)
	CheckPet(t, pet2, pet, db)

	// Count
	AssertAssociationCount(t, pet, "Toy", 1, "", db)

	// Append
	var toy = Toy{Name: "toy-has-one-append"}

	if err := db.Model(&pet2).Association("Toy").Append(&toy); err != nil {
		t.Fatalf("Error happened when append toy, got %v", err)
	}

	if toy.ID == 0 {
		t.Fatalf("Toy's ID should be created")
	}

	pet.Toy = toy
	CheckPet(t, pet2, pet, db)

	AssertAssociationCount(t, pet, "Toy", 1, "AfterAppend", db)

	// Replace
	var toy2 = Toy{Name: "toy-has-one-replace"}

	if err := db.Model(&pet2).Association("Toy").Replace(&toy2); err != nil {
		t.Fatalf("Error happened when append Toy, got %v", err)
	}

	if toy2.ID == 0 {
		t.Fatalf("toy2's ID should be created")
	}

	pet.Toy = toy2
	CheckPet(t, pet2, pet, db)

	AssertAssociationCount(t, pet2, "Toy", 1, "AfterReplace", db)

	// Delete
	if err := db.Model(&pet2).Association("Toy").Delete(&Toy{}); err != nil {
		t.Fatalf("Error happened when delete toy, got %v", err)
	}
	AssertAssociationCount(t, pet2, "Toy", 1, "after delete non-existing data", db)

	if err := db.Model(&pet2).Association("Toy").Delete(&toy2); err != nil {
		t.Fatalf("Error happened when delete Toy, got %v", err)
	}
	AssertAssociationCount(t, pet2, "Toy", 0, "after delete", db)

	// Prepare Data for Clear
	toy = Toy{Name: "toy-has-one-append"}
	if err := db.Model(&pet2).Association("Toy").Append(&toy); err != nil {
		t.Fatalf("Error happened when append Toy, got %v", err)
	}

	AssertAssociationCount(t, pet2, "Toy", 1, "after prepare data", db)

	// Clear
	if err := db.Model(&pet2).Association("Toy").Clear(); err != nil {
		t.Errorf("Error happened when clear Toy, got %v", err)
	}

	AssertAssociationCount(t, pet2, "Toy", 0, "after clear", db)
}

func TestPolymorphicHasOneAssociationForSlice(t *testing.T) {
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
	var pets = []Pet{
		{Name: "hasone-1", Toy: Toy{Name: "toy-has-one"}},
		{Name: "hasone-2", Toy: Toy{}},
		{Name: "hasone-3", Toy: Toy{Name: "toy-has-one"}},
	}

	db.Create(&pets)

	// Count
	AssertAssociationCount(t, pets, "Toy", 2, "", db)

	// Find
	var toys []Toy
	if db.Model(&pets).Association("Toy").Find(&toys); len(toys) != 2 {
		t.Errorf("toys count should be %v, but got %v", 3, len(toys))
	}

	// Append
	db.Model(&pets).Association("Toy").Append(
		&Toy{Name: "toy-slice-append-1"},
		&Toy{Name: "toy-slice-append-2"},
		&Toy{Name: "toy-slice-append-3"},
	)

	AssertAssociationCount(t, pets, "Toy", 3, "After Append", db)

	// Replace -> same as append

	// Delete
	if err := db.Model(&pets).Association("Toy").Delete(&pets[0].Toy); err != nil {
		t.Errorf("no error should happened when deleting toy, but got %v", err)
	}

	AssertAssociationCount(t, pets, "Toy", 2, "after delete", db)

	// Clear
	db.Model(&pets).Association("Toy").Clear()
	AssertAssociationCount(t, pets, "Toy", 0, "After Clear", db)
}
