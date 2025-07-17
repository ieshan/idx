module integration-tests

go 1.24

replace github.com/ieshan/idx => ../

require (
	github.com/ieshan/idx v0.0.0
	go.mongodb.org/mongo-driver/v2 v2.2.2
	gorm.io/driver/mysql v1.6.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.30.0
)
