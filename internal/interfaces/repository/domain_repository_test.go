package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/infrastructure"

	"gorm.io/gorm"
)

var testDB *gorm.DB
var cleanup func()

func setup() {
	testDB, cleanup = infrastructure.NewTestContainerDBClient()
}

func teardown() {
	if cleanup != nil {
		cleanup()
	}
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestDomainRepository_Save(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	d := &domain.Domain{
		Name:     "example.com",
		CanView:  true,
		IsSend:   false,
		Title:    "Example Domain",
		OwnerID:  "owner123",
		Address:  "Tokyo",
		Phone:    "03-1234-5678",
		Industry: "IT",
		IsSSL:    true,
		RawPage:  "<html></html>",
		PageNum:  1,
		Status:   domain.StatusInitialize,
	}

	err := repo.Save(ctx, d)
	if err != nil {
		t.Fatalf("Failed to save domain: %v", err)
	}

	// 検証: 取得できるか
	saved, err := repo.Get(ctx, DomainFilter{Name: &d.Name})
	if err != nil {
		t.Fatalf("Failed to get saved domain: %v", err)
	}

	if saved.Name != d.Name {
		t.Errorf("Expected name %s, got %s", d.Name, saved.Name)
	}
	if saved.Title != d.Title {
		t.Errorf("Expected title %s, got %s", d.Title, saved.Title)
	}
	if saved.CreateAt.IsZero() {
		t.Error("CreateAt should be set automatically")
	}
	if saved.UpdateAt.IsZero() {
		t.Error("UpdateAt should be set automatically")
	}
}

func TestDomainRepository_Get(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// テストデータを準備
	d := &domain.Domain{
		Name:    "test-get.com",
		Title:   "Test Get",
		OwnerID: "owner456",
	}
	err := repo.Save(ctx, d)
	if err != nil {
		t.Fatalf("Failed to save test domain: %v", err)
	}

	// テスト: 名前で取得
	result, err := repo.Get(ctx, DomainFilter{Name: &d.Name})
	if err != nil {
		t.Fatalf("Failed to get domain: %v", err)
	}

	if result.Name != d.Name {
		t.Errorf("Expected name %s, got %s", d.Name, result.Name)
	}
	if result.Title != d.Title {
		t.Errorf("Expected title %s, got %s", d.Title, result.Title)
	}

	// テスト: 存在しないドメイン
	nonExistent := "non-existent.com"
	_, err = repo.Get(ctx, DomainFilter{Name: &nonExistent})
	if err == nil {
		t.Error("Expected error for non-existent domain, got nil")
	}
}

func TestDomainRepository_FindAll(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// テストデータを準備
	domains := []*domain.Domain{
		{Name: "findall1.com", Title: "FindAll 1", Industry: "Tech"},
		{Name: "findall2.com", Title: "FindAll 2", Industry: "Finance"},
		{Name: "findall3.com", Title: "FindAll 3", Industry: "Tech"},
	}

	for _, d := range domains {
		err := repo.Save(ctx, d)
		if err != nil {
			t.Fatalf("Failed to save test domain: %v", err)
		}
	}

	// テスト: 全件取得
	results, err := repo.FindAll(ctx, DomainFilter{})
	if err != nil {
		t.Fatalf("Failed to find all domains: %v", err)
	}

	if len(results) < len(domains) {
		t.Errorf("Expected at least %d domains, got %d", len(domains), len(results))
	}

	// テスト: 部分一致検索
	partial := "findall"
	results, err = repo.FindAll(ctx, DomainFilter{PartialName: &partial})
	if err != nil {
		t.Fatalf("Failed to find domains with partial name: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 domains with partial name 'findall', got %d", len(results))
	}

	// テスト: Limit/Offset
	limit := 2
	offset := 1
	results, err = repo.FindAll(ctx, DomainFilter{
		PartialName: &partial,
		Limit:       &limit,
		Offset:      &offset,
	})
	if err != nil {
		t.Fatalf("Failed to find domains with limit/offset: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 domains with limit=2, got %d", len(results))
	}
}

func TestDomainRepository_BulkInsert(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// テストデータを準備
	domains := []*domain.Domain{
		{Name: "bulk1.com", Title: "Bulk 1"},
		{Name: "bulk2.com", Title: "Bulk 2"},
		{Name: "bulk3.com", Title: "Bulk 3"},
	}

	err := repo.BulkInsert(ctx, domains)
	if err != nil {
		t.Fatalf("Failed to bulk insert domains: %v", err)
	}

	// 検証: 全て挿入されたか
	for _, d := range domains {
		saved, err := repo.Get(ctx, DomainFilter{Name: &d.Name})
		if err != nil {
			t.Errorf("Failed to get bulk inserted domain %s: %v", d.Name, err)
		}
		if saved.Name != d.Name {
			t.Errorf("Expected name %s, got %s", d.Name, saved.Name)
		}
	}

	// テスト: 重複挿入（DoNothingで無視される）
	err = repo.BulkInsert(ctx, domains)
	if err != nil {
		t.Fatalf("Failed to bulk insert duplicate domains: %v", err)
	}
}

func TestDomainRepository_Delete(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// テストデータを準備
	d := &domain.Domain{
		Name:  "delete-test.com",
		Title: "Delete Test",
	}
	err := repo.Save(ctx, d)
	if err != nil {
		t.Fatalf("Failed to save test domain: %v", err)
	}

	// テスト: 削除
	err = repo.Delete(ctx, DomainFilter{Name: &d.Name})
	if err != nil {
		t.Fatalf("Failed to delete domain: %v", err)
	}

	// 検証: 削除されたか
	_, err = repo.Get(ctx, DomainFilter{Name: &d.Name})
	if err == nil {
		t.Error("Expected error after deletion, got nil")
	}
}

func TestDomainRepository_Update(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// テストデータを準備
	d := &domain.Domain{
		Name:  "update-test.com",
		Title: "Original Title",
	}
	err := repo.Save(ctx, d)
	if err != nil {
		t.Fatalf("Failed to save test domain: %v", err)
	}

	// 最初のUpdateAt取得
	firstSave, _ := repo.Get(ctx, DomainFilter{Name: &d.Name})
	firstUpdateAt := firstSave.UpdateAt

	// 少し待機
	time.Sleep(1 * time.Second)

	// テスト: 更新
	d.Title = "Updated Title"
	d.Phone = "090-1234-5678"
	err = repo.Save(ctx, d)
	if err != nil {
		t.Fatalf("Failed to update domain: %v", err)
	}

	// 検証: 更新されたか
	updated, err := repo.Get(ctx, DomainFilter{Name: &d.Name})
	if err != nil {
		t.Fatalf("Failed to get updated domain: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", updated.Title)
	}
	if updated.Phone != "090-1234-5678" {
		t.Errorf("Expected phone '090-1234-5678', got %s", updated.Phone)
	}

	// UpdateAtが更新されているか確認
	if !updated.UpdateAt.After(firstUpdateAt) {
		t.Errorf("UpdateAt should be updated automatically. First: %v, Updated: %v", firstUpdateAt, updated.UpdateAt)
	}

	// CreateAtは変わらないことを確認
	if !updated.CreateAt.Equal(firstSave.CreateAt) {
		t.Errorf("CreateAt should not change. First: %v, Updated: %v", firstSave.CreateAt, updated.CreateAt)
	}
}

func TestDomainRepository_GetForUpdate(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// テストデータを準備
	d := &domain.Domain{
		Name:  "forupdate-test.com",
		Title: "For Update Test",
	}
	err := repo.Save(ctx, d)
	if err != nil {
		t.Fatalf("Failed to save test domain: %v", err)
	}

	// テスト: 悲観的ロックで取得
	result, err := repo.GetForUpdate(ctx, DomainFilter{Name: &d.Name})
	if err != nil {
		t.Fatalf("Failed to get domain for update: %v", err)
	}

	if result.Name != d.Name {
		t.Errorf("Expected name %s, got %s", d.Name, result.Name)
	}
	if result.Title != d.Title {
		t.Errorf("Expected title %s, got %s", d.Title, result.Title)
	}
}
