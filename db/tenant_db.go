// File: common/prism-common-libs/db/tenant_db.go
package db

import (
	"context"
	"fmt"

	// Impor auth package Anda, pastikan path-nya benar
	"github.com/Lumina-Enterprise-Solutions/prism-common-libs/auth"
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
// Ini adalah metode utama yang harus digunakan oleh service layer.
func (d *TenantDB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	// 1. Ekstrak tenantID dari context yang datang dari middleware JWT.
	tenantID, err := auth.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi tenant-aware: %w", err)
	}

	// 2. Mulai transaksi database biasa.
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Atur variabel sesi 'app.tenant_id' HANYA untuk transaksi ini.
	// Menggunakan `SET LOCAL` memastikan pengaturan ini hanya berlaku untuk transaksi saat ini.
	_, err = tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return nil, fmt.Errorf("gagal rollback setelah error RLS context: %w (original error: %v)", rbErr, err)
		}
		return nil, fmt.Errorf("gagal mengatur RLS context: %w", err)
	}

	// 4. Kembalikan transaksi yang sekarang sudah "tenant-aware".
	return tx, nil
}

// GetPool mengembalikan pool koneksi mentah. Hati-hati menggunakannya,
// karena tidak akan secara otomatis mengatur RLS.
// Berguna untuk operasi yang tidak spesifik-tenant.
func (d *TenantDB) GetPool() *pgxpool.Pool {
	return d.pool
}
