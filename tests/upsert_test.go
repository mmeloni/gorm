package tests_test

import (
	"os"
	"regexp"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	. "gorm.io/gorm/utils/tests"
)

func TestUpsert(t *testing.T) {
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
	lang := Language{Code: "upsert", Name: "Upsert"}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&lang).Error; err != nil {
		t.Fatalf("failed to upsert, got %v", err)
	}

	lang2 := Language{Code: "upsert", Name: "Upsert"}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&lang2).Error; err != nil {
		t.Fatalf("failed to upsert, got %v", err)
	}

	var langs []Language
	if err := db.Find(&langs, "code = ?", lang.Code).Error; err != nil {
		t.Errorf("no error should happen when find languages with code, but got %v", err)
	} else if len(langs) != 1 {
		t.Errorf("should only find only 1 languages, but got %+v", langs)
	}

	lang3 := Language{Code: "upsert", Name: "Upsert"}
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"name": "upsert-new"}),
	}).Create(&lang3).Error; err != nil {
		t.Fatalf("failed to upsert, got %v", err)
	}

	if err := db.Find(&langs, "code = ?", lang.Code).Error; err != nil {
		t.Errorf("no error should happen when find languages with code, but got %v", err)
	} else if len(langs) != 1 {
		t.Errorf("should only find only 1 languages, but got %+v", langs)
	} else if langs[0].Name != "upsert-new" {
		t.Errorf("should update name on conflict, but got name %+v", langs[0].Name)
	}

	lang = Language{Code: "upsert", Name: "Upsert-Newname"}
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&lang).Error; err != nil {
		t.Fatalf("failed to upsert, got %v", err)
	}

	var result Language
	if err := db.Find(&result, "code = ?", lang.Code).Error; err != nil || result.Name != lang.Name {
		t.Fatalf("failed to upsert, got name %v", result.Name)
	}

	if name := db.Dialector.Name(); name != "sqlserver" {
		type RestrictedLanguage struct {
			Code string `gorm:"primarykey"`
			Name string
			Lang string `gorm:"<-:create"`
		}

		r := db.Session(&gorm.Session{DryRun: true}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&RestrictedLanguage{Code: "upsert_code", Name: "upsert_name", Lang: "upsert_lang"})
		if !regexp.MustCompile(`INTO .restricted_languages. .*\(.code.,.name.,.lang.\) .* (SET|UPDATE) .name.=.*.name.[^\w]*$`).MatchString(r.Statement.SQL.String()) {
			t.Errorf("Table with escape character, got %v", r.Statement.SQL.String())
		}
	}

	var user = *GetUser("upsert_on_conflict", Config{})
	user.Age = 20
	if err := db.Create(&user).Error; err != nil {
		t.Errorf("failed to create user, got error %v", err)
	}

	var user2 User
	db.First(&user2, user.ID)
	user2.Age = 30
	time.Sleep(time.Second)
	if err := db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&user2).Error; err != nil {
		t.Fatalf("failed to onconflict create user, got error %v", err)
	} else {
		var user3 User
		db.First(&user3, user.ID)
		if user3.UpdatedAt.UnixNano() == user2.UpdatedAt.UnixNano() {
			t.Fatalf("failed to update user's updated_at, old: %v, new: %v", user2.UpdatedAt, user3.UpdatedAt)
		}
	}
}

func TestUpsertSlice(t *testing.T) {
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
	langs := []Language{
		{Code: "upsert-slice1", Name: "Upsert-slice1"},
		{Code: "upsert-slice2", Name: "Upsert-slice2"},
		{Code: "upsert-slice3", Name: "Upsert-slice3"},
	}
	db.Clauses(clause.OnConflict{DoNothing: true}).Create(&langs)

	var langs2 []Language
	if err := db.Find(&langs2, "code LIKE ?", "upsert-slice%").Error; err != nil {
		t.Errorf("no error should happen when find languages with code, but got %v", err)
	} else if len(langs2) != 3 {
		t.Errorf("should only find only 3 languages, but got %+v", langs2)
	}

	db.Clauses(clause.OnConflict{DoNothing: true}).Create(&langs)
	var langs3 []Language
	if err := db.Find(&langs3, "code LIKE ?", "upsert-slice%").Error; err != nil {
		t.Errorf("no error should happen when find languages with code, but got %v", err)
	} else if len(langs3) != 3 {
		t.Errorf("should only find only 3 languages, but got %+v", langs3)
	}

	for idx, lang := range langs {
		lang.Name = lang.Name + "_new"
		langs[idx] = lang
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "code"}},
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Create(&langs).Error; err != nil {
		t.Fatalf("failed to upsert, got %v", err)
	}

	for _, lang := range langs {
		var results []Language
		if err := db.Find(&results, "code = ?", lang.Code).Error; err != nil {
			t.Errorf("no error should happen when find languages with code, but got %v", err)
		} else if len(results) != 1 {
			t.Errorf("should only find only 1 languages, but got %+v", langs)
		} else if results[0].Name != lang.Name {
			t.Errorf("should update name on conflict, but got name %+v", results[0].Name)
		}
	}
}

