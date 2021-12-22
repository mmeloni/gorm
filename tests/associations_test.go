package tests_test

import (
	"os"
	"testing"

	"gorm.io/gorm"
	. "gorm.io/gorm/utils/tests"
)

func AssertAssociationCount(t *testing.T, data interface{}, name string, result int64, reason string, db *gorm.DB) {

	if count := db.Model(data).Association(name).Count(); count != result {
		t.Fatalf("invalid %v count %v, expects: %v got %v", name, reason, result, count)
	}

	var newUser User
	if user, ok := data.(User); ok {
		db.Find(&newUser, "id = ?", user.ID)
	} else if user, ok := data.(*User); ok {
		db.Find(&newUser, "id = ?", user.ID)
	}

	if newUser.ID != 0 {
		if count := db.Model(&newUser).Association(name).Count(); count != result {
			t.Fatalf("invalid %v count %v, expects: %v got %v", name, reason, result, count)
		}
	}
}

func TestInvalidAssociation(t *testing.T) {
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
	var user = *GetUser("invalid", Config{Company: true, Manager: true})
	if err := db.Model(&user).Association("Invalid").Find(&user.Company).Error; err == nil {
		t.Fatalf("should return errors for invalid association, but got nil")
	}
}

func TestAssociationNotNullClear(t *testing.T) {
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
	type Profile struct {
		gorm.Model
		Number   string
		MemberID uint `gorm:"not null"`
	}

	type Member struct {
		gorm.Model
		Profiles []Profile
	}

	db.Migrator().DropTable(&Member{}, &Profile{})

	if err := db.AutoMigrate(&Member{}, &Profile{}); err != nil {
		t.Fatalf("Failed to migrate, got error: %v", err)
	}

	member := &Member{
		Profiles: []Profile{{
			Number: "1",
		}, {
			Number: "2",
		}},
	}

	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("Failed to create test data, got error: %v", err)
	}

	if err := db.Model(member).Association("Profiles").Clear(); err == nil {
		t.Fatalf("No error occurred during clearind not null association")
	}
}

func TestForeignKeyConstraints(t *testing.T) {
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
	type Profile struct {
		ID       uint
		Name     string
		MemberID uint
	}

	type Member struct {
		ID      uint
		Refer   uint `gorm:"uniqueIndex"`
		Name    string
		Profile Profile `gorm:"Constraint:OnUpdate:CASCADE,OnDelete:CASCADE;FOREIGNKEY:MemberID;References:Refer"`
	}

	db.Migrator().DropTable(&Profile{}, &Member{})

	if err := db.AutoMigrate(&Profile{}, &Member{}); err != nil {
		t.Fatalf("Failed to migrate, got error: %v", err)
	}

	member := Member{Refer: 1, Name: "foreign_key_constraints", Profile: Profile{Name: "my_profile"}}

	db.Create(&member)

	var profile Profile
	if err := db.First(&profile, "id = ?", member.Profile.ID).Error; err != nil {
		t.Fatalf("failed to find profile, got error: %v", err)
	} else if profile.MemberID != member.ID {
		t.Fatalf("member id is not equal: expects: %v, got: %v", member.ID, profile.MemberID)
	}

	member.Profile = Profile{}
	db.Model(&member).Update("Refer", 100)

	var profile2 Profile
	if err := db.First(&profile2, "id = ?", profile.ID).Error; err != nil {
		t.Fatalf("failed to find profile, got error: %v", err)
	} else if profile2.MemberID != 100 {
		t.Fatalf("member id is not equal: expects: %v, got: %v", 100, profile2.MemberID)
	}

	if r := db.Delete(&member); r.Error != nil || r.RowsAffected != 1 {
		t.Fatalf("Should delete member, got error: %v, affected: %v", r.Error, r.RowsAffected)
	}

	var result Member
	if err := db.First(&result, member.ID).Error; err == nil {
		t.Fatalf("Should not find deleted member")
	}

	if err := db.First(&profile2, profile.ID).Error; err == nil {
		t.Fatalf("Should not find deleted profile")
	}
}

func TestForeignKeyConstraintsBelongsTo(t *testing.T) {
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
	type Profile struct {
		ID    uint
		Name  string
		Refer uint `gorm:"uniqueIndex"`
	}

	type Member struct {
		ID        uint
		Name      string
		ProfileID uint
		Profile   Profile `gorm:"Constraint:OnUpdate:CASCADE,OnDelete:CASCADE;FOREIGNKEY:ProfileID;References:Refer"`
	}

	db.Migrator().DropTable(&Profile{}, &Member{})

	if err := db.AutoMigrate(&Profile{}, &Member{}); err != nil {
		t.Fatalf("Failed to migrate, got error: %v", err)
	}

	member := Member{Name: "foreign_key_constraints_belongs_to", Profile: Profile{Name: "my_profile_belongs_to", Refer: 1}}

	db.Create(&member)

	var profile Profile
	if err := db.First(&profile, "id = ?", member.Profile.ID).Error; err != nil {
		t.Fatalf("failed to find profile, got error: %v", err)
	} else if profile.Refer != member.ProfileID {
		t.Fatalf("member id is not equal: expects: %v, got: %v", profile.Refer, member.ProfileID)
	}

	db.Model(&profile).Update("Refer", 100)

	var member2 Member
	if err := db.First(&member2, "id = ?", member.ID).Error; err != nil {
		t.Fatalf("failed to find member, got error: %v", err)
	} else if member2.ProfileID != 100 {
		t.Fatalf("member id is not equal: expects: %v, got: %v", 100, member2.ProfileID)
	}

	if r := db.Delete(&profile); r.Error != nil || r.RowsAffected != 1 {
		t.Fatalf("Should delete member, got error: %v, affected: %v", r.Error, r.RowsAffected)
	}

	var result Member
	if err := db.First(&result, member.ID).Error; err == nil {
		t.Fatalf("Should not find deleted member")
	}

	if err := db.First(&profile, profile.ID).Error; err == nil {
		t.Fatalf("Should not find deleted profile")
	}
}

func _TestFullSaveAssociations(t *testing.T) {
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
	coupon := &Coupon{
		AppliesToProduct: []*CouponProduct{
			{ProductId: "full-save-association-product1"},
		},
		AmountOff:  10,
		PercentOff: 0.0,
	}

	err = db.
		Session(&gorm.Session{FullSaveAssociations: true}).
		Create(coupon).Error

	if err != nil {
		t.Errorf("Failed, got error: %v", err)
	}

	if db.First(&Coupon{}, "id = ?", coupon.ID).Error != nil {
		t.Errorf("Failed to query saved coupon")
	}

	if db.First(&CouponProduct{}, "coupon_id = ? AND product_id = ?", coupon.ID, "full-save-association-product1").Error != nil {
		t.Errorf("Failed to query saved association")
	}

	orders := []Order{{Num: "order1", Coupon: coupon}, {Num: "order2", Coupon: coupon}}
	if err := db.Create(&orders).Error; err != nil {
		t.Errorf("failed to create orders, got %v", err)
	}

	coupon2 := Coupon{
		AppliesToProduct: []*CouponProduct{{Desc: "coupon-description"}},
	}

	db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&coupon2)
	var result Coupon
	if err := db.Preload("AppliesToProduct").First(&result, "id = ?", coupon2.ID).Error; err != nil {
		t.Errorf("Failed to create coupon w/o name, got error: %v", err)
	}

	if len(result.AppliesToProduct) != 1 {
		t.Errorf("Failed to preload AppliesToProduct")
	}
}
