// File: common/prism-common-libs/db/tenant_db.go
package db

import (
	"context"
	"fmt"
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/auth"
	"github.com/google/uuid" // <-- Tambahkan import ini
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX adalah interface generik untuk pgxpool.Pool dan pgx.Tx
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// TenantDB adalah wrapper di sekitar pgxpool yang mengelola logika RLS.
type TenantDB struct {
	pool *pgxpool.Pool
}

func NewTenantDB(pool *pgxpool.Pool) *TenantDB {
	return &TenantDB{pool: pool}
}

// BeginTx memulai transaksi dan secara otomatis mengatur session variable untuk RLS.
func (d *TenantDB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	// 1. Ekstrak tenantID dari context yang datang dari middleware JWT.
	tenantIDStr, err := auth.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi tenant-aware: %w", err)
	}

	// 2. [KRITIKAL] Validasi bahwa tenantID adalah UUID yang valid.
	// Ini adalah langkah keamanan untuk mencegah SQL Injection.
	// Jika parsing gagal, berarti string tersebut bukan UUID dan request harus ditolak.
	_, err = uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID format provided in context: %w", err)
	}

	// 3. Mulai transaksi database biasa.
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// 4. Bangun query SET LOCAL menggunakan Sprintf.
	// Ini aman karena kita sudah memvalidasi tenantIDStr sebagai UUID.
	// Perhatikan tanda kutip tunggal di sekitar '%s'.
	setRLSQuery := fmt.Sprintf("SET LOCAL app.tenant_id = '%s'", tenantIDStr)

	// 5. Jalankan query yang sudah diformat TANPA placeholder.
	_, err = tx.Exec(ctx, setRLSQuery)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return nil, fmt.Errorf("gagal rollback setelah error RLS context: %w (original error: %v)", rbErr, err)
		}
		return nil, fmt.Errorf("gagal mengatur RLS context: %w", err)
	}

	// 6. Kembalikan transaksi yang sekarang sudah "tenant-aware".
	return tx, nil
}

// GetPool mengembalikan pool koneksi mentah. Hati-hati menggunakannya,
// karena tidak akan secara otomatis mengatur RLS.
func (d *TenantDB) GetPool() *pgxpool.Pool {
    return d.pool
}