func TestUpsertWithSave(t *testing.T) {
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
	langs := []Language{
		{Code: "upsert-save-1", Name: "Upsert-save-1"},
		{Code: "upsert-save-2", Name: "Upsert-save-2"},
	}

	if err := db.Save(&langs).Error; err != nil {
		t.Errorf("Failed to create, got error %v", err)
	}

	for _, lang := range langs {
		var result Language
		if err := db.First(&result, "code = ?", lang.Code).Error; err != nil {
			t.Errorf("Failed to query lang, got error %v", err)
		} else {
			AssertEqual(t, result, lang)
		}
	}

	for idx, lang := range langs {
		lang.Name += "_new"
		langs[idx] = lang
	}

	if err := db.Save(&langs).Error; err != nil {
		t.Errorf("Failed to upsert, got error %v", err)
	}

	for _, lang := range langs {
		var result Language
		if err := db.First(&result, "code = ?", lang.Code).Error; err != nil {
			t.Errorf("Failed to query lang, got error %v", err)
		} else {
			AssertEqual(t, result, lang)
		}
	}

	lang := Language{Code: "upsert-save-3", Name: "Upsert-save-3"}
	if err := db.Save(&lang).Error; err != nil {
		t.Errorf("Failed to create, got error %v", err)
	}

	var result Language
	if err := db.First(&result, "code = ?", lang.Code).Error; err != nil {
		t.Errorf("Failed to query lang, got error %v", err)
	} else {
		AssertEqual(t, result, lang)
	}

	lang.Name += "_new"
	if err := db.Save(&lang).Error; err != nil {
		t.Errorf("Failed to create, got error %v", err)
	}

	var result2 Language
	if err := db.First(&result2, "code = ?", lang.Code).Error; err != nil {
		t.Errorf("Failed to query lang, got error %v", err)
	} else {
		AssertEqual(t, result2, lang)
	}
}

func TestFindOrInitialize(t *testing.T) {
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
	var user1, user2, user3, user4, user5, user6 User
	if err := db.Where(&User{Name: "find or init", Age: 33}).FirstOrInit(&user1).Error; err != nil {
		t.Errorf("no error should happen when FirstOrInit, but got %v", err)
	}

	if user1.Name != "find or init" || user1.ID != 0 || user1.Age != 33 {
		t.Errorf("user should be initialized with search value")
	}

	db.Where(User{Name: "find or init", Age: 33}).FirstOrInit(&user2)
	if user2.Name != "find or init" || user2.ID != 0 || user2.Age != 33 {
		t.Errorf("user should be initialized with search value")
	}

	db.FirstOrInit(&user3, map[string]interface{}{"name": "find or init 2"})
	if user3.Name != "find or init 2" || user3.ID != 0 {
		t.Errorf("user should be initialized with inline search value")
	}

	db.Where(&User{Name: "find or init"}).Attrs(User{Age: 44}).FirstOrInit(&user4)
	if user4.Name != "find or init" || user4.ID != 0 || user4.Age != 44 {
		t.Errorf("user should be initialized with search value and attrs")
	}

	db.Where(&User{Name: "find or init"}).Assign("age", 44).FirstOrInit(&user4)
	if user4.Name != "find or init" || user4.ID != 0 || user4.Age != 44 {
		t.Errorf("user should be initialized with search value and assign attrs")
	}

	db.Save(&User{Name: "find or init", Age: 33})
	db.Where(&User{Name: "find or init"}).Attrs("age", 44).FirstOrInit(&user5)
	if user5.Name != "find or init" || user5.ID == 0 || user5.Age != 33 {
		t.Errorf("user should be found and not initialized by Attrs")
	}

	db.Where(&User{Name: "find or init", Age: 33}).FirstOrInit(&user6)
	if user6.Name != "find or init" || user6.ID == 0 || user6.Age != 33 {
		t.Errorf("user should be found with FirstOrInit")
	}

	db.Where(&User{Name: "find or init"}).Assign(User{Age: 44}).FirstOrInit(&user6)
	if user6.Name != "find or init" || user6.ID == 0 || user6.Age != 44 {
		t.Errorf("user should be found and updated with assigned attrs")
	}
}

