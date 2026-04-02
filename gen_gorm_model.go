// Copyright 2026 onexstack. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/onexstack/realworld/common"
	"github.com/spf13/pflag"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

const helpText = `Usage: gen_model [flags]

Generate GORM models from database tables.

Flags:
`

var (
	addr      = pflag.StringP("addr", "a", "localhost:3306", "MySQL host address.")
	username  = pflag.StringP("username", "u", "realworld", "Username to connect to the database.")
	password  = pflag.StringP("password", "p", "", "Password to use when connecting to the database.")
	database  = pflag.StringP("db", "d", "realworld", "Database name to connect to.")
	modelPath = pflag.String("model-path", "apiserver/model", "Generated model code path.")
	help      = pflag.BoolP("help", "h", false, "Show this help message.")
)

func main() {
	pflag.Usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults()
	}
	pflag.Parse()

	if *help {
		pflag.Usage()
		return
	}

	db, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	generator := createGenerator(resolveModelPackagePath("apiserver/model"))
	generator.UseDB(db)
	applyGeneratorOptions(generator)
	generateRealWorldModels(generator)
	generator.Execute()

	log.Println("Model generation completed successfully")
}

func initializeDatabase() (*gorm.DB, error) {
	return common.NewMySQL(&common.MySQLOptions{
		Addr:     *addr,
		Username: *username,
		Password: *password,
		Database: *database,
	})
}

func resolveModelPackagePath(defaultPath string) string {
	if *modelPath != "" {
		return *modelPath
	}

	absPath, err := filepath.Abs(defaultPath)
	if err != nil {
		log.Printf("Error resolving path: %v", err)
		return defaultPath
	}

	return absPath
}

func createGenerator(packagePath string) *gen.Generator {
	return gen.NewGenerator(gen.Config{
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface | gen.WithoutContext,
		ModelPkgPath:      packagePath,
		WithUnitTest:      true,
		FieldNullable:     true,
		FieldSignable:     false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	})
}

func applyGeneratorOptions(g *gen.Generator) {
	g.WithOpts(
		gen.FieldGORMTag("createdAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "CURRENT_TIMESTAMP")
			return tag
		}),
		gen.FieldGORMTag("updatedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("default", "CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP")
			return tag
		}),
	)
}

func generateRealWorldModels(g *gen.Generator) {
	g.GenerateModelAs(
		"user_models",
		"UserM",
		gen.FieldGORMTag("username", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_user_models_username")
			return tag
		}),
		gen.FieldGORMTag("email", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_user_models_email")
			return tag
		}),
		gen.FieldGORMTag("deletedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_user_models_deleted_at")
			return tag
		}),
	)

	g.GenerateModelAs(
		"follow_models",
		"FollowM",
		gen.FieldGORMTag("followingId", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_follow_models_following_id")
			return tag
		}),
		gen.FieldGORMTag("followedById", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_follow_models_followed_by_id")
			return tag
		}),
		gen.FieldGORMTag("deletedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_follow_models_deleted_at")
			return tag
		}),
	)

	g.GenerateModelAs(
		"article_models",
		"ArticleM",
		gen.FieldGORMTag("slug", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_article_models_slug")
			return tag
		}),
		gen.FieldGORMTag("authorId", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_article_models_author_id")
			return tag
		}),
		gen.FieldGORMTag("deletedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_article_models_deleted_at")
			return tag
		}),
	)

	g.GenerateModelAs(
		"tag_models",
		"TagM",
		gen.FieldGORMTag("tag", func(tag field.GormTag) field.GormTag {
			tag.Set("uniqueIndex", "idx_tag_models_tag")
			return tag
		}),
		gen.FieldGORMTag("deletedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_tag_models_deleted_at")
			return tag
		}),
	)

	g.GenerateModelAs("article_tags", "ArticleTagM")

	g.GenerateModelAs(
		"favorite_models",
		"FavoriteM",
		gen.FieldGORMTag("favoriteId", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_favorite_models_favorite_id")
			return tag
		}),
		gen.FieldGORMTag("favoriteById", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_favorite_models_favorite_by_id")
			return tag
		}),
		gen.FieldGORMTag("deletedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_favorite_models_deleted_at")
			return tag
		}),
	)

	g.GenerateModelAs(
		"comment_models",
		"CommentM",
		gen.FieldGORMTag("articleId", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_comment_models_article_id")
			return tag
		}),
		gen.FieldGORMTag("authorId", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_comment_models_author_id")
			return tag
		}),
		gen.FieldGORMTag("deletedAt", func(tag field.GormTag) field.GormTag {
			tag.Set("index", "idx_comment_models_deleted_at")
			return tag
		}),
	)
}
