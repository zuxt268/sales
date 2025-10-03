package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/zuxt268/sales/internal/domain"
	"github.com/zuxt268/sales/internal/infrastructure"
	"github.com/zuxt268/sales/internal/util"

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

// GORMタグの検証テスト
func TestDomainRepository_GormTags(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	t.Run("PrimaryKey and AutoIncrement", func(t *testing.T) {
		d1 := &domain.Domain{Name: "pk-test1.com"}
		d2 := &domain.Domain{Name: "pk-test2.com"}

		if err := repo.Save(ctx, d1); err != nil {
			t.Fatalf("Failed to save domain 1: %v", err)
		}
		if err := repo.Save(ctx, d2); err != nil {
			t.Fatalf("Failed to save domain 2: %v", err)
		}

		if d1.ID == 0 {
			t.Error("ID should be auto-incremented, got 0")
		}
		if d2.ID <= d1.ID {
			t.Errorf("ID should be auto-incremented: d1.ID=%d, d2.ID=%d", d1.ID, d2.ID)
		}
	})

	t.Run("Unique constraint on Name", func(t *testing.T) {
		d1 := &domain.Domain{Name: "unique-test.com", Title: "First"}
		d2 := &domain.Domain{Name: "unique-test.com", Title: "Second"}

		if err := repo.Save(ctx, d1); err != nil {
			t.Fatalf("Failed to save first domain: %v", err)
		}

		// 同じNameで保存しようとするとエラーになるはず
		err := testDB.Create(d2).Error
		if err == nil {
			t.Error("Expected unique constraint error, got nil")
		}
	})

	t.Run("AutoCreateTime and AutoUpdateTime", func(t *testing.T) {
		d := &domain.Domain{Name: "time-test.com", Title: "Time Test"}

		if err := repo.Save(ctx, d); err != nil {
			t.Fatalf("Failed to save domain: %v", err)
		}

		saved, err := repo.Get(ctx, DomainFilter{Name: &d.Name})
		if err != nil {
			t.Fatalf("Failed to get domain: %v", err)
		}

		// CreateAtが自動設定されているか
		if saved.CreateAt.IsZero() {
			t.Error("CreateAt should be auto-set by autoCreateTime tag")
		}

		// UpdateAtが自動設定されているか
		if saved.UpdateAt.IsZero() {
			t.Error("UpdateAt should be auto-set by autoUpdateTime tag")
		}

		// CreateAtとUpdateAtが同じタイミングで設定される（新規作成時）
		if saved.CreateAt.Unix() != saved.UpdateAt.Unix() {
			t.Errorf("CreateAt and UpdateAt should be the same on creation: CreateAt=%v, UpdateAt=%v", saved.CreateAt, saved.UpdateAt)
		}

		firstCreateAt := saved.CreateAt
		firstUpdateAt := saved.UpdateAt
		time.Sleep(1 * time.Second)

		// 更新時にUpdateAtが自動更新されるか
		saved.Title = "Updated"
		if err := repo.Save(ctx, &saved); err != nil {
			t.Fatalf("Failed to update domain: %v", err)
		}

		updated, err := repo.Get(ctx, DomainFilter{Name: &saved.Name})
		if err != nil {
			t.Fatalf("Failed to get updated domain: %v", err)
		}

		if !updated.UpdateAt.After(firstUpdateAt) {
			t.Errorf("UpdateAt should be updated automatically: first=%v, updated=%v", firstUpdateAt, updated.UpdateAt)
		}
		// CreateAtは変わらないことを確認（秒単位で比較）
		if updated.CreateAt.Unix() != firstCreateAt.Unix() {
			t.Errorf("CreateAt should not change on update: first=%v, updated=%v", firstCreateAt, updated.CreateAt)
		}
	})
}