func TestFindOrCreate(t *testing.T) {
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
	var user1, user2, user3, user4, user5, user6, user7, user8 User
	if err := db.Where(&User{Name: "find or create", Age: 33}).FirstOrCreate(&user1).Error; err != nil {
		t.Errorf("no error should happen when FirstOrInit, but got %v", err)
	}

	if user1.Name != "find or create" || user1.ID == 0 || user1.Age != 33 {
		t.Errorf("user should be created with search value")
	}

	db.Where(&User{Name: "find or create", Age: 33}).FirstOrCreate(&user2)
	if user1.ID != user2.ID || user2.Name != "find or create" || user2.ID == 0 || user2.Age != 33 {
		t.Errorf("user should be created with search value")
	}

	db.FirstOrCreate(&user3, map[string]interface{}{"name": "find or create 2"})
	if user3.Name != "find or create 2" || user3.ID == 0 {
		t.Errorf("user should be created with inline search value")
	}

	db.Where(&User{Name: "find or create 3"}).Attrs("age", 44).FirstOrCreate(&user4)
	if user4.Name != "find or create 3" || user4.ID == 0 || user4.Age != 44 {
		t.Errorf("user should be created with search value and attrs")
	}

	updatedAt1 := user4.UpdatedAt
	db.Where(&User{Name: "find or create 3"}).Assign("age", 55).FirstOrCreate(&user4)

	if user4.Age != 55 {
		t.Errorf("Failed to set change to 55, got %v", user4.Age)
	}

	if updatedAt1.Format(time.RFC3339Nano) == user4.UpdatedAt.Format(time.RFC3339Nano) {
		t.Errorf("UpdateAt should be changed when update values with assign")
	}

	db.Where(&User{Name: "find or create 4"}).Assign(User{Age: 44}).FirstOrCreate(&user4)
	if user4.Name != "find or create 4" || user4.ID == 0 || user4.Age != 44 {
		t.Errorf("user should be created with search value and assigned attrs")
	}

	db.Where(&User{Name: "find or create"}).Attrs("age", 44).FirstOrInit(&user5)
	if user5.Name != "find or create" || user5.ID == 0 || user5.Age != 33 {
		t.Errorf("user should be found and not initialized by Attrs")
	}

	db.Where(&User{Name: "find or create"}).Assign(User{Age: 44}).FirstOrCreate(&user6)
	if user6.Name != "find or create" || user6.ID == 0 || user6.Age != 44 {
		t.Errorf("user should be found and updated with assigned attrs")
	}

	db.Where(&User{Name: "find or create"}).Find(&user7)
	if user7.Name != "find or create" || user7.ID == 0 || user7.Age != 44 {
		t.Errorf("user should be found and updated with assigned attrs")
	}

	db.Where(&User{Name: "find or create embedded struct"}).Assign(User{Age: 44, Account: Account{Number: "1231231231"}, Pets: []*Pet{{Name: "first_or_create_pet1"}, {Name: "first_or_create_pet2"}}}).FirstOrCreate(&user8)
	if err := db.Where("name = ?", "first_or_create_pet1").First(&Pet{}).Error; err != nil {
		t.Errorf("has many association should be saved")
	}

	if err := db.Where("number = ?", "1231231231").First(&Account{}).Error; err != nil {
		t.Errorf("belongs to association should be saved")
	}
}

func TestUpdateWithMissWhere(t *testing.T) {
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
	type User struct {
		ID   uint   `gorm:"column:id;<-:create"`
		Name string `gorm:"column:name"`
	}
	user := User{ID: 1, Name: "king"}
	tx := db.Session(&gorm.Session{DryRun: true}).Save(&user)

	if err := tx.Error; err != nil {
		t.Fatalf("failed to update user,missing where condtion,err=%+v", err)

	}

	if !regexp.MustCompile("WHERE .id. = [^ ]+$").MatchString(tx.Statement.SQL.String()) {
		t.Fatalf("invalid updating SQL, got %v", tx.Statement.SQL.String())
	}

}
