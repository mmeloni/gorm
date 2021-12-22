package tests_test

import (
	"gorm.io/gorm"
	"os"
	"testing"

	. "gorm.io/gorm/utils/tests"
)

func TestExceptionsWithInvalidSql(t *testing.T) {
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
	if name := db.Dialector.Name(); name == "sqlserver" {
		t.Skip("skip sqlserver due to it will raise data race for invalid sql")
	}

	var columns []string
	if db.Where("sdsd.zaaa = ?", "sd;;;aa").Pluck("aaa", &columns).Error == nil {
		t.Errorf("Should got error with invalid SQL")
	}

	if db.Model(&User{}).Where("sdsd.zaaa = ?", "sd;;;aa").Pluck("aaa", &columns).Error == nil {
		t.Errorf("Should got error with invalid SQL")
	}

	if db.Where("sdsd.zaaa = ?", "sd;;;aa").Find(&User{}).Error == nil {
		t.Errorf("Should got error with invalid SQL")
	}

	var count1, count2 int64
	db.Model(&User{}).Count(&count1)
	if count1 <= 0 {
		t.Errorf("Should find some users")
	}

	if db.Where("name = ?", "jinzhu; delete * from users").First(&User{}).Error == nil {
		t.Errorf("Should got error with invalid SQL")
	}

	db.Model(&User{}).Count(&count2)
	if count1 != count2 {
		t.Errorf("No user should not be deleted by invalid SQL")
	}
}

func TestSetAndGet(t *testing.T) {
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
	if value, ok := db.Set("hello", "world").Get("hello"); !ok {
		t.Errorf("Should be able to get setting after set")
	} else {
		if value.(string) != "world" {
			t.Errorf("Setted value should not be changed")
		}
	}

	if _, ok := db.Get("non_existing"); ok {
		t.Errorf("Get non existing key should return error")
	}
}
