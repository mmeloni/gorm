package tests_test

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	. "gorm.io/gorm/utils/tests"
)

func _TestPreloadWithAssociations(t *testing.T) {
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
	var user = *GetUser("preload_with_associations", Config{
		Account:   true,
		Pets:      2,
		Toys:      3,
		Company:   true,
		Manager:   true,
		Team:      4,
		Languages: 3,
		Friends:   1,
	})

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	CheckUser(t, user, user, db)

	var user2 User
	db.Preload(clause.Associations).Find(&user2, "id = ?", user.ID)
	CheckUser(t, user2, user, db)

	var user3 = *GetUser("preload_with_associations_new", Config{
		Account:   true,
		Pets:      2,
		Toys:      3,
		Company:   true,
		Manager:   true,
		Team:      4,
		Languages: 3,
		Friends:   1,
	})

	db.Preload(clause.Associations).Find(&user3, "id = ?", user.ID)
	CheckUser(t, user3, user, db)
}

func _TestNestedPreload(t *testing.T) {
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
	var user = *GetUser("nested_preload", Config{Pets: 2})

	for idx, pet := range user.Pets {
		pet.Toy = Toy{Name: "toy_nested_preload_" + strconv.Itoa(idx+1)}
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	var user2 User
	db.Preload("Pets.Toy").Find(&user2, "id = ?", user.ID)
	CheckUser(t, user2, user, db)

	var user3 User
	db.Preload(clause.Associations+"."+clause.Associations).Find(&user3, "id = ?", user.ID)
	CheckUser(t, user3, user, db)

	var user4 *User
	db.Preload("Pets.Toy").Find(&user4, "id = ?", user.ID)
	CheckUser(t, *user4, user, db)
}

func _TestNestedPreloadForSlice(t *testing.T) {
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
		*GetUser("slice_nested_preload_1", Config{Pets: 2}),
		*GetUser("slice_nested_preload_2", Config{Pets: 0}),
		*GetUser("slice_nested_preload_3", Config{Pets: 3}),
	}

	for _, user := range users {
		for idx, pet := range user.Pets {
			pet.Toy = Toy{Name: user.Name + "_toy_nested_preload_" + strconv.Itoa(idx+1)}
		}
	}

	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var users2 []User
	db.Preload("Pets.Toy").Find(&users2, "id IN ?", userIDs)

	for idx, user := range users2 {
		CheckUser(t, user, users[idx], db)
	}
}

func _TestPreloadWithConds(t *testing.T) {
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
		*GetUser("slice_nested_preload_1", Config{Account: true}),
		*GetUser("slice_nested_preload_2", Config{Account: false}),
		*GetUser("slice_nested_preload_3", Config{Account: true}),
	}

	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var users2 []User
	db.Preload("Account", clause.Eq{Column: "number", Value: users[0].Account.Number}).Find(&users2, "id IN ?", userIDs)
	sort.Slice(users2, func(i, j int) bool {
		return users2[i].ID < users2[j].ID
	})

	for idx, user := range users2[1:2] {
		if user.Account.Number != "" {
			t.Errorf("No account should found for user %v but got %v", idx+2, user.Account.Number)
		}
	}

	CheckUser(t, users2[0], users[0], db)

	var users3 []User
	if err := db.Preload("Account", func(tx *gorm.DB) *gorm.DB {
		return tx.Table("accounts AS a").Select("a.*")
	}).Find(&users3, "id IN ?", userIDs).Error; err != nil {
		t.Errorf("failed to query, got error %v", err)
	}
	sort.Slice(users3, func(i, j int) bool {
		return users2[i].ID < users2[j].ID
	})

	for i, u := range users3 {
		CheckUser(t, u, users[i], db)
	}

	var user4 User
	db.Delete(&users3[0].Account)

	if err := db.Preload(clause.Associations).Take(&user4, "id = ?", users3[0].ID).Error; err != nil || user4.Account.ID != 0 {
		t.Errorf("failed to query, got error %v, account: %#v", err, user4.Account)
	}

	if err := db.Preload(clause.Associations, func(tx *gorm.DB) *gorm.DB {
		return tx.Unscoped()
	}).Take(&user4, "id = ?", users3[0].ID).Error; err != nil || user4.Account.ID == 0 {
		t.Errorf("failed to query, got error %v, account: %#v", err, user4.Account)
	}
}

func _TestNestedPreloadWithConds(t *testing.T) {
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
		*GetUser("slice_nested_preload_1", Config{Pets: 2}),
		*GetUser("slice_nested_preload_2", Config{Pets: 0}),
		*GetUser("slice_nested_preload_3", Config{Pets: 3}),
	}

	for _, user := range users {
		for idx, pet := range user.Pets {
			pet.Toy = Toy{Name: user.Name + "_toy_nested_preload_" + strconv.Itoa(idx+1)}
		}
	}

	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("errors happened when create: %v", err)
	}

	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	var users2 []User
	db.Preload("Pets.Toy", "name like ?", `%preload_3`).Find(&users2, "id IN ?", userIDs)

	for idx, user := range users2[0:2] {
		for _, pet := range user.Pets {
			if pet.Toy.Name != "" {
				t.Errorf("No toy should for user %v's pet %v but got %v", idx+1, pet.Name, pet.Toy.Name)
			}
		}
	}

	if len(users2[2].Pets) != 3 {
		t.Errorf("Invalid pet toys found for user 3 got %v", len(users2[2].Pets))
	} else {
		sort.Slice(users2[2].Pets, func(i, j int) bool {
			return users2[2].Pets[i].ID < users2[2].Pets[j].ID
		})

		for _, pet := range users2[2].Pets[0:2] {
			if pet.Toy.Name != "" {
				t.Errorf("No toy should for user %v's pet %v but got %v", 3, pet.Name, pet.Toy.Name)
			}
		}

		CheckPet(t, *users2[2].Pets[2], *users[2].Pets[2], db)
	}
}

func _TestPreloadEmptyData(t *testing.T) {
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
	var user = *GetUser("user_without_associations", Config{})
	db.Create(&user)

	db.Preload("Team").Preload("Languages").Preload("Friends").First(&user, "name = ?", user.Name)

	if r, err := json.Marshal(&user); err != nil {
		t.Errorf("failed to marshal users, got error %v", err)
	} else if !regexp.MustCompile(`"Team":\[\],"Languages":\[\],"Friends":\[\]`).MatchString(string(r)) {
		t.Errorf("json marshal is not empty slice, got %v", string(r))
	}

	var results []User
	db.Preload("Team").Preload("Languages").Preload("Friends").Find(&results, "name = ?", user.Name)

	if r, err := json.Marshal(&results); err != nil {
		t.Errorf("failed to marshal users, got error %v", err)
	} else if !regexp.MustCompile(`"Team":\[\],"Languages":\[\],"Friends":\[\]`).MatchString(string(r)) {
		t.Errorf("json marshal is not empty slice, got %v", string(r))
	}
}

func _TestPreloadGoroutine(t *testing.T) {
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
	var wg sync.WaitGroup

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			var user2 []User
			tx := db.Where("id = ?", 1).Session(&gorm.Session{})

			if err := tx.Preload("Team").Find(&user2).Error; err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}
