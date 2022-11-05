package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zacjones91/pantry-api/internal/database"
	"github.com/zacjones91/pantry-api/internal/leveledlog"
	"github.com/zacjones91/pantry-api/internal/server"
	"github.com/zacjones91/pantry-api/internal/version"
)

type config struct {
	addr    string
	baseURL string
	env     string
	auth    struct {
		username       string
		hashedPassword string
	}
	db struct {
		dsn         string
		automigrate bool
	}
	jwt struct {
		secretKey string
	}
	tls struct {
		certFile string
		keyFile  string
	}
	version bool
}

type application struct {
	config config
	db     *database.DB
	logger *leveledlog.Logger
}

func main() {
	var cfg config

	flag.StringVar(&cfg.addr, "addr", "localhost:4444", "server address to listen on")
	flag.StringVar(&cfg.baseURL, "base-url", "https://localhost:4444", "base URL for the application")
	flag.StringVar(&cfg.env, "env", "development", "operating environment: development, testing, staging or production")
	flag.StringVar(&cfg.auth.username, "auth-username", "admin", "basic auth username")
	flag.StringVar(&cfg.auth.hashedPassword, "auth-hashed-password", "$2a$10$jRb2qniNcoCyQM23T59RfeEQUbgdAXfR6S0scynmKfJa5Gj3arGJa", "basic auth password hashed with bcrpyt")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "user:pass@localhost:5432/db", "postgreSQL DSN")
	flag.BoolVar(&cfg.db.automigrate, "db-automigrate", true, "run migrations on startup")
	flag.StringVar(&cfg.jwt.secretKey, "jwt-secret-key", "wb7j4z637voznspxymcjjfqng4oxpuln", "secret key for JWT authentication")
	flag.StringVar(&cfg.tls.certFile, "tls-cert-file", "./tls/cert.pem", "tls certificate file")
	flag.StringVar(&cfg.tls.keyFile, "tls-key-file", "./tls/key.pem", "tls key file")
	flag.BoolVar(&cfg.version, "version", false, "display version and exit")

	flag.Parse()

	if cfg.version {
		fmt.Printf("version: %s\n", version.Get())
		return
	}

	logger := leveledlog.NewLogger(os.Stdout, leveledlog.LevelAll, true)

	db, err := database.New(cfg.db.dsn, cfg.db.automigrate)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
	}

	logger.Info("starting server on %s (version %s)", cfg.addr, version.Get())

	err = server.Run(cfg.addr, app.routes(), cfg.tls.certFile, cfg.tls.keyFile)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Info("server stopped")
}
