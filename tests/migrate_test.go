package tests_test

import (
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

func _TestMigrate(t *testing.T) {
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
	allModels := []interface{}{&User{}, &Account{}, &Pet{}, &Company{}, &Toy{}, &Language{}}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(allModels), func(i, j int) { allModels[i], allModels[j] = allModels[j], allModels[i] })
	db.Migrator().DropTable("user_speaks", "user_friends", "ccc")

	if err := db.Migrator().DropTable(allModels...); err != nil {
		t.Fatalf("Failed to drop table, got error %v", err)
	}

	if err := db.AutoMigrate(allModels...); err != nil {
		t.Fatalf("Failed to auto migrate, but got error %v", err)
	}

	if tables, err := db.Migrator().GetTables(); err != nil {
		t.Fatalf("Failed to get database all tables, but got error %v", err)
	} else {
		for _, t1 := range []string{"users", "accounts", "pets", "companies", "toys", "languages"} {
			hasTable := false
			for _, t2 := range tables {
				if t2 == t1 {
					hasTable = true
					break
				}
			}
			if !hasTable {
				t.Fatalf("Failed to get table %v when GetTables", t1)
			}
		}
	}

	for _, m := range allModels {
		if !db.Migrator().HasTable(m) {
			t.Fatalf("Failed to create table for %#v---", m)
		}
	}

	db.Scopes(func(db *gorm.DB) *gorm.DB {
		return db.Table("ccc")
	}).Migrator().CreateTable(&Company{})

	if !db.Migrator().HasTable("ccc") {
		t.Errorf("failed to create table ccc")
	}

	for _, indexes := range [][2]string{
		{"user_speaks", "fk_user_speaks_user"},
		{"user_speaks", "fk_user_speaks_language"},
		{"user_friends", "fk_user_friends_user"},
		{"user_friends", "fk_user_friends_friends"},
		{"accounts", "fk_users_account"},
		{"users", "fk_users_team"},
		{"users", "fk_users_company"},
	} {
		if !db.Migrator().HasConstraint(indexes[0], indexes[1]) {
			t.Fatalf("Failed to find index for many2many for %v %v", indexes[0], indexes[1])
		}
	}
}

func _TestAutoMigrateSelfReferential(t *testing.T) {
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
	type MigratePerson struct {
		ID        uint
		Name      string
		ManagerID *uint
		Manager   *MigratePerson
	}

	db.Migrator().DropTable(&MigratePerson{})

	if err := db.AutoMigrate(&MigratePerson{}); err != nil {
		t.Fatalf("Failed to auto migrate, but got error %v", err)
	}

	if !db.Migrator().HasConstraint("migrate_people", "fk_migrate_people_manager") {
		t.Fatalf("Failed to find has one constraint between people and managers")
	}
}

