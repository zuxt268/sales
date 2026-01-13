package usecase

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/zuxt268/sales/assets"
	"github.com/zuxt268/sales/internal/config"
	"github.com/zuxt268/sales/internal/entity"
	"github.com/zuxt268/sales/internal/interfaces/adapter"
	"github.com/zuxt268/sales/internal/interfaces/dto/request"
	"github.com/zuxt268/sales/internal/interfaces/repository"
	"github.com/zuxt268/sales/internal/model"
	"golang.org/x/exp/slices"
)

type DeployUsecase interface {
	Deploy(ctx context.Context, body request.DeployRequest)
	DeployOne(ctx context.Context, body request.DeployOneRequest) error
	FetchDomains(ctx context.Context) ([]string, error)
	FetchDomainDetails(ctx context.Context) error
}

type deployUsecase struct {
	sshAdapter       adapter.SSHAdapter
	siteSheetAdapter adapter.SheetAdapter
	homstaRepo       repository.HomstaRepository
}

func NewDeployUsecase(
	sshAdapter adapter.SSHAdapter,
	siteSheetAdapter adapter.SheetAdapter,
	homstaRepo repository.HomstaRepository,
) DeployUsecase {
	return &deployUsecase{
		sshAdapter:       sshAdapter,
		siteSheetAdapter: siteSheetAdapter,
		homstaRepo:       homstaRepo,
	}
}

