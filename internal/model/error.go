package model

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
