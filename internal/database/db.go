package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Driver SQLite pur Go (sans dépendance CGO)
)

// DB encapsule la connexion à la base SQLite de ZeroDrop.
type DB struct {
	*sql.DB
}

// New initialise la base de données SQLite avec les pragmas de haute performance.
// Le mode WAL (Write-Ahead Logging) permet des lectures et écritures concurrentes sans blocage.
func New(dbPath string) (*DB, error) {
	// Création du répertoire parent si nécessaire
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("impossible de créer le répertoire de la base de données : %w", err)
		}
	}

	// Paramètres de connexion SQLite pour la robustesse et la rapidité
	// _pragma=busy_timeout(5000) évite les erreurs "database is locked"
	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", dbPath)

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de SQLite : %w", err)
	}

	// Configuration du pool de connexions
	// SQLite gère parfaitement plusieurs connexions en lecture avec WAL
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	db := &DB{DB: sqlDB}

	// Exécution automatique des migrations initiales
	if err := db.Migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("échec des migrations de la base : %w", err)
	}

	return db, nil
}
