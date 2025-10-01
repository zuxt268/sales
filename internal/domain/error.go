package domain

import (
	"errors"
	"fmt"
)

// リソース関連エラー
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrConflict      = errors.New("conflict")
)

// バリデーション関連エラー
var (
	ErrValidation   = errors.New("validation error")
	ErrInvalidInput = errors.New("invalid input")
)

// 外部API関連エラー
var (
	ErrExternalAPI = errors.New("external API error")
	ErrTimeout     = errors.New("timeout")
)

// データベース関連エラー
var (
	ErrDatabase    = errors.New("database error")
	ErrTransaction = errors.New("transaction error")
	ErrDeadlock    = errors.New("deadlock")
)

// 権限関連エラー
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// その他
var (
	ErrInternal = errors.New("internal error")
	ErrUnknown  = errors.New("unknown error")
)

// エラーラップヘルパー関数

func WrapNotFound(resource string) error {
	return fmt.Errorf("%s: %w", resource, ErrNotFound)
}

func WrapAlreadyExists(resource string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w: %v", resource, ErrAlreadyExists, err)
	}
	return fmt.Errorf("%s: %w", resource, ErrAlreadyExists)
}

func WrapValidation(message string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w: %v", message, ErrValidation, err)
	}
	return fmt.Errorf("%s: %w", message, ErrValidation)
}

func WrapExternalAPI(service string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w: %v", service, ErrExternalAPI, err)
	}
	return fmt.Errorf("%s: %w", service, ErrExternalAPI)
}

func WrapDatabase(message string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w: %v", message, ErrDatabase, err)
	}
	return fmt.Errorf("%s: %w", message, ErrDatabase)
}

func WrapTransaction(message string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w: %v", message, ErrTransaction, err)
	}
	return fmt.Errorf("%s: %w", message, ErrTransaction)
}

func WrapUnauthorized(message string) error {
	return fmt.Errorf("%s: %w", message, ErrUnauthorized)
}
