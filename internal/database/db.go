// Package database assure la gestion de la persistance relationnelle de ZeroDrop.
// Il utilise le moteur SQLite pur Go (modernc.org/sqlite) configuré en mode WAL (Write-Ahead Logging)
// pour garantir un débit d'écriture élevé et des lectures concurrentes non bloquantes.
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Enregistrement du pilote de base de données pur Go
)

// DB encapsule la connexion au pool SQLite et expose l'ensemble des méthodes d'accès aux données.
type DB struct {
	*sql.DB
}

// New instancie une nouvelle connexion à la base SQLite en appliquant les pragmas de performance nécessaires.
//
// Pragmas appliqués :
//   - journal_mode(WAL) : Active le Write-Ahead Logging pour autoriser les lectures simultanées pendant les écritures.
//   - synchronous(NORMAL) : Réduit les appels fsync bloquants tout en maintenant l'intégrité en mode WAL.
//   - busy_timeout(5000)  : Définit un délai d'attente de 5000ms avant de retourner une erreur de verrouillage.
//   - foreign_keys(ON)    : Active la vérification stricte des contraintes d'intégrité référentielle et suppressions en cascade.
//
// Paramètres :
//   - dbPath : Chemin d'accès au fichier de base de données SQLite (ex: "zerodrop.db" ou "/data/zerodrop.db").
func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("echec de la creation du repertoire de base de donnees (%s) : %w", dir, err)
		}
	}

	dsn := fmt.Sprintf("%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", dbPath)

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("impossible d'ouvrir la connexion SQLite : %w", err)
	}

	// Configuration du pool de connexions
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	db := &DB{DB: sqlDB}

	// Application idempotente des schémas et migrations
	if err := db.Migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("echec lors de l'application des migrations initiales : %w", err)
	}

	return db, nil
}