// 全フィールドの保存・取得テスト
func TestDomainRepository_AllFields(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// 全フィールドを設定
	original := &domain.Domain{
		Name:     "allfields.com",
		CanView:  true,
		IsSend:   true,
		Title:    "All Fields Test",
		OwnerID:  "owner999",
		Address:  "Tokyo, Japan",
		Phone:    "03-9999-9999",
		Industry: "Technology",
		IsSSL:    true,
		RawPage:  "<html><body>Test</body></html>",
		PageNum:  5,
		Status:   domain.StatusDone,
	}

	if err := repo.Save(ctx, original); err != nil {
		t.Fatalf("Failed to save domain: %v", err)
	}

	// 全フィールドが正しく保存・取得できるか検証
	saved, err := repo.Get(ctx, DomainFilter{Name: &original.Name})
	if err != nil {
		t.Fatalf("Failed to get domain: %v", err)
	}

	if saved.ID == 0 {
		t.Error("ID should be set")
	}
	if saved.Name != original.Name {
		t.Errorf("Name: expected %s, got %s", original.Name, saved.Name)
	}
	if saved.CanView != original.CanView {
		t.Errorf("CanView: expected %v, got %v", original.CanView, saved.CanView)
	}
	if saved.IsSend != original.IsSend {
		t.Errorf("IsSend: expected %v, got %v", original.IsSend, saved.IsSend)
	}
	if saved.Title != original.Title {
		t.Errorf("Title: expected %s, got %s", original.Title, saved.Title)
	}
	if saved.OwnerID != original.OwnerID {
		t.Errorf("OwnerID: expected %s, got %s", original.OwnerID, saved.OwnerID)
	}
	if saved.Address != original.Address {
		t.Errorf("Address: expected %s, got %s", original.Address, saved.Address)
	}
	if saved.Phone != original.Phone {
		t.Errorf("Phone: expected %s, got %s", original.Phone, saved.Phone)
	}
	if saved.Industry != original.Industry {
		t.Errorf("Industry: expected %s, got %s", original.Industry, saved.Industry)
	}
	if saved.IsSSL != original.IsSSL {
		t.Errorf("IsSSL: expected %v, got %v", original.IsSSL, saved.IsSSL)
	}
	if saved.RawPage != original.RawPage {
		t.Errorf("RawPage: expected %s, got %s", original.RawPage, saved.RawPage)
	}
	if saved.PageNum != original.PageNum {
		t.Errorf("PageNum: expected %d, got %d", original.PageNum, saved.PageNum)
	}
	if saved.Status != original.Status {
		t.Errorf("Status: expected %s, got %s", original.Status, saved.Status)
	}
	if saved.CreateAt.IsZero() {
		t.Error("CreateAt should be set")
	}
	if saved.UpdateAt.IsZero() {
		t.Error("UpdateAt should be set")
	}
}

