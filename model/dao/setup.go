// Package dao contains all data access logic for models.
//
//nolint:unused
package dao

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/dingdayu/go-project-template/pkg/logger"
	pkgOtel "github.com/dingdayu/go-project-template/pkg/otel"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

var (
	once sync.Once
	db   *gorm.DB

	tracer = otel.Tracer("github.com/dingdayu/go-project-template/model/dao")
)

// Setup setup db connection.
func Setup() {
	var err error
	once.Do(func() {
		dsn := viper.GetString("db")
		dbCfg := &gorm.Config{
			Logger: logger.NewGormLogger(logger.WithNamespace("gorm")),
		}

		var dialector gorm.Dialector

		// try parse dsn as URL to detect scheme first
		if u, perr := url.Parse(dsn); perr == nil && u.Scheme != "" {
			fmt.Printf("\033[1;30;42m[info]\033[0m db dsn scheme detected: %s\n", u.Scheme)
			switch strings.ToLower(u.Scheme) {
			case "postgres", "postgresql":
				dialector = postgres.Open(dsn)
			case "mysql":
				dialector = mysql.Open(dsn)
			case "sqlite", "sqlite3":
				normalized := normalizeSQLiteDSN(dsn)
				fmt.Printf("\033[1;30;42m[info]\033[0m db sqlite normalized dsn: %s\n", normalized)
				dialector = sqlite.Open(normalized)
			case "file":
				// DSN already in sqlite URI form: file:xxx?params
				dialector = sqlite.Open(dsn)
			}
		}

		// fallback to simple prefix/content checks when URL parsing didn't help
		if dialector == nil {
			ld := strings.ToLower(dsn)
			switch {
			case strings.HasPrefix(ld, "postgres"), strings.HasPrefix(ld, "postgresql"), strings.Contains(ld, "host="):
				dialector = postgres.Open(dsn)
			case strings.HasPrefix(ld, "mysql"), strings.Contains(ld, "@tcp("), strings.Contains(ld, "charset="):
				dialector = mysql.Open(dsn)
			case strings.HasSuffix(ld, ".db"), strings.HasSuffix(ld, ".sqlite"), strings.HasPrefix(ld, "file:"):
				dialector = sqlite.Open(dsn)
			}
		}

		if dialector == nil {
			fmt.Printf("\033[1;30;41m[error]\033[0m db [master] connect error: unsupported dsn or driver for '%s'\n", dsn)
			os.Exit(1)
		}

		db, err = gorm.Open(dialector, dbCfg)
		if err == nil {
			var databaseName string
			// use a special session to avoid printing SQL statements
			pkgOtel.NeverLogSessionWithGORM(db).Raw("SELECT current_database()").Scan(&databaseName)
			fmt.Printf("\033[1;30;42m[info]\033[0m db [master:%s] connect success\n", databaseName)

			// enable gorm OpenTelemetry tracing
			if err := db.Use(tracing.NewPlugin()); err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Printf("\033[1;30;41m[error]\033[0m db [master] connect error: %s", err.Error())
			os.Exit(1)
		}
	})
}

// GetDB gorm default DB connection.
func GetDB() *gorm.DB {
	return db
}

// GetContextDB returns a gorm DB with the provided context.
func GetContextDB(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}

// normalizeSQLiteDSN converts sqlite URL-style DSNs into forms accepted by gorm's sqlite driver.
// Examples:
//
//	sqlite://app.db           -> app.db
//	sqlite:///abs/app.db      -> /abs/app.db
//	sqlite://data/app.db      -> data/app.db
//	sqlite://app.db?cache=on  -> file:app.db?cache=on
//	sqlite:/tmp/app.db        -> /tmp/app.db (via Opaque)
//	file:uri?params           -> file:uri?params (unchanged)
func normalizeSQLiteDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil || u.Scheme == "" {
		return dsn
	}
	// Pass-through file: URIs
	if strings.EqualFold(u.Scheme, "file") {
		return dsn
	}
	// Only handle sqlite schemes here
	if !strings.EqualFold(u.Scheme, "sqlite") && !strings.EqualFold(u.Scheme, "sqlite3") {
		return dsn
	}

	var path string
	if u.Opaque != "" {
		// e.g., sqlite:/tmp/app.db or sqlite:relative.db
		path = u.Opaque
	} else if u.Host != "" && (u.Path == "" || u.Path == "/") {
		// sqlite://app.db
		path = u.Host
	} else if u.Path != "" {
		// sqlite:///abs/app.db -> /abs/app.db
		// sqlite://data/app.db -> data/app.db
		if u.Host != "" && u.Path != "/" {
			path = u.Host + "/" + strings.TrimPrefix(u.Path, "/")
		} else {
			path = u.Path
		}
	} else {
		// Fallback: strip scheme prefixes
		path = strings.TrimPrefix(dsn, u.Scheme+"://")
		path = strings.TrimPrefix(path, u.Scheme+":")
	}

	// Preserve query parameters by using SQLite URI mode when present
	if u.RawQuery != "" {
		if strings.Contains(path, "?") {
			return path + "&" + u.RawQuery
		}
		return "file:" + path + "?" + u.RawQuery
	}
	return path
}