func _TestSmartMigrateColumn(t *testing.T) {
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
	fullSupported := map[string]bool{"mysql": true}[db.Dialector.Name()]

	type UserMigrateColumn struct {
		ID       uint
		Name     string
		Salary   float64
		Birthday time.Time `gorm:"precision:4"`
	}

	db.Migrator().DropTable(&UserMigrateColumn{})

	db.AutoMigrate(&UserMigrateColumn{})

	type UserMigrateColumn2 struct {
		ID                  uint
		Name                string    `gorm:"size:128"`
		Salary              float64   `gorm:"precision:2"`
		Birthday            time.Time `gorm:"precision:2"`
		NameIgnoreMigration string    `gorm:"size:100"`
	}

	if err := db.Table("user_migrate_columns").AutoMigrate(&UserMigrateColumn2{}); err != nil {
		t.Fatalf("failed to auto migrate, got error: %v", err)
	}

	columnTypes, err := db.Table("user_migrate_columns").Migrator().ColumnTypes(&UserMigrateColumn{})
	if err != nil {
		t.Fatalf("failed to get column types, got error: %v", err)
	}

	for _, columnType := range columnTypes {
		switch columnType.Name() {
		case "name":
			if length, _ := columnType.Length(); (fullSupported || length != 0) && length != 128 {
				t.Fatalf("name's length should be 128, but got %v", length)
			}
		case "salary":
			if precision, o, _ := columnType.DecimalSize(); (fullSupported || precision != 0) && precision != 2 {
				t.Fatalf("salary's precision should be 2, but got %v %v", precision, o)
			}
		case "birthday":
			if precision, _, _ := columnType.DecimalSize(); (fullSupported || precision != 0) && precision != 2 {
				t.Fatalf("birthday's precision should be 2, but got %v", precision)
			}
		}
	}

	type UserMigrateColumn3 struct {
		ID                  uint
		Name                string    `gorm:"size:256"`
		Salary              float64   `gorm:"precision:3"`
		Birthday            time.Time `gorm:"precision:3"`
		NameIgnoreMigration string    `gorm:"size:128;-:migration"`
	}

	if err := db.Table("user_migrate_columns").AutoMigrate(&UserMigrateColumn3{}); err != nil {
		t.Fatalf("failed to auto migrate, got error: %v", err)
	}

	columnTypes, err = db.Table("user_migrate_columns").Migrator().ColumnTypes(&UserMigrateColumn{})
	if err != nil {
		t.Fatalf("failed to get column types, got error: %v", err)
	}

	for _, columnType := range columnTypes {
		switch columnType.Name() {
		case "name":
			if length, _ := columnType.Length(); (fullSupported || length != 0) && length != 256 {
				t.Fatalf("name's length should be 128, but got %v", length)
			}
		case "salary":
			if precision, _, _ := columnType.DecimalSize(); (fullSupported || precision != 0) && precision != 3 {
				t.Fatalf("salary's precision should be 2, but got %v", precision)
			}
		case "birthday":
			if precision, _, _ := columnType.DecimalSize(); (fullSupported || precision != 0) && precision != 3 {
				t.Fatalf("birthday's precision should be 2, but got %v", precision)
			}
		case "name_ignore_migration":
			if length, _ := columnType.Length(); (fullSupported || length != 0) && length != 100 {
				t.Fatalf("name_ignore_migration's length should still be 100 but got %v", length)
			}
		}
	}

}

func _TestMigrateWithColumnComment(t *testing.T) {
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
	type UserWithColumnComment struct {
		gorm.Model
		Name string `gorm:"size:111;comment:this is a 字段"`
	}

	if err := db.Migrator().DropTable(&UserWithColumnComment{}); err != nil {
		t.Fatalf("Failed to drop table, got error %v", err)
	}

	if err := db.AutoMigrate(&UserWithColumnComment{}); err != nil {
		t.Fatalf("Failed to auto migrate, but got error %v", err)
	}
}

func _TestMigrateWithIndexComment(t *testing.T) {
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
	if db.Dialector.Name() != "mysql" {
		t.Skip()
	}

	type UserWithIndexComment struct {
		gorm.Model
		Name string `gorm:"size:111;index:,comment:这是一个index"`
	}

	if err := db.Migrator().DropTable(&UserWithIndexComment{}); err != nil {
		t.Fatalf("Failed to drop table, got error %v", err)
	}

	if err := db.AutoMigrate(&UserWithIndexComment{}); err != nil {
		t.Fatalf("Failed to auto migrate, but got error %v", err)
	}
}

func _TestMigrateWithUniqueIndex(t *testing.T) {
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
	type UserWithUniqueIndex struct {
		ID   int
		Name string    `gorm:"size:20;index:idx_name,unique"`
		Date time.Time `gorm:"index:idx_name,unique"`
	}

	db.Migrator().DropTable(&UserWithUniqueIndex{})
	if err := db.AutoMigrate(&UserWithUniqueIndex{}); err != nil {
		t.Fatalf("failed to migrate, got %v", err)
	}

	if !db.Migrator().HasIndex(&UserWithUniqueIndex{}, "idx_name") {
		t.Errorf("Failed to find created index")
	}
}

