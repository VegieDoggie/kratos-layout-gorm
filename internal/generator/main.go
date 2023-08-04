package main

import (
	"flag"

	"github.com/go-kratos/kratos-layout/internal/conf"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

var (
	flagconf string
)

//go:generate go run main.go
func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(bc.Data.Database.GetSource()))
	if err != nil {
		panic(err)
	}

	// stat, err := os.Stat("../data/db/mysql")
	// if err != nil || stat == nil {

	// }
	generate(db)
}

func generate(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath:          "../data/db/mysql",
		ModelPkgPath:     "../biz/model",
		Mode:             gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
		FieldNullable:    false,
		FieldWithTypeTag: true,
	})
	g.UseDB(db)

	// 兼容protobuf,把所有的数字转为int64
	dataMap := map[string]func(detailType gorm.ColumnType) (dataType string){
		"tinyint":   func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"smallint":  func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"mediumint": func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"bigint":    func(detailType gorm.ColumnType) (dataType string) { return "int64" },
		"int":       func(detailType gorm.ColumnType) (dataType string) { return "int64" },
	}

	g.WithDataTypeMap(dataMap)

	// Generate basic type-safe DAO API for struct `model.User` following conventions
	bc := g.GenerateModel("s_table")
	customer := g.GenerateModel("s_function_table", gen.FieldRelate(field.HasOne, "STableRef", bc,
		&field.RelateConfig{
			// RelateSlice: true,
			GORMTag: field.GormTag{}.Append("foreignKey", "s_table_id"),
		}),
	)

	g.ApplyBasic(
		// Generate struct `User` based on table `users`

		// Generate struct `User` based on table `users` and generating options
		bc, customer,
	)
	// Generate the code
	g.Execute()
}

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}
