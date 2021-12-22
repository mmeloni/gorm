package tests_test

import (
	"gorm.io/gorm"
	"os"
	"testing"

	. "gorm.io/gorm/utils/tests"
)

func TestMany2ManyAssociation(t *testing.T) {
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
	var user = *GetUser("many2many", Config{Languages: 2})

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	CheckUser(t, user, user, db)

	// Find
	var user2 User
	db.Find(&user2, "id = ?", user.ID)
	db.Model(&user2).Association("Languages").Find(&user2.Languages)

	CheckUser(t, user2, user, db)

	// Count
	AssertAssociationCount(t, user, "Languages", 2, "", db)

	// Append
	var language = Language{Code: "language-many2many-append", Name: "language-many2many-append"}
	db.Create(&language)

	if err := db.Model(&user2).Association("Languages").Append(&language); err != nil {
		t.Fatalf("Error happened when append account, got %v", err)
	}

	user.Languages = append(user.Languages, language)
	CheckUser(t, user2, user, db)

	AssertAssociationCount(t, user, "Languages", 3, "AfterAppend", db)

	var languages = []Language{
		{Code: "language-many2many-append-1-1", Name: "language-many2many-append-1-1"},
		{Code: "language-many2many-append-2-1", Name: "language-many2many-append-2-1"},
	}
	db.Create(&languages)

	if err := db.Model(&user2).Association("Languages").Append(&languages); err != nil {
		t.Fatalf("Error happened when append language, got %v", err)
	}

	user.Languages = append(user.Languages, languages...)

	CheckUser(t, user2, user, db)

	AssertAssociationCount(t, user, "Languages", 5, "AfterAppendSlice", db)

	// Replace
	var language2 = Language{Code: "language-many2many-replace", Name: "language-many2many-replace"}
	db.Create(&language2)

	if err := db.Model(&user2).Association("Languages").Replace(&language2); err != nil {
		t.Fatalf("Error happened when append language, got %v", err)
	}

	user.Languages = []Language{language2}
	CheckUser(t, user2, user, db)

	AssertAssociationCount(t, user2, "Languages", 1, "AfterReplace", db)

	// Delete
	if err := db.Model(&user2).Association("Languages").Delete(&Language{}); err != nil {
		t.Fatalf("Error happened when delete language, got %v", err)
	}
	AssertAssociationCount(t, user2, "Languages", 1, "after delete non-existing data", db)

	if err := db.Model(&user2).Association("Languages").Delete(&language2); err != nil {
		t.Fatalf("Error happened when delete Languages, got %v", err)
	}
	AssertAssociationCount(t, user2, "Languages", 0, "after delete", db)

	// Prepare Data for Clear
	if err := db.Model(&user2).Association("Languages").Append(&language); err != nil {
		t.Fatalf("Error happened when append Languages, got %v", err)
	}

	AssertAssociationCount(t, user2, "Languages", 1, "after prepare data", db)

	// Clear
	if err := db.Model(&user2).Association("Languages").Clear(); err != nil {
		t.Errorf("Error happened when clear Languages, got %v", err)
	}

	AssertAssociationCount(t, user2, "Languages", 0, "after clear", db)
}

func TestMany2ManyOmitAssociations(t *testing.T) {
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
	var user = *GetUser("many2many_omit_associations", Config{Languages: 2})

	if err := db.Omit("Languages.*").Create(&user).Error; err == nil {
		t.Fatalf("should raise error when create users without languages reference")
	}

	if err := db.Create(&user.Languages).Error; err != nil {
		t.Fatalf("no error should happen when create languages, but got %v", err)
	}

	if err := db.Omit("Languages.*").Create(&user).Error; err != nil {
		t.Fatalf("no error should happen when create user when languages exists, but got %v", err)
	}

	// Find
	var languages []Language
	if db.Model(&user).Association("Languages").Find(&languages); len(languages) != 2 {
		t.Errorf("languages count should be %v, but got %v", 2, len(languages))
	}

	var newLang = Language{Code: "omitmany2many", Name: "omitmany2many"}
	if err := db.Model(&user).Omit("Languages.*").Association("Languages").Replace(&newLang); err == nil {
		t.Errorf("should failed to insert languages due to constraint failed, error: %v", err)
	}
}