func _TestMigrateTable(t *testing.T) {
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
	type TableStruct struct {
		gorm.Model
		Name string
	}

	db.Migrator().DropTable(&TableStruct{})
	db.AutoMigrate(&TableStruct{})

	if !db.Migrator().HasTable(&TableStruct{}) {
		t.Fatalf("should found created table")
	}

	type NewTableStruct struct {
		gorm.Model
		Name string
	}

	if err := db.Migrator().RenameTable(&TableStruct{}, &NewTableStruct{}); err != nil {
		t.Fatalf("Failed to rename table, got error %v", err)
	}

	if !db.Migrator().HasTable("new_table_structs") {
		t.Fatal("should found renamed table")
	}

	db.Migrator().DropTable("new_table_structs")

	if db.Migrator().HasTable(&NewTableStruct{}) {
		t.Fatal("should not found droped table")
	}
}

func _TestMigrateIndexes(t *testing.T) {
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
	type IndexStruct struct {
		gorm.Model
		Name string `gorm:"size:255;index"`
	}

	db.Migrator().DropTable(&IndexStruct{})
	db.AutoMigrate(&IndexStruct{})

	if err := db.Migrator().DropIndex(&IndexStruct{}, "Name"); err != nil {
		t.Fatalf("Failed to drop index for user's name, got err %v", err)
	}

	if err := db.Migrator().CreateIndex(&IndexStruct{}, "Name"); err != nil {
		t.Fatalf("Got error when tried to create index: %+v", err)
	}

	if !db.Migrator().HasIndex(&IndexStruct{}, "Name") {
		t.Fatalf("Failed to find index for user's name")
	}

	if err := db.Migrator().DropIndex(&IndexStruct{}, "Name"); err != nil {
		t.Fatalf("Failed to drop index for user's name, got err %v", err)
	}

	if db.Migrator().HasIndex(&IndexStruct{}, "Name") {
		t.Fatalf("Should not find index for user's name after delete")
	}

	if err := db.Migrator().CreateIndex(&IndexStruct{}, "Name"); err != nil {
		t.Fatalf("Got error when tried to create index: %+v", err)
	}

	if err := db.Migrator().RenameIndex(&IndexStruct{}, "idx_index_structs_name", "idx_users_name_1"); err != nil {
		t.Fatalf("no error should happen when rename index, but got %v", err)
	}

	if !db.Migrator().HasIndex(&IndexStruct{}, "idx_users_name_1") {
		t.Fatalf("Should find index for user's name after rename")
	}

	if err := db.Migrator().DropIndex(&IndexStruct{}, "idx_users_name_1"); err != nil {
		t.Fatalf("Failed to drop index for user's name, got err %v", err)
	}

	if db.Migrator().HasIndex(&IndexStruct{}, "idx_users_name_1") {
		t.Fatalf("Should not find index for user's name after delete")
	}
}

