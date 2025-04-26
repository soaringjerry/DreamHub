package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5" // Correct import path for pgx.Tx
	"github.com/jackc/pgx/v5/pgxpool"

	// "github.com/pgvector/pgvector-go" // Remove unused import again
	"github.com/soaringjerry/dreamhub/pkg/apperr"  // 引入 apperr 包
	"github.com/soaringjerry/dreamhub/pkg/config"  // 引入 config 包
	"github.com/soaringjerry/dreamhub/pkg/ctxutil" // 引入 ctxutil 包
	"github.com/soaringjerry/dreamhub/pkg/logger"  // 引入 logger 包
)

// DB représente le pool de connexions à la base de données PostgreSQL.
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB crée et retourne un nouveau pool de connexions PostgreSQL.
// Il tente de se connecter à la base de données en utilisant l'URL fournie dans la configuration.
// Il effectue également un ping pour vérifier la connexion.
func NewDB(ctx context.Context, cfg *config.Config) (*DB, error) {
	// Utiliser SanitizeConnectionString pour logger l'URL sans le mot de passe
	sanitizedURL := SanitizeConnectionString(cfg.DatabaseURL)
	logger.InfoContext(ctx, "初始化 PostgreSQL 连接池...", "url", sanitizedURL)

	// Configuration du pool de connexions
	// Voir https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#ParseConfig pour plus d'options
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法解析数据库连接字符串")
	}

	// Définir des limites de connexion raisonnables
	poolConfig.MaxConns = 10 // Ajuster en fonction de la charge attendue
	poolConfig.MinConns = 2
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// --- Remove AfterConnect hook again ---
	// poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
	// 	pgvector.Register(conn.TypeMap()) // This function seems unavailable or problematic
	// 	return nil
	// }
	// --- End of removal ---

	// Créer le pool de connexions
	// Use the standard NewWithConfig without the AfterConnect hook for now
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig) // Revert to original call if AfterConnect was the only change
	if err != nil {
		return nil, apperr.Wrap(err, apperr.CodeUnavailable, "无法创建数据库连接池") // Use CodeUnavailable
	}

	// Vérifier la connexion avec un Ping
	// Utiliser un contexte avec timeout pour le ping initial
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()                                                               // Fermer le pool si le ping échoue
		return nil, apperr.Wrap(err, apperr.CodeUnavailable, "无法连接到数据库 (ping 失败)") // Use CodeUnavailable
	}

	// No registration needed here

	logger.InfoContext(ctx, "PostgreSQL 连接池初始化成功。") // Revert log message
	return &DB{Pool: pool}, nil
}

// Close ferme le pool de connexions à la base de données.
func (db *DB) Close() {
	if db.Pool != nil {
		logger.Info("正在关闭 PostgreSQL 连接池...")
		db.Pool.Close()
		logger.Info("PostgreSQL 连接池已关闭。")
	}
}

// BeginTx démarre une nouvelle transaction.
// Utilise pgx.TxOptions pour définir le niveau d'isolation si nécessaire.
func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) { // Return pgx.Tx
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		// Return nil explicitly for the interface type on error
		return nil, apperr.Wrap(err, apperr.CodeInternal, "无法开始数据库事务")
	}
	return tx, nil
}

// CommitTx valide une transaction.
func CommitTx(ctx context.Context, tx pgx.Tx) error { // Accept pgx.Tx
	err := tx.Commit(ctx)
	if err != nil {
		return apperr.Wrap(err, apperr.CodeInternal, "无法提交数据库事务")
	}
	return nil
}

// RollbackTx annule une transaction.
func RollbackTx(ctx context.Context, tx pgx.Tx) error { // Accept pgx.Tx
	err := tx.Rollback(ctx)
	if err != nil {
		// Log l'erreur de rollback mais ne la propage pas nécessairement comme une erreur critique
		// car l'erreur originale qui a causé le rollback est souvent plus importante.
		logger.ErrorContext(ctx, "数据库事务回滚失败", "error", err)
		// Retourner nil ou une erreur spécifique au rollback si nécessaire
		return apperr.Wrap(err, apperr.CodeInternal, "数据库事务回滚失败")
	}
	return nil
}

// ExecuteInTx exécute une fonction à l'intérieur d'une transaction gérée.
// La transaction est automatiquement validée si la fonction réussit,
// ou annulée si la fonction retourne une erreur.
func (db *DB) ExecuteInTx(ctx context.Context, fn func(tx pgx.Tx) error) error { // Accept func(tx pgx.Tx)
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err // L'erreur est déjà enveloppée par BeginTx
	}
	defer func() {
		// Utiliser un nouveau contexte pour le rollback au cas où le contexte original serait annulé
		rollbackCtx := context.Background()
		if p := recover(); p != nil {
			_ = RollbackTx(rollbackCtx, tx)
			panic(p) // Relancer la panique après le rollback
		} else if err != nil {
			// Si fn a retourné une erreur, annuler la transaction
			if rollbackErr := RollbackTx(rollbackCtx, tx); rollbackErr != nil {
				// Log l'erreur de rollback mais retourner l'erreur originale de fn
				logger.ErrorContext(rollbackCtx, "事务回滚失败（在 ExecuteInTx 中）", "rollback_error", rollbackErr, "original_error", err)
			}
		} else {
			// Si fn a réussi, valider la transaction
			err = CommitTx(ctx, tx) // Utiliser le contexte original pour le commit
		}
	}()

	err = fn(tx) // Exécuter la fonction fournie avec la transaction
	return err   // Retourner l'erreur de fn (ou nil si succès)
}

// GetUserIDFromCtx est une fonction utilitaire pour extraire user_id du contexte.
// Elle retourne une erreur si user_id est manquant, pour renforcer la politique d'isolation.
func GetUserIDFromCtx(ctx context.Context) (string, error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		// Ceci est une erreur critique car cela signifie que la logique d'isolation des données pourrait être contournée.
		logger.ErrorContext(ctx, "无法从上下文中获取 user_id，数据隔离可能被破坏！")
		return "", apperr.New(apperr.CodeInternal, "无法从上下文中获取 user_id")
	}
	return userID, nil
}

// SanitizeConnectionString masque le mot de passe dans une URL de base de données pour le logging.
// Attention : ceci est une implémentation basique et peut ne pas couvrir tous les formats d'URL.
func SanitizeConnectionString(dbURL string) string {
	// Implémentation simple pour les formats courants postgres://user:password@host:port/db
	// Trouve la position de '://'
	schemaEnd := -1
	for i := 0; i+2 < len(dbURL); i++ {
		if dbURL[i:i+3] == "://" {
			schemaEnd = i + 3
			break
		}
	}
	if schemaEnd == -1 {
		return "invalid_db_url_format" // Format non reconnu
	}

	atIndex := -1
	colonIndex := -1
	// Cherche ':' et '@' après '://'
	for i := schemaEnd; i < len(dbURL); i++ {
		if dbURL[i] == ':' && colonIndex == -1 {
			// Trouve le premier ':' après '://', potentiellement avant le mot de passe
			colonIndex = i
		} else if dbURL[i] == '@' {
			atIndex = i
			break // Trouvé le '@', on peut arrêter
		}
	}

	// Vérifie si on a trouvé user:password@
	if colonIndex != -1 && atIndex != -1 && colonIndex < atIndex {
		// Masquer le mot de passe entre ':' et '@'
		return dbURL[:colonIndex+1] + "***" + dbURL[atIndex:]
	}

	// Si pas de format user:password@, retourne l'URL telle quelle (ou une version partiellement masquée si nécessaire)
	// Par exemple, si c'est juste postgres://host:port/db
	return dbURL
}