func TestMany2ManyAssociationForSlice(t *testing.T) {
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
		*GetUser("slice-many2many-1", Config{Languages: 2}),
		*GetUser("slice-many2many-2", Config{Languages: 0}),
		*GetUser("slice-many2many-3", Config{Languages: 4}),
	}

	db.Create(&users)

	// Count
	AssertAssociationCount(t, users, "Languages", 6, "", db)

	// Find
	var languages []Language
	if db.Model(&users).Association("Languages").Find(&languages); len(languages) != 6 {
		t.Errorf("languages count should be %v, but got %v", 6, len(languages))
	}

	// Append
	var languages1 = []Language{
		{Code: "language-many2many-append-1", Name: "language-many2many-append-1"},
	}
	var languages2 = []Language{}
	var languages3 = []Language{
		{Code: "language-many2many-append-3-1", Name: "language-many2many-append-3-1"},
		{Code: "language-many2many-append-3-2", Name: "language-many2many-append-3-2"},
	}
	db.Create(&languages1)
	db.Create(&languages3)

	db.Model(&users).Association("Languages").Append(&languages1, &languages2, &languages3)

	AssertAssociationCount(t, users, "Languages", 9, "After Append", db)

	languages2_1 := []*Language{
		{Code: "language-slice-replace-1-1", Name: "language-slice-replace-1-1"},
		{Code: "language-slice-replace-1-2", Name: "language-slice-replace-1-2"},
	}
	languages2_2 := []*Language{
		{Code: "language-slice-replace-2-1", Name: "language-slice-replace-2-1"},
		{Code: "language-slice-replace-2-2", Name: "language-slice-replace-2-2"},
	}
	languages2_3 := &Language{Code: "language-slice-replace-3", Name: "language-slice-replace-3"}
	db.Create(&languages2_1)
	db.Create(&languages2_2)
	db.Create(&languages2_3)

	// Replace
	db.Model(&users).Association("Languages").Replace(&languages2_1, &languages2_2, languages2_3)

	AssertAssociationCount(t, users, "Languages", 5, "After Replace", db)

	// Delete
	if err := db.Model(&users).Association("Languages").Delete(&users[2].Languages); err != nil {
		t.Errorf("no error should happened when deleting language, but got %v", err)
	}

	AssertAssociationCount(t, users, "Languages", 4, "after delete", db)

	if err := db.Model(&users).Association("Languages").Delete(users[0].Languages[0], users[1].Languages[1]); err != nil {
		t.Errorf("no error should happened when deleting language, but got %v", err)
	}

	AssertAssociationCount(t, users, "Languages", 2, "after delete", db)

	// Clear
	db.Model(&users).Association("Languages").Clear()
	AssertAssociationCount(t, users, "Languages", 0, "After Clear", db)
}

