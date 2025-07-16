//go:build integration
// +build integration

package idx

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIdForMongo(t *testing.T) {
	type IdTestStruct struct {
		ID    ID     `bson:"_id"`
		Value string `bson:"value"`
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://root:password@mongo:27017/?maxPoolSize=5&w=majority").SetServerAPIOptions(serverAPI)
	c := context.TODO()
	client, err := mongo.Connect(opts)
	if err != nil {
		t.Fatalf("Error connecting server: %v", err)
	}
	db := client.Database("idx_test").Collection("idx_test")
	defer func() {
		if err = db.Drop(c); err != nil {
			panic(err)
		}
		if err = client.Disconnect(c); err != nil {
			panic(err)
		}
	}()

	data := IdTestStruct{
		ID:    NewID(),
		Value: "test-1",
	}
	if _, err = db.InsertOne(c, &data); err != nil {
		t.Fatalf("Error inserting record: %v", err)
	}

	var actualData IdTestStruct
	if err = db.FindOne(c, bson.D{{"_id", data.ID}}).Decode(&actualData); err != nil {
		t.Fatalf("Error retrieving record: %v", err)
	}
	if data.ID != actualData.ID || data.Value != actualData.Value {
		t.Fatalf("Original value did not match with actual value")
	}

	if _, err = db.UpdateOne(c, bson.D{{"_id", data.ID}}, bson.D{{"$set", bson.D{{"value", "test-2"}}}}); err != nil {
		t.Fatalf("Error updating record: %v", err)
	}
	if err = db.FindOne(c, bson.D{{"_id", data.ID}}).Decode(&actualData); err != nil {
		t.Fatalf("Error retrieving record: %v", err)
	}
	if data.ID != actualData.ID || actualData.Value != "test-2" {
		t.Fatalf("Original value did not match with actual value")
	}

	if _, err = db.DeleteOne(c, bson.D{{"_id", data.ID}}); err != nil {
		t.Fatalf("Error deleting record: %v", err)
	}
	if err = db.FindOne(c, bson.D{{"_id", data.ID}}).Decode(&actualData); !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("Was expecting no document error, got %v", err)
	}
}

func TestIdForMySQL(t *testing.T) {
	type IdTestStruct struct {
		ID    ID     `gorm:"column:id"`
		FkID  ID     `gorm:"column:fk_id"`
		Value string `gorm:"column:value"`
	}
	dsn := "root:password@tcp(mariadb:3306)/?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("MySQL Open error: %v", err)
	}
	if err = db.Exec("CREATE DATABASE IF NOT EXISTS `id_experiment` COLLATE 'utf8mb4_unicode_ci';").Error; err != nil {
		t.Fatalf("MySQL database creation error: %v", err)
	}

	dsn = "root:password@tcp(mariadb:3306)/id_experiment?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("MySQL Open error: %v", err)
	}
	defer func() {
		if err = db.Exec("DROP TABLE IF EXISTS `id_test_structs`;").Error; err != nil {
			t.Fatalf("MySQL table drop error: %v", err)
		}
		if err = db.Exec("DROP DATABASE IF EXISTS `id_experiment`;").Error; err != nil {
			t.Fatalf("MySQL database drop error: %v", err)
		}
	}()
	table := `
	CREATE TABLE IF NOT EXISTS id_test_structs (
  		id binary(16) NOT NULL,
  		value text NOT NULL,
	    fk_id binary(16) DEFAULT NULL,
  		PRIMARY KEY (id)
	)ENGINE=InnoDB;
	`
	if err = db.Exec(table).Error; err != nil {
		t.Fatalf("MySQL table creation error: %v", err)
	}
	data := IdTestStruct{
		ID:    NewID(),
		Value: "test-1",
	}
	// Test create
	if err = db.Create(&data).Error; err != nil {
		t.Fatalf("Error while creating: %v", err)
	}
	// Test retrieve
	result := IdTestStruct{}
	if err = db.First(&result, "id = ?", data.ID).Error; err != nil {
		t.Fatalf("Error while selecting: %v", err)
	}
	if data.ID != result.ID || data.Value != result.Value || data.FkID != NilID {
		t.Fatalf("Original value did not match with actual value")
	}
	// Test update
	newTestId := NewID()
	if err = db.Model(&data).Where("id = ?", data.ID).Updates(map[string]interface{}{"value": "test-2", "fk_id": newTestId}).Error; err != nil {
		t.Fatalf("Error while updating: %v", err)
	}
	if err = db.First(&result, "id = ?", data.ID).Error; err != nil {
		t.Fatalf("Error while selecting: %v", err)
	}
	if data.ID != result.ID || result.Value != "test-2" || result.FkID != newTestId {
		t.Fatalf("Updated record value did not match")
	}
	// Test delete
	if err = db.Where("id = ?", data.ID).Delete(&data).Error; err != nil {
		t.Fatalf("Error while deleting: %v", err)
	}
	result = IdTestStruct{}
	if err = db.First(&result, "id = ?", data.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Record found even though it should be deleted")
	}
}

