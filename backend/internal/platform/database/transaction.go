package database

import (
	"context"

	"gorm.io/gorm"
)

type transactionContextKey struct{}

type TransactionRunner interface {
	WithinTransaction(ctx context.Context, work func(ctx context.Context) error) error
}

type GormTransactionRunner struct {
	database *gorm.DB
}

func NewTransactionRunner(database *gorm.DB) *GormTransactionRunner {
	return &GormTransactionRunner{database: database}
}

func (runner *GormTransactionRunner) WithinTransaction(ctx context.Context, work func(ctx context.Context) error) error {
	if _, isInsideTransaction := ctx.Value(transactionContextKey{}).(*gorm.DB); isInsideTransaction {
		return work(ctx)
	}
	return runner.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		return work(context.WithValue(ctx, transactionContextKey{}, transaction))
	})
}

func Session(ctx context.Context, database *gorm.DB) *gorm.DB {
	if transaction, isInsideTransaction := ctx.Value(transactionContextKey{}).(*gorm.DB); isInsideTransaction {
		return transaction.WithContext(ctx)
	}
	return database.WithContext(ctx)
}