func TestSingleTableMany2ManyAssociation(t *testing.T) {
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
	var user = *GetUser("many2many", Config{Friends: 2})

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	CheckUser(t, user, user, db)

	// Find
	var user2 User
	db.Find(&user2, "id = ?", user.ID)
	db.Model(&user2).Association("Friends").Find(&user2.Friends)

	CheckUser(t, user2, user, db)

	// Count
	AssertAssociationCount(t, user, "Friends", 2, "", db)

	// Append
	var friend = *GetUser("friend", Config{})

	if err := db.Model(&user2).Association("Friends").Append(&friend); err != nil {
		t.Fatalf("Error happened when append account, got %v", err)
	}

	user.Friends = append(user.Friends, &friend)
	CheckUser(t, user2, user, db)

	AssertAssociationCount(t, user, "Friends", 3, "AfterAppend", db)

	var friends = []*User{GetUser("friend-append-1", Config{}), GetUser("friend-append-2", Config{})}

	if err := db.Model(&user2).Association("Friends").Append(&friends); err != nil {
		t.Fatalf("Error happened when append friend, got %v", err)
	}

	user.Friends = append(user.Friends, friends...)

	CheckUser(t, user2, user, db)

	AssertAssociationCount(t, user, "Friends", 5, "AfterAppendSlice", db)

	// Replace
	var friend2 = *GetUser("friend-replace-2", Config{})

	if err := db.Model(&user2).Association("Friends").Replace(&friend2); err != nil {
		t.Fatalf("Error happened when append friend, got %v", err)
	}

	user.Friends = []*User{&friend2}
	CheckUser(t, user2, user, db)

	AssertAssociationCount(t, user2, "Friends", 1, "AfterReplace", db)

	// Delete
	if err := db.Model(&user2).Association("Friends").Delete(&User{}); err != nil {
		t.Fatalf("Error happened when delete friend, got %v", err)
	}
	AssertAssociationCount(t, user2, "Friends", 1, "after delete non-existing data", db)

	if err := db.Model(&user2).Association("Friends").Delete(&friend2); err != nil {
		t.Fatalf("Error happened when delete Friends, got %v", err)
	}
	AssertAssociationCount(t, user2, "Friends", 0, "after delete", db)

	// Prepare Data for Clear
	if err := db.Model(&user2).Association("Friends").Append(&friend); err != nil {
		t.Fatalf("Error happened when append Friends, got %v", err)
	}

	AssertAssociationCount(t, user2, "Friends", 1, "after prepare data", db)

	// Clear
	if err := db.Model(&user2).Association("Friends").Clear(); err != nil {
		t.Errorf("Error happened when clear Friends, got %v", err)
	}

	AssertAssociationCount(t, user2, "Friends", 0, "after clear", db)
}

func TestSingleTableMany2ManyAssociationForSlice(t *testing.T) {
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
		*GetUser("slice-many2many-1", Config{Team: 2}),
		*GetUser("slice-many2many-2", Config{Team: 0}),
		*GetUser("slice-many2many-3", Config{Team: 4}),
	}

	db.Create(&users)

	// Count
	AssertAssociationCount(t, users, "Team", 6, "", db)

	// Find
	var teams []User
	if db.Model(&users).Association("Team").Find(&teams); len(teams) != 6 {
		t.Errorf("teams count should be %v, but got %v", 6, len(teams))
	}

	// Append
	var teams1 = []User{*GetUser("friend-append-1", Config{})}
	var teams2 = []User{}
	var teams3 = []*User{GetUser("friend-append-3-1", Config{}), GetUser("friend-append-3-2", Config{})}

	db.Model(&users).Association("Team").Append(&teams1, &teams2, &teams3)

	AssertAssociationCount(t, users, "Team", 9, "After Append", db)

	var teams2_1 = []User{*GetUser("friend-replace-1", Config{}), *GetUser("friend-replace-2", Config{})}
	var teams2_2 = []User{*GetUser("friend-replace-2-1", Config{}), *GetUser("friend-replace-2-2", Config{})}
	var teams2_3 = GetUser("friend-replace-3-1", Config{})

	// Replace
	db.Model(&users).Association("Team").Replace(&teams2_1, &teams2_2, teams2_3)

	AssertAssociationCount(t, users, "Team", 5, "After Replace", db)

	// Delete
	if err := db.Model(&users).Association("Team").Delete(&users[2].Team); err != nil {
		t.Errorf("no error should happened when deleting team, but got %v", err)
	}

	AssertAssociationCount(t, users, "Team", 4, "after delete", db)

	if err := db.Model(&users).Association("Team").Delete(users[0].Team[0], users[1].Team[1]); err != nil {
		t.Errorf("no error should happened when deleting team, but got %v", err)
	}

	AssertAssociationCount(t, users, "Team", 2, "after delete", db)

	// Clear
	db.Model(&users).Association("Team").Clear()
	AssertAssociationCount(t, users, "Team", 0, "After Clear", db)
}
