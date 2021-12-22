module gorm.io/gorm/tests

go 1.13

require (
	github.com/codenotary/immudb v1.2.1
	github.com/codenotary/immugorm v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.3.0
	github.com/jackc/pgx/v4 v4.14.1 // indirect
	github.com/jinzhu/now v1.1.4
	github.com/lib/pq v1.10.4
	google.golang.org/grpc v1.40.0
	gorm.io/driver/mysql v1.2.1
	gorm.io/driver/postgres v1.2.3
	gorm.io/driver/sqlite v1.2.6
	gorm.io/driver/sqlserver v1.2.1
	gorm.io/gorm v1.22.4
)

replace gorm.io/gorm => ../

replace github.com/codenotary/immudb => ../../immudb/src

replace github.com/codenotary/immugorm => ../../immugorm

replace github.com/spf13/afero => github.com/spf13/afero v1.5.1