func (u *deployUsecase) FetchDomainDetails(ctx context.Context) error {
	serverIDs := strings.Split(config.Env.ServerIDs, ",")

	type result struct {
		out string
		err error
	}

	ch := make(chan result, len(serverIDs))
	var wg sync.WaitGroup
	for _, serverID := range serverIDs {
		wg.Add(1)

		go func() {
			defer wg.Done()
			sshConf, err := config.GetSSHConfig(serverID)
			if err != nil {
				slog.Error("SSH設定取得失敗", "serverID", serverID, "error", err.Error())
				ch <- result{err: err}
				return
			}

			cmd := "walk fetchDomainDetails"
			stdout, err := u.sshAdapter.RunOutput(sshConf, cmd)
			if err != nil {
				ch <- result{err: err}
				return
			}
			ch <- result{out: stdout}
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	existsPaths := make([]string, 0, 4048)
	var firstErr error

	for r := range ch {
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
			}
			continue
		}

		var partial []entity.DomainDetails
		if err := json.Unmarshal([]byte(r.out), &partial); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		fmt.Println("partial", len(partial))

		for _, d := range partial {
			dbName, dbUsage := getDb(d.DBUsage)
			homsta := &model.Homsta{
				Domain:      getDomain(d.SiteUrl),
				BlogName:    d.BlogName,
				Path:        d.Path,
				SiteURL:     d.SiteUrl,
				Description: d.Description,
				Users:       d.Users,
				DBName:      dbName,
				DBUsage:     dbUsage,
				DiscUsage:   d.DiscUsage,
			}
			existsPaths = append(existsPaths, homsta.Path)
			exists, err := u.homstaRepo.FindAll(ctx, repository.HomstaFilter{
				Path: &d.Path,
			})
			if err != nil {
				return err
			}
			updated := false
			if len(exists) != 0 {
				exist := exists[0]

				if exist.Domain == homsta.Domain &&
					exist.DBName == dbName &&
					exist.DBUsage == dbUsage &&
					exist.Path == d.Path &&
					exist.Description == d.Description &&
					exist.DiscUsage == d.DiscUsage &&
					exist.Users == d.Users &&
					exist.SiteURL == d.SiteUrl &&
					exist.BlogName == d.BlogName {
					fmt.Println("same")
					continue
				}
				updated = true
				homsta.ID = exist.ID
				homsta.Industry = exist.Industry
				homsta.CreatedAt = exist.CreatedAt
			}

			err = u.homstaRepo.Save(ctx, homsta)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if updated {
				fmt.Println("updated", homsta.Path)
			} else {
				fmt.Println("created", homsta.Path)
			}
		}
	}
	if firstErr != nil {
		return firstErr
	}

	fmt.Println("=====")
	fmt.Println(len(existsPaths))
	fmt.Println("=====")

	domains, err := u.homstaRepo.FindAll(ctx, repository.HomstaFilter{})
	if err != nil {
		return err
	}
	for _, d := range domains {
		if !slices.Contains(existsPaths, d.Path) {
			if err := u.homstaRepo.Delete(ctx, repository.HomstaFilter{
				Path: &d.Path,
			}); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

func (u *deployUsecase) FetchDomains(ctx context.Context) ([]string, error) {

	var domains []string
	serverIDs := strings.Split(config.Env.ServerIDs, ",")
	for _, serverID := range serverIDs {
		sshConf, err := config.GetSSHConfig(serverID)
		if err != nil {
			slog.Error("SSH設定取得失敗", "error", err.Error())
			return nil, err
		}

		cmd := "walk fetchDomains"

		stdout, err := u.sshAdapter.RunOutput(sshConf, cmd)
		if err != nil {
			return nil, err
		}

		var result []string
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			return nil, err
		}
		domains = append(domains, result...)
	}
	return domains, nil
}

func (u *deployUsecase) Deploy(ctx context.Context, req request.DeployRequest) {

	slog.Info("デプロイ開始", "src", req.Src)

	if err := os.MkdirAll("./tmp", 0755); err != nil {
		slog.Error("ディレクトリ作成に失敗", "error", err.Error())
		return
	}

	start := time.Now()

	srcConfig, err := config.GetSSHConfig(req.Src.ServerID)
	if err != nil {
		slog.Error("Srcのconfigの取得に失敗", "error", err.Error())
		return
	}

	// サーバーに入り、バックアップを作ります。
	slog.Info("リモートでバックアップ作成開始", "domain", req.Src.Domain)
	err = u.createBackup(req.Src, srcConfig)
	if err != nil {
		slog.Error("バックアップコマンドの失敗", "error", err.Error())
		return
	}
	slog.Info("リモートでバックアップ作成完了", "domain", req.Src.Domain)

	slog.Info("バックアップ取得開始", "domain", req.Src.Domain)

	if err := u.sshAdapter.DownloadFile(ctx,
		srcConfig,
		fmt.Sprintf("%s/%s.sql", req.Src.WordpressRootDirectory(), req.Src.Domain),
		fmt.Sprintf("./tmp/%s.sql", req.Src.Domain),
	); err != nil {
		slog.Error("sqlファイルのダウンロードに失敗", "error", err.Error())
		return
	}

	if err := u.sshAdapter.DownloadFile(ctx,
		srcConfig,
		fmt.Sprintf("%s/%s.zip", req.Src.WordpressRootDirectory(), req.Src.Domain),
		fmt.Sprintf("./tmp/%s.zip", req.Src.Domain),
	); err != nil {
		slog.Error("zipファイルのダウンロードに失敗", "error", err.Error())
		return
	}

	slog.Info("バックアップ取得完了", "domain", req.Src.Domain)

	// ServerIDでdestinationをグルーピング
	serverMap := make(map[string][]entity.Deploy)
	for _, dst := range req.Dst {
		serverMap[dst.ServerID] = append(serverMap[dst.ServerID], dst)
	}
	slog.Info("サーバーグルーピング完了", "server_count", len(serverMap), "destination_count", len(req.Dst))

	// 各サーバーの/tmpへzip/sqlをアップロード（並列）
	slog.Info("各サーバーへ/tmpアップロード開始", "server_count", len(serverMap))
	var uploadWg sync.WaitGroup
	uploadSem := make(chan struct{}, 5) // 最大5並列
	uploadErrors := make(chan error, len(serverMap))

	for serverID := range serverMap {
		serverID := serverID
		uploadWg.Add(1)
		uploadSem <- struct{}{}

		go func() {
			defer uploadWg.Done()
			defer func() { <-uploadSem }()

			slog.Info("/tmpへアップロード開始", "server_id", serverID)
			serverConfig, err := config.GetSSHConfig(serverID)
			if err != nil {
				slog.Error("サーバーconfig取得失敗", "error", err.Error(), "server_id", serverID)
				uploadErrors <- err
				return
			}

			// zip アップロード
			if err := u.sshAdapter.UploadFile(ctx,
				serverConfig,
				fmt.Sprintf("./tmp/%s.zip", req.Src.Domain),
				fmt.Sprintf("/tmp/%s.zip", req.Src.Domain),
			); err != nil {
				slog.Error("/tmpへzipアップロード失敗", "error", err.Error(), "server_id", serverID)
				uploadErrors <- err
				return
			}

			// sql アップロード
			if err := u.sshAdapter.UploadFile(ctx,
				serverConfig,
				fmt.Sprintf("./tmp/%s.sql", req.Src.Domain),
				fmt.Sprintf("/tmp/%s.sql", req.Src.Domain),
			); err != nil {
				slog.Error("/tmpへsqlアップロード失敗", "error", err.Error(), "server_id", serverID)
				uploadErrors <- err
				return
			}

			slog.Info("/tmpへアップロード完了", "server_id", serverID)
		}()
	}

	uploadWg.Wait()
	close(uploadErrors)

	// アップロードエラーチェック
	for err := range uploadErrors {
		slog.Error("アップロードエラー発生", "error", err.Error())
		return
	}
	slog.Info("全サーバーへ/tmpアップロード完了", "server_count", len(serverMap))

	sites := make([]string, len(req.Dst))
	var mu sync.Mutex
	var wg sync.WaitGroup         // 全goroutine待機
	sem := make(chan struct{}, 5) // 並列最大20件

	for _, dst := range req.Dst {
		dst := dst // ループ変数のキャプチャ
		wg.Add(1)
		sem <- struct{}{} // 空きを取得（満杯ならブロック）

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			slog.Info("デプロイ処理開始", "domain", dst.Domain)

			dstConfig, err := config.GetSSHConfig(dst.ServerID)
			if err != nil {
				slog.Error("Dstのconfigの取得に失敗", "error", err.Error(), "domain", dst.Domain)
				return
			}

			// dstディレクトリをクリーンアップ
			slog.Info("dstディレクトリクリーンアップ開始", "domain", dst.Domain)
			cleanupCmd := fmt.Sprintf("cd %s && find . -mindepth 1 -maxdepth 1 -exec rm -rf {} +", dst.WordpressRootDirectory())
			if err := u.sshAdapter.Run(dstConfig, cleanupCmd); err != nil {
				slog.Error("クリーンアップコマンド失敗", "error", err.Error(), "domain", dst.Domain)
				return
			}
			slog.Info("dstディレクトリクリーンアップ完了", "domain", dst.Domain)

			// /tmp からコピー
			slog.Info("/tmpからコピー開始", "domain", dst.Domain)
			copyCmd := fmt.Sprintf("cp /tmp/%s.zip %s/%s.zip && cp /tmp/%s.sql %s/%s.sql",
				req.Src.Domain, dst.WordpressRootDirectory(), req.Src.Domain,
				req.Src.Domain, dst.WordpressRootDirectory(), req.Src.Domain,
			)
			if err := u.sshAdapter.Run(dstConfig, copyCmd); err != nil {
				slog.Error("/tmpからコピー失敗", "error", err.Error(), "domain", dst.Domain)
				return
			}
			slog.Info("/tmpからコピー完了", "domain", dst.Domain)

			slog.Info("展開 & インポート開始", "domain", dst.Domain)
			err = u.restoreBackup(req.Src, dst, dstConfig)
			if err != nil {
				slog.Error("展開&インポート失敗", "error", err.Error(), "domain", dst.Domain)
				return
			}
			slog.Info("展開 & DB復元完了", "domain", dst.Domain)

			slog.Info("Rootの.htaccess書き込み開始", "domain", dst.Domain)
			defaultHtaccess := `# BEGIN WordPress
<IfModule mod_rewrite.c>
RewriteEngine On
RewriteBase /
RewriteRule ^index\.php$ - [L]
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule . /index.php [L]
</IfModule>
# END WordPress
`
			htaccessPath := fmt.Sprintf("%s/.htaccess", dst.WordpressRootDirectory())
			if err := u.sshAdapter.WriteFile(dstConfig, []byte(defaultHtaccess), htaccessPath); err != nil {
				slog.Error("Rootの.htaccess書き込み失敗", "error", err)
				return
			}
			slog.Info("Rootの.htaccess書き込み完了", "domain", dst.Domain)

			// PHPファイルを配布
			slog.Info("PHPファイル配布開始", "domain", dst.Domain)
			if dst.IsSubDomain() {
				content, err := assets.Root.ReadFile("php/mamoru.php")
				if err != nil {
					slog.Error("mamoru.php読み込み失敗", "error", err.Error(), "domain", dst.Domain)
					return
				}

				var remotePath string
				remotePath = fmt.Sprintf("%s/mamoru.php", dst.MuPluginDirectory())

				if err := u.sshAdapter.WriteFile(dstConfig, content, remotePath); err != nil {
					slog.Error("mamoru.php書き込み失敗", "error", err.Error(), "domain", dst.Domain)
					return
				}

				// .hash_dataファイルを作成してMuPluginDirectoryに書き込み
				slog.Info(".hash_dataファイル作成開始", "domain", dst.Domain)

				// .hash_dataファイルをリモートに書き込み (0600パーミッション)
				hashFilePath := fmt.Sprintf("%s/.hash_data", dst.MuPluginDirectory())
				if err := u.sshAdapter.WriteFileWithPerm(dstConfig, []byte(dst.GetHashData()), hashFilePath, "0644"); err != nil {
					slog.Error(".hash_data書き込み失敗", "error", err.Error(), "domain", dst.Domain)
					return
				}
				slog.Info(".hash_dataファイル作成完了", "domain", dst.Domain)
			} else {
				// mamoru.phpと.hash_dataを削除
				slog.Info("mamoru.phpと.hash_dataの削除開始", "domain", dst.Domain)
				cleanupCmd := fmt.Sprintf("rm -f %s/mamoru.php %s/.hash_data",
					dst.MuPluginDirectory(),
					dst.MuPluginDirectory())
				if err := u.sshAdapter.Run(dstConfig, cleanupCmd); err != nil {
					slog.Warn("ファイル削除失敗", "domain", dst.Domain, "error", err)
					return
				}
				slog.Info("mamoru.phpと.hash_dataの削除完了", "domain", dst.Domain)
			}

			slog.Info("rodut配布開始", "domain", dst.Domain)

			err = u.rodut(dst, dstConfig)
			if err != nil {
				return
			}

			slog.Info("rodut配布完了", "domain", dst.Domain)

			slog.Info(".zipと.sqlの削除開始", "domain", dst.Domain)
			cleanupCmd = fmt.Sprintf("rm -f %s/%s.zip %s/%s.sql",
				dst.WordpressRootDirectory(),
				req.Src.Domain,
				dst.WordpressRootDirectory(),
				req.Src.Domain,
			)
			if err := u.sshAdapter.Run(dstConfig, cleanupCmd); err != nil {
				slog.Error(".ファイル削除失敗", "error", err.Error(), "domain", dst.Domain)
				return
			}
			slog.Info(".zipと.sqlの削除完了", "domain", dst.Domain)

			mu.Lock()
			sites = append(sites, dst.Domain)
			mu.Unlock()
		}()
	}

	wg.Wait()

	// 各サーバーの/tmp内のzip/sqlをクリーンアップ（並列）
	slog.Info("各サーバーの/tmpクリーンアップ開始", "server_count", len(serverMap))
	var cleanupWg sync.WaitGroup
	cleanupSem := make(chan struct{}, 5)

	for serverID := range serverMap {
		serverID := serverID
		cleanupWg.Add(1)
		cleanupSem <- struct{}{}

		go func() {
			defer cleanupWg.Done()
			defer func() { <-cleanupSem }()

			serverConfig, err := config.GetSSHConfig(serverID)
			if err != nil {
				slog.Warn("/tmpクリーンアップ用config取得失敗", "error", err.Error(), "server_id", serverID)
				return
			}

			cleanupCmd := fmt.Sprintf("rm -f /tmp/%s.zip /tmp/%s.sql", req.Src.Domain, req.Src.Domain)
			if err := u.sshAdapter.Run(serverConfig, cleanupCmd); err != nil {
				slog.Warn("/tmpクリーンアップ失敗", "error", err.Error(), "server_id", serverID)
			} else {
				slog.Info("/tmpクリーンアップ完了", "server_id", serverID)
			}
		}()
	}

	cleanupWg.Wait()
	slog.Info("全サーバーの/tmpクリーンアップ完了")

	_ = os.Remove(fmt.Sprintf("./tmp/%s.sql", req.Src.Domain))
	_ = os.Remove(fmt.Sprintf("./tmp/%s.zip", req.Src.Domain))

	slog.Info("全デプロイ完了", "duration", time.Since(start).Seconds())

}

func (u *deployUsecase) DeployOne(ctx context.Context, req request.DeployOneRequest) error {
	slog.Info("デプロイ開始", "src", req.Src)
	if err := os.MkdirAll("./tmp", 0755); err != nil {
		return err
	}
	srcConfig, err := config.GetSSHConfig(req.Src.ServerID)
	if err != nil {
		return err
	}
	// サーバーに入り、バックアップを作ります。
	slog.Info("リモートでバックアップ作成開始", "domain", req.Src.Domain)
	err = u.createBackup(req.Src, srcConfig)
	if err != nil {
		return err
	}
	slog.Info("リモートでバックアップ作成完了", "domain", req.Src.Domain)

	slog.Info("バックアップ取得開始", "domain", req.Src.Domain)

	if err := u.sshAdapter.DownloadFile(ctx,
		srcConfig,
		fmt.Sprintf("%s/%s.sql", req.Src.WordpressRootDirectory(), req.Src.Domain),
		fmt.Sprintf("./tmp/%s.sql", req.Src.Domain),
	); err != nil {
		return err
	}

	if err := u.sshAdapter.DownloadFile(ctx,
		srcConfig,
		fmt.Sprintf("%s/%s.zip", req.Src.WordpressRootDirectory(), req.Src.Domain),
		fmt.Sprintf("./tmp/%s.zip", req.Src.Domain),
	); err != nil {
		return err
	}

	slog.Info("バックアップ取得完了", "domain", req.Src.Domain)

	dstConfig, err := config.GetSSHConfig(req.Dst.ServerID)
	if err != nil {
		return err
	}

	// zip アップロード
	if err := u.sshAdapter.UploadFile(ctx,
		dstConfig,
		fmt.Sprintf("./tmp/%s.zip", req.Src.Domain),
		fmt.Sprintf("/tmp/%s.zip", req.Src.Domain),
	); err != nil {
		return err
	}

	// sql アップロード
	if err := u.sshAdapter.UploadFile(ctx,
		dstConfig,
		fmt.Sprintf("./tmp/%s.sql", req.Src.Domain),
		fmt.Sprintf("/tmp/%s.sql", req.Src.Domain),
	); err != nil {
		return err
	}

	dst := req.Dst

	// dstディレクトリをクリーンアップ
	cleanupCmd := fmt.Sprintf("cd %s && find . -mindepth 1 -maxdepth 1 -exec rm -rf {} +", dst.WordpressRootDirectory())
	if err := u.sshAdapter.Run(dstConfig, cleanupCmd); err != nil {
		return err
	}
	slog.Info("dstディレクトリクリーンアップ完了", "domain", dst.Domain)

	// /tmp からコピー
	slog.Info("/tmpからコピー開始", "domain", dst.Domain)
	copyCmd := fmt.Sprintf("cp /tmp/%s.zip %s/%s.zip && cp /tmp/%s.sql %s/%s.sql",
		req.Src.Domain, dst.WordpressRootDirectory(), req.Src.Domain,
		req.Src.Domain, dst.WordpressRootDirectory(), req.Src.Domain,
	)
	if err := u.sshAdapter.Run(dstConfig, copyCmd); err != nil {
		return err
	}
	slog.Info("/tmpからコピー完了", "domain", dst.Domain)

	slog.Info("展開 & インポート開始", "domain", dst.Domain)
	err = u.restoreBackup(req.Src, dst, dstConfig)
	if err != nil {
		return err
	}
	slog.Info("展開 & DB復元完了", "domain", dst.Domain)

	slog.Info("Rootの.htaccess書き込み開始", "domain", dst.Domain)
	defaultHtaccess := `# BEGIN WordPress
<IfModule mod_rewrite.c>
RewriteEngine On
RewriteBase /
RewriteRule ^index\.php$ - [L]
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule . /index.php [L]
</IfModule>
# END WordPress
`
	htaccessPath := fmt.Sprintf("%s/.htaccess", dst.WordpressRootDirectory())
	if err := u.sshAdapter.WriteFile(dstConfig, []byte(defaultHtaccess), htaccessPath); err != nil {
		return err
	}
	slog.Info("Rootの.htaccess書き込み完了", "domain", dst.Domain)

	// PHPファイルを配布
	slog.Info("PHPファイル配布開始", "domain", dst.Domain)
	if dst.IsSubDomain() {
		content, err := assets.Root.ReadFile("php/mamoru.php")
		if err != nil {
			return err
		}

		var remotePath string
		remotePath = fmt.Sprintf("%s/mamoru.php", dst.MuPluginDirectory())

		if err := u.sshAdapter.WriteFile(dstConfig, content, remotePath); err != nil {
			return err
		}

		// .hash_dataファイルを作成してMuPluginDirectoryに書き込み
		slog.Info(".hash_dataファイル作成開始", "domain", dst.Domain)

		// .hash_dataファイルをリモートに書き込み (0644パーミッション)
		hashFilePath := fmt.Sprintf("%s/.hash_data", dst.MuPluginDirectory())
		if err := u.sshAdapter.WriteFileWithPerm(dstConfig, []byte(dst.GetHashData()), hashFilePath, "0644"); err != nil {
			return err
		}
		slog.Info(".hash_dataファイル作成完了", "domain", dst.Domain)
	} else {
		if err := u.sshAdapter.Run(dstConfig, fmt.Sprintf("rm -f %s/mamoru.php %s/.hash_data",
			dst.MuPluginDirectory(),
			dst.MuPluginDirectory())); err != nil {
			return err
		}
		slog.Info("mamoru.phpと.hash_dataの削除完了", "domain", dst.Domain)
	}

	slog.Info("rodut配布開始", "domain", dst.Domain)

	err = u.rodut(dst, dstConfig)
	if err != nil {
		return err
	}

	slog.Info("rodut配布完了", "domain", dst.Domain)

	slog.Info(".zipと.sqlの削除開始", "domain", dst.Domain)
	if err := u.sshAdapter.Run(dstConfig, fmt.Sprintf("rm -f %s/%s.zip %s/%s.sql",
		dst.WordpressRootDirectory(),
		req.Src.Domain,
		dst.WordpressRootDirectory(),
		req.Src.Domain,
	)); err != nil {
		slog.Error(".ファイル削除失敗", "error", err.Error(), "domain", dst.Domain)
		return err
	}
	slog.Info(".zipと.sqlの削除完了", "domain", dst.Domain)

	_ = u.sshAdapter.Run(
		dstConfig,
		fmt.Sprintf("rm -f /tmp/%s.zip /tmp/%s.sql", req.Src.Domain, req.Src.Domain),
	)
	_ = os.Remove(fmt.Sprintf("./tmp/%s.sql", req.Src.Domain))
	_ = os.Remove(fmt.Sprintf("./tmp/%s.zip", req.Src.Domain))

	return nil
}

func (u *deployUsecase) createBackup(src entity.Deploy, srcConfig config.SSHConfig) error {
	backUpCmd := fmt.Sprintf(`
cd %s &&
bash -lc '
set -euo pipefail

rm -f "%[2]s.sql" "%[2]s.zip"

wp db export "%[2]s.sql"

zip -rq "%[2]s.zip" . \
  -x "*/.git/*" \
  -x "wp-content/cache/*" \
  -x "wp-content/uploads/backups/*" \
  -x "wp-content/upgrade/*" \
  -x "wp-content/ai1wm-backups/*" \
  -x ".htpasswd" \
  -x ".htaccess" \
  -x "*.zip" \
  -x "*.tar.gz" \
  -x "*.log" \
  -x "*.sql"
'
`, src.WordpressRootDirectory(), src.Domain)
	if err := u.sshAdapter.Run(srcConfig, backUpCmd); err != nil {
		return err
	}
	return nil
}

func (u *deployUsecase) restoreBackup(src entity.Deploy, dst entity.Deploy, dstConfig config.SSHConfig) error {
	deployCmd := fmt.Sprintf(`
cd %s &&
unzip -oq %s.zip &&

# DB 接続先を差し替え
sed -i "s/define( 'DB_NAME'.*/define( 'DB_NAME', '%s' );/" wp-config.php &&
sed -i "s/define( 'DB_USER'.*/define( 'DB_USER', '%s' );/" wp-config.php &&
sed -i "s/define( 'DB_PASSWORD'.*/define( 'DB_PASSWORD', '%s' );/" wp-config.php &&
sed -i "s/define( 'DB_HOST'.*/define( 'DB_HOST', '%s' );/" wp-config.php &&

# WP_HOME / WP_SITEURL が wp-config.php に固定されていると option update が効かないので除去
php -r '
$f="wp-config.php";
$s=file_get_contents($f);
$s=preg_replace("/^.*define\\(\\s*\\x27WP_HOME\\x27.*\\R?/m","",$s);
$s=preg_replace("/^.*define\\(\\s*\\x27WP_SITEURL\\x27.*\\R?/m","",$s);
file_put_contents($f,$s);
' &&

# DB インポート
php8.2 ~/wp-cli.phar db import %s.sql &&

# 置換（全テーブル対象）
php8.2 ~/wp-cli.phar search-replace 'https://%s' 'https://%s' --skip-columns=guid --all-tables-with-prefix &&
php8.2 ~/wp-cli.phar search-replace 'http://%s' 'http://%s' --skip-columns=guid --all-tables-with-prefix &&
php8.2 ~/wp-cli.phar search-replace '%s' '%s' --skip-columns=guid --all-tables-with-prefix &&

# サイトURL確定
php8.2 ~/wp-cli.phar option update home 'https://%s' &&
php8.2 ~/wp-cli.phar option update siteurl 'https://%s' &&

# キャッシュ類を掃除（古いURLが残りやすい）
php8.2 ~/wp-cli.phar cache flush || true &&
php8.2 ~/wp-cli.phar transient delete --all || true &&
php8.2 ~/wp-cli.phar rewrite flush --hard || true
`,
		dst.WordpressRootDirectory(),
		src.Domain,
		dst.GetDbName(),
		dst.GetDbUser(),
		dst.GetDbPassword(),
		dst.GetDbHost(),
		src.Domain,
		src.Domain, dst.Domain,
		src.Domain, dst.Domain,
		src.Domain, dst.Domain,
		dst.Domain,
		dst.Domain,
	)

	if err := u.sshAdapter.Run(dstConfig, deployCmd); err != nil {
		return err
	}
	return nil
}

func (u *deployUsecase) rodut(dst entity.Deploy, dstConfig config.SSHConfig) error {
	rodutPhp, err := assets.Root.ReadFile("php/rodut.php")
	if err != nil {
		return errors.Wrap(err, "rodut.php読み込み失敗")
	}
	rodutPhpPath := fmt.Sprintf("%s/rodut.php", dst.MuPluginDirectory())
	if err := u.sshAdapter.WriteFile(dstConfig, rodutPhp, rodutPhpPath); err != nil {
		return errors.Wrap(err, "rodut.php書き込み失敗")
	}

	rodutCss, err := assets.Root.ReadFile("php/rodut.css")
	if err != nil {
		return errors.Wrap(err, "rodut.css読み込み失敗")
	}
	rodutCssPath := fmt.Sprintf("%s/rodut-style.css", dst.MuPluginDirectory())
	if err := u.sshAdapter.WriteFile(dstConfig, rodutCss, rodutCssPath); err != nil {
		return errors.Wrap(err, "rodut.css書き込み失敗")
	}

	// ApiKeyの生成
	apiKeyInput := fmt.Sprintf("%s%s", config.Env.RodutSecretPhrase, dst.Domain)
	apiKeyHash := sha256.Sum256([]byte(apiKeyInput))
	apiKey := fmt.Sprintf("%x", apiKeyHash)

	// テンプレート読み込み
	secretConfigTemplate, err := assets.Root.ReadFile("php/secret-config.php")
	if err != nil {
		return errors.Wrap(err, "secret-config.php読み込み失敗")
	}

	// テンプレートをパース
	tmpl, err := template.New("secret-config").Parse(string(secretConfigTemplate))
	if err != nil {
		return errors.Wrap(err, "secret-config.phpテンプレートパース失敗")
	}

	// テンプレートにデータを埋め込み
	var buf bytes.Buffer
	data := map[string]string{
		"ApiKey":   apiKey,
		"SlackUrl": config.Env.NoticeWebAppChannelUrl,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return errors.Wrap(err, "secret-config.phpテンプレート実行失敗")
	}

	secretConfigPath := fmt.Sprintf("%s/wp-content/secret-config.php", dst.WordpressRootDirectory())
	if err := u.sshAdapter.WriteFileWithPerm(dstConfig, buf.Bytes(), secretConfigPath, "0600"); err != nil {
		return errors.Wrap(err, "secret-config.php書き込み失敗")
	}

	htaccess, err := assets.Root.ReadFile("php/.htaccess")
	if err != nil {
		return errors.Wrap(err, ".htaccess読み込み失敗")
	}

	htaccessPath := fmt.Sprintf("%s/wp-content/.htaccess", dst.WordpressRootDirectory())
	if err := u.sshAdapter.WriteFile(dstConfig, htaccess, htaccessPath); err != nil {
		return errors.Wrap(err, ".htaccess書き込み失敗")
	}
	return nil
}