func TestIdForPostgres(t *testing.T) {
	type IdTestStruct struct {
		ID    ID     `gorm:"column:id"`
		FkID  ID     `gorm:"column:fk_id"`
		Value string `gorm:"column:value"`
	}
	dsnOp := "host=postgres user=postgres password=password port=5432 sslmode=disable TimeZone=UTC"
	dbOp, err := gorm.Open(postgres.Open(dsnOp), &gorm.Config{})
	if err != nil {
		t.Fatalf("Postgres Open error: %v", err)
	}
	if err = dbOp.Exec("CREATE DATABASE id_experiment;").Error; err != nil {
		t.Fatalf("Postgres database creation error: %v", err)
	}

	dsn := "host=postgres user=postgres password=password dbname=id_experiment port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Postgres Open error: %v", err)
	}
	defer func() {
		if err = db.Exec("DROP TABLE IF EXISTS id_test_structs;").Error; err != nil {
			t.Fatalf("Postgres table drop error: %v", err)
		}
		sqlDb, err := db.DB()
		if err != nil {
			t.Fatalf("Conversion to sql interface error: %v", err)
		}
		if err = sqlDb.Close(); err != nil {
			t.Fatalf("Database connection closing error: %v", err)
		}
		if err = dbOp.Exec("DROP DATABASE IF EXISTS id_experiment;").Error; err != nil {
			t.Fatalf("Postgres database drop error: %v", err)
		}
	}()
	table := `
		CREATE TABLE IF NOT EXISTS id_test_structs (
	  		id bytea NOT NULL,
	  		value text NOT NULL,
	  		fk_id bytea DEFAULT NULL,
	  		PRIMARY KEY (id)
		);
	`
	if err = db.Exec(table).Error; err != nil {
		t.Fatalf("Postgres table creation error: %v", err)
	}
	data := IdTestStruct{
		ID:    NewID(),
		Value: "test-1",
	}
	// Test create
	if err = db.Create(&data).Error; err != nil {
		t.Fatalf("Error while creating: %v", err)
	}
	// Test retrieve
	result := IdTestStruct{}
	if err = db.First(&result, "id = ?", data.ID).Error; err != nil {
		t.Fatalf("Error while selecting: %v", err)
	}
	if data.ID != result.ID || data.Value != result.Value {
		t.Fatalf("Original value did not match with actual value")
	}
	// Test update
	newTestId := NewID()
	if err = db.Model(&data).Where("id = ?", data.ID).Updates(map[string]interface{}{"value": "test-2", "fk_id": newTestId}).Error; err != nil {
		t.Fatalf("Error while updating: %v", err)
	}
	if err = db.First(&result, "id = ?", data.ID).Error; err != nil {
		t.Fatalf("Error while selecting: %v", err)
	}
	if data.ID != result.ID || result.Value != "test-2" || result.FkID != newTestId {
		t.Fatalf("Updated record value did not match")
	}
	// Test delete
	if err = db.Where("id = ?", data.ID).Delete(&data).Error; err != nil {
		t.Fatalf("Error while deleting: %v", err)
	}
	result = IdTestStruct{}
	if err = db.First(&result, "id = ?", data.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Record found even though it should be deleted")
	}
}

func TestIdForSQLite(t *testing.T) {
	type IdTestStruct struct {
		ID    ID     `gorm:"column:id"`
		FkID  ID     `gorm:"column:fk_id"`
		Value string `gorm:"column:value"`
	}

	// Use in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("SQLite Open error: %v", err)
	}

	// Create table
	table := `
		CREATE TABLE IF NOT EXISTS id_test_structs (
			id BLOB NOT NULL,
			value TEXT NOT NULL,
			fk_id BLOB DEFAULT NULL,
			PRIMARY KEY (id)
		);
	`
	if err = db.Exec(table).Error; err != nil {
		t.Fatalf("SQLite table creation error: %v", err)
	}

	data := IdTestStruct{
		ID:    NewID(),
		Value: "test-1",
	}

	// Test create
	if err = db.Create(&data).Error; err != nil {
		t.Fatalf("Error while creating: %v", err)
	}

	// Test retrieve
	result := IdTestStruct{}
	if err = db.First(&result, "id = ?", data.ID).Error; err != nil {
		t.Fatalf("Error while selecting: %v", err)
	}
	if data.ID != result.ID || data.Value != result.Value || result.FkID != NilID {
		t.Fatalf("Original value did not match with actual value")
	}

	// Test update
	newTestId := NewID()
	if err = db.Model(&data).Where("id = ?", data.ID).Updates(map[string]interface{}{"value": "test-2", "fk_id": newTestId}).Error; err != nil {
		t.Fatalf("Error while updating: %v", err)
	}
	if err = db.First(&result, "id = ?", data.ID).Error; err != nil {
		t.Fatalf("Error while selecting: %v", err)
	}
	if data.ID != result.ID || result.Value != "test-2" || result.FkID != newTestId {
		t.Fatalf("Updated record value did not match")
	}

	// Test delete
	if err = db.Where("id = ?", data.ID).Delete(&data).Error; err != nil {
		t.Fatalf("Error while deleting: %v", err)
	}
	result = IdTestStruct{}
	if err = db.First(&result, "id = ?", data.ID).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("Record found even though it should be deleted")
	}
}