// フィルター機能の網羅的テスト
func TestDomainRepository_Filters(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// テストデータの準備
	testDomains := []*domain.Domain{
		{Name: "filter1.com", CanView: true, IsSend: false, OwnerID: "owner1", Industry: "Tech", IsSSL: true, Status: domain.StatusInitialize},
		{Name: "filter2.com", CanView: false, IsSend: true, OwnerID: "owner2", Industry: "Finance", IsSSL: false, Status: domain.StatusDone},
		{Name: "filter3.com", CanView: true, IsSend: true, OwnerID: "owner1", Industry: "Tech", IsSSL: true, Status: domain.StatusPhone},
	}

	for _, d := range testDomains {
		if err := repo.Save(ctx, d); err != nil {
			t.Fatalf("Failed to save test domain: %v", err)
		}
	}

	t.Run("Filter by ID", func(t *testing.T) {
		saved, _ := repo.Get(ctx, DomainFilter{Name: &testDomains[0].Name})
		result, err := repo.Get(ctx, DomainFilter{ID: &saved.ID})
		if err != nil {
			t.Fatalf("Failed to get domain by ID: %v", err)
		}
		if result.ID != saved.ID {
			t.Errorf("Expected ID %d, got %d", saved.ID, result.ID)
		}
	})

	t.Run("Filter by CanView", func(t *testing.T) {
		canView := true
		results, err := repo.FindAll(ctx, DomainFilter{
			PartialName: util.Pointer("filter"),
			CanView:     &canView,
		})
		if err != nil {
			t.Fatalf("Failed to find domains: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 domains with CanView=true, got %d", len(results))
		}
		for _, r := range results {
			if !r.CanView {
				t.Error("Found domain with CanView=false")
			}
		}
	})

	t.Run("Filter by IsSend", func(t *testing.T) {
		isSend := true
		results, err := repo.FindAll(ctx, DomainFilter{
			PartialName: util.Pointer("filter"),
			IsSend:      &isSend,
		})
		if err != nil {
			t.Fatalf("Failed to find domains: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 domains with IsSend=true, got %d", len(results))
		}
	})

	t.Run("Filter by OwnerID", func(t *testing.T) {
		ownerID := "owner1"
		results, err := repo.FindAll(ctx, DomainFilter{
			PartialName: util.Pointer("filter"),
			OwnerID:     &ownerID,
		})
		if err != nil {
			t.Fatalf("Failed to find domains: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 domains with OwnerID=owner1, got %d", len(results))
		}
	})

	t.Run("Filter by Industry", func(t *testing.T) {
		industry := "Tech"
		results, err := repo.FindAll(ctx, DomainFilter{
			PartialName: util.Pointer("filter"),
			Industry:    &industry,
		})
		if err != nil {
			t.Fatalf("Failed to find domains: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 domains with Industry=Tech, got %d", len(results))
		}
	})

	t.Run("Filter by IsSSL", func(t *testing.T) {
		isSSL := true
		results, err := repo.FindAll(ctx, DomainFilter{
			PartialName: util.Pointer("filter"),
			IsSSL:       &isSSL,
		})
		if err != nil {
			t.Fatalf("Failed to find domains: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 domains with IsSSL=true, got %d", len(results))
		}
	})

	t.Run("Filter by Status", func(t *testing.T) {
		status := domain.StatusDone
		results, err := repo.FindAll(ctx, DomainFilter{
			PartialName: util.Pointer("filter"),
			Status:      &status,
		})
		if err != nil {
			t.Fatalf("Failed to find domains: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 domain with Status=done, got %d", len(results))
		}
	})

	t.Run("Multiple filters combined", func(t *testing.T) {
		canView := true
		industry := "Tech"
		results, err := repo.FindAll(ctx, DomainFilter{
			PartialName: util.Pointer("filter"),
			CanView:     &canView,
			Industry:    &industry,
		})
		if err != nil {
			t.Fatalf("Failed to find domains: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 domains, got %d", len(results))
		}
	})
}

// デフォルト値とNULL値のテスト
func TestDomainRepository_DefaultValues(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	// 最小限のフィールドのみ設定
	d := &domain.Domain{
		Name: "minimal.com",
	}

	if err := repo.Save(ctx, d); err != nil {
		t.Fatalf("Failed to save domain: %v", err)
	}

	saved, err := repo.Get(ctx, DomainFilter{Name: &d.Name})
	if err != nil {
		t.Fatalf("Failed to get domain: %v", err)
	}

	// bool型のデフォルト値はfalse
	if saved.CanView {
		t.Error("CanView should default to false")
	}
	if saved.IsSend {
		t.Error("IsSend should default to false")
	}
	if saved.IsSSL {
		t.Error("IsSSL should default to false")
	}

	// string型のデフォルト値は空文字列
	if saved.Title != "" {
		t.Errorf("Title should be empty, got %s", saved.Title)
	}

	// int型のデフォルト値は0
	if saved.PageNum != 0 {
		t.Errorf("PageNum should be 0, got %d", saved.PageNum)
	}

	// タイムスタンプは自動設定される
	if saved.CreateAt.IsZero() {
		t.Error("CreateAt should be set even with minimal fields")
	}
	if saved.UpdateAt.IsZero() {
		t.Error("UpdateAt should be set even with minimal fields")
	}
}

// Exists メソッドのテスト
func TestDomainRepository_Exists(t *testing.T) {
	repo := NewDomainRepository(testDB)
	ctx := context.Background()

	d := &domain.Domain{Name: "exists-test.com"}
	if err := repo.Save(ctx, d); err != nil {
		t.Fatalf("Failed to save domain: %v", err)
	}

	t.Run("Exists returns true for existing domain", func(t *testing.T) {
		exists, err := repo.Exists(ctx, DomainFilter{Name: &d.Name})
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if !exists {
			t.Error("Expected domain to exist")
		}
	})

	t.Run("Exists returns false for non-existing domain", func(t *testing.T) {
		nonExistent := "non-existent-domain.com"
		exists, err := repo.Exists(ctx, DomainFilter{Name: &nonExistent})
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists {
			t.Error("Expected domain not to exist")
		}
	})
}
