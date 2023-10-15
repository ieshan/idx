package idx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
)

func TestNewID(t *testing.T) {
	id := NewID()
	if id == NilID {
		t.Fatalf("%s is NilId", id.String())
	}
}

func TestCompare(t *testing.T) {
	if NilID.Compare(NilID) != 0 {
		t.Fatalf("NilID should be equal to NilID")
	}
	if NilID.Compare(NotNullNilID) != -1 {
		t.Fatalf("NilID should be less than NotNullNilId")
	}
	if NotNullNilID.Compare(NilID) != 1 {
		t.Fatalf("NotNullNilID should be greater than NilID")
	}
	if NotNullNilID.Compare(NotNullNilID) != 0 {
		t.Fatalf("NotNullNilID should be equal to NotNullNilID")
	}
	if NotNullNilID.Compare([16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}) != -1 {
		t.Fatalf("NotNullNilID should be less")
	}
}

func TestFromString(t *testing.T) {
	id := NewID()
	idFromStr, err := FromString(id.String())
	if err != nil {
		t.Fatalf("Got error while creating ID from String %v", err)
	}
	if id != idFromStr || id.String() != idFromStr.String() {
		t.Fatalf("Original ID (%s) did not match with generated ID (%s)", id.String(), idFromStr.String())
	}
	invalidIds := []string{"null", "wrong", "00000", "01HAJ2Q3T69IJMMBDNAMVZ3FQB"}
	for _, val := range invalidIds {
		_, err = FromString(val)
		if err == nil {
			t.Fatalf("Was expecting error, not there was no error")
		}
	}
}

func TestIsValidID(t *testing.T) {
	id := NewID()
	if !IsValidID(id.String()) {
		t.Fatal("Expecting to be valid, but it's invalid")
	}

	invalidIds := []string{"null", "wrong", "00000", "01HAJ2Q3T69IJMMBDNAMVZ3FQB"}
	for _, val := range invalidIds {
		if IsValidID(val) {
			t.Fatalf("Expecting to be invalid, but it's valid")
		}
	}
}

func TestID_MarshalJSON(t *testing.T) {
	type IdTestStruct struct {
		Id ID `json:"id"`
	}
	id := NewID()
	jsonVal, err := json.Marshal(&IdTestStruct{Id: id})
	if err != nil {
		t.Fatalf("Got error while marshaling to JSON %v", err)
	}
	if string(jsonVal) != fmt.Sprintf(`{"id":"%s"}`, id.String()) {
		t.Fatalf("Original ID (%s) did not match with the ID in generated JSON %s", id.String(), string(jsonVal))
	}
}

func TestID_UnmarshalJSON(t *testing.T) {
	type IdTestStruct struct {
		ID ID `json:"id"`
	}
	id := NewID()
	jsonStrs := []string{
		`{"id":"01HAK8JPF7S0SFMJ2X96W37WXI"}`,
		`{"id":null}`,
		`{"id":""}`,
		fmt.Sprintf(`{"id":"%s"}`, id.String()),
	}
	idVals := []ID{
		NilID,
		NilID,
		NilID,
		id,
	}
	errVals := []error{
		ulid.ErrInvalidCharacters,
		nil,
		nil,
		nil,
	}
	unmVal := IdTestStruct{}
	for index, str := range jsonStrs {
		if err := json.Unmarshal([]byte(str), &unmVal); !errors.Is(err, errVals[index]) {
			t.Fatalf("Error did not match expectation %v : %v", err, errVals[index])
		}
		if unmVal.ID != idVals[index] {
			t.Fatalf("Original ID (%s) did not match with the ID from JSON %s %d", idVals[index].String(), unmVal.ID.String(), index)
		}
	}
}

func TestIdForMongo(t *testing.T) {
	type IdTestStruct struct {
		ID    ID     `bson:"_id"`
		Value string `bson:"value"`
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb://root:password@mongo:27017/?maxPoolSize=5&w=majority").SetServerAPIOptions(serverAPI)
	c := context.TODO()
	client, err := mongo.Connect(c, opts)
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