func _TestMigrateColumns(t *testing.T) {
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
	type ColumnStruct struct {
		gorm.Model
		Name string
	}

	db.Migrator().DropTable(&ColumnStruct{})

	if err := db.AutoMigrate(&ColumnStruct{}); err != nil {
		t.Errorf("Failed to migrate, got %v", err)
	}

	type ColumnStruct2 struct {
		gorm.Model
		Name string `gorm:"size:100"`
	}

	if err := db.Table("column_structs").Migrator().AlterColumn(&ColumnStruct2{}, "Name"); err != nil {
		t.Fatalf("no error should happened when alter column, but got %v", err)
	}

	if columnTypes, err := db.Migrator().ColumnTypes(&ColumnStruct{}); err != nil {
		t.Fatalf("no error should returns for ColumnTypes")
	} else {
		stmt := &gorm.Statement{DB: DB}
		stmt.Parse(&ColumnStruct2{})

		for _, columnType := range columnTypes {
			if columnType.Name() == "name" {
				dataType := db.Dialector.DataTypeOf(stmt.Schema.LookUpField(columnType.Name()))
				if !strings.Contains(strings.ToUpper(dataType), strings.ToUpper(columnType.DatabaseTypeName())) {
					t.Errorf("column type should be correct, name: %v, length: %v, expects: %v", columnType.Name(), columnType.DatabaseTypeName(), dataType)
				}
			}
		}
	}

	type NewColumnStruct struct {
		gorm.Model
		Name    string
		NewName string
	}

	if err := db.Table("column_structs").Migrator().AddColumn(&NewColumnStruct{}, "NewName"); err != nil {
		t.Fatalf("Failed to add column, got %v", err)
	}

	if !db.Table("column_structs").Migrator().HasColumn(&NewColumnStruct{}, "NewName") {
		t.Fatalf("Failed to find added column")
	}

	if err := db.Table("column_structs").Migrator().DropColumn(&NewColumnStruct{}, "NewName"); err != nil {
		t.Fatalf("Failed to add column, got %v", err)
	}

	if db.Table("column_structs").Migrator().HasColumn(&NewColumnStruct{}, "NewName") {
		t.Fatalf("Found deleted column")
	}

	if err := db.Table("column_structs").Migrator().AddColumn(&NewColumnStruct{}, "NewName"); err != nil {
		t.Fatalf("Failed to add column, got %v", err)
	}

	if err := db.Table("column_structs").Migrator().RenameColumn(&NewColumnStruct{}, "NewName", "new_new_name"); err != nil {
		t.Fatalf("Failed to add column, got %v", err)
	}

	if !db.Table("column_structs").Migrator().HasColumn(&NewColumnStruct{}, "new_new_name") {
		t.Fatalf("Failed to found renamed column")
	}

	if err := db.Table("column_structs").Migrator().DropColumn(&NewColumnStruct{}, "new_new_name"); err != nil {
		t.Fatalf("Failed to add column, got %v", err)
	}

	if db.Table("column_structs").Migrator().HasColumn(&NewColumnStruct{}, "new_new_name") {
		t.Fatalf("Found deleted column")
	}
}

func _TestMigrateConstraint(t *testing.T) {
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
	names := []string{"Account", "fk_users_account", "Pets", "fk_users_pets", "Company", "fk_users_company", "Team", "fk_users_team", "Languages", "fk_users_languages"}

	for _, name := range names {
		if !db.Migrator().HasConstraint(&User{}, name) {
			db.Migrator().CreateConstraint(&User{}, name)
		}

		if err := db.Migrator().DropConstraint(&User{}, name); err != nil {
			t.Fatalf("failed to drop constraint %v, got error %v", name, err)
		}

		if db.Migrator().HasConstraint(&User{}, name) {
			t.Fatalf("constraint %v should been deleted", name)
		}

		if err := db.Migrator().CreateConstraint(&User{}, name); err != nil {
			t.Fatalf("failed to create constraint %v, got error %v", name, err)
		}

		if !db.Migrator().HasConstraint(&User{}, name) {
			t.Fatalf("failed to found constraint %v", name)
		}
	}
}

type DynamicUser struct {
	gorm.Model
	Name      string
	CompanyID string `gorm:"index"`
}

// To test auto migrate crate indexes for dynamic table name
// https://github.com/go-gorm/gorm/issues/4752
func _TestMigrateIndexesWithDynamicTableName(t *testing.T) {
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
	// Create primary table
	if err := db.AutoMigrate(&DynamicUser{}); err != nil {
		t.Fatalf("AutoMigrate create table error: %#v", err)
	}

	// Create sub tables
	for _, v := range []string{"01", "02", "03"} {
		tableName := "dynamic_users_" + v
		m := db.Scopes(func(db *gorm.DB) *gorm.DB {
			return db.Table(tableName)
		}).Migrator()

		if err := m.AutoMigrate(&DynamicUser{}); err != nil {
			t.Fatalf("AutoMigrate create table error: %#v", err)
		}

		if !m.HasTable(tableName) {
			t.Fatalf("AutoMigrate expected %#v exist, but not.", tableName)
		}

		if !m.HasIndex(&DynamicUser{}, "CompanyID") {
			t.Fatalf("Should have index on %s", "CompanyI.")
		}

		if !m.HasIndex(&DynamicUser{}, "DeletedAt") {
			t.Fatalf("Should have index on deleted_at.")
		}
	}
}
