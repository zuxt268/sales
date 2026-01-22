<?php
/**
 * Plugin Name: mamoru
 * Description: 仮ドメイントークン認証
 * Version: 0.9.1
 * Author URI: https://github.com/zuxt268
 */

// 直接アクセスを防ぐ
if (!defined('ABSPATH')) {
    exit;
}

/**
 * トークンの有効性を検証する
 *
 * @param string $token 検証するトークン
 * @return bool トークンが有効な場合true
 */
function verify_temp_domain_access(string $token): bool
{
    $pluginDir = dirname(__FILE__);
    $filePath = $pluginDir . '/.hash_data';

    if (!file_exists($filePath)) {
        return false;
    }

    $hash_data = file_get_contents($filePath);
    $token_hash = hash('sha256', $token);

    return hash_equals($hash_data, $token_hash);
}

/**
 * 親ドメインを取得する
 *
 * @return string 親ドメイン（例: example.com）
 */
function get_parent_domain(): string
{
    $current_domain = $_SERVER['HTTP_HOST'] ?? 'localhost';
    $host_parts = explode('.', $current_domain);

    return implode('.', array_slice($host_parts, -2));
}

/**
 * トークンをCookieに保存する
 *
 * @param string $token 保存するトークン
 * @return void
 */
function set_temp_domain_token(string $token): void
{
    $lifetime = 31536000; // 1年分の秒数
    $secure = !empty($_SERVER['HTTPS']) && $_SERVER['HTTPS'] !== 'off';

    setcookie(
        'temp_domain_token',
        hash('sha256', $token),
        time() + $lifetime,
        '/',
        '',
        $secure,
        true // HttpOnly
    );
}

/**
 * トークンをCookieから取得する
 *
 * @return string|null Cookieに保存されたトークンハッシュ
 */
function get_temp_domain_token(): ?string
{
    return $_COOKIE['temp_domain_token'] ?? null;
}

/**
 * トークンをCookieから削除する
 *
 * @return void
 */
function unset_temp_domain_token(): void
{
    setcookie('temp_domain_token', '', time() - 3600, '/');
}

/**
 * トークン認証を全ページに適用する
 *
 * @return void
 */
function token_auth_check_all_pages(): void
{
    // REST APIリクエストの場合はスキップ
    if (defined('REST_REQUEST') && REST_REQUEST) {
        return;
    }

    // AJAXリクエストの場合もスキップ
    if (defined('DOING_AJAX') && DOING_AJAX) {
        return;
    }

    // ログイン済みユーザーはトークン認証をスキップ
    if (is_user_logged_in()) {
        return;
    }

    // Cookieにトークンがある場合、それを使って認証
    $cookie_token_hash = get_temp_domain_token();
    if ($cookie_token_hash) {
        $pluginDir = dirname(__FILE__);
        $filePath = $pluginDir . '/.hash_data';

        if (file_exists($filePath)) {
            $expected_hash = file_get_contents($filePath);
            if (hash_equals($expected_hash, $cookie_token_hash)) {
                return; // 認証成功
            }
        }

        // 無効なトークンの場合はCookieを削除
        unset_temp_domain_token();
    }

    // URLパラメータからトークンを取得して認証
    if (isset($_GET['token'])) {
        $token = trim($_GET['token']);

        if (!verify_temp_domain_access($token)) {
            wp_die('無効なトークンです。アクセスが拒否されました。');
        }

        // 認証成功時にトークンをCookieに保存
        set_temp_domain_token($token);
        return;
    }

    // トークンが見つからない場合
    wp_die('トークンが見つかりません。');
}
/**
 * 条件に応じてアクションを登録する
 *
 * @return void
 */
function conditional_add_actions(): void
{
    $temp_domains = [
        'hp-standard.net',
        'hp-standard.com',
        'hp-standard.info',
        'sv511.com',
        'sv533.com',
        'hp-standard.xyz'
    ];

    // ドメインが条件に一致する場合のみトークン認証を有効化
    if (in_array(get_parent_domain(), $temp_domains)) {
        add_action('template_redirect', 'token_auth_check_all_pages');
    }
}

// プラグイン初期化
add_action('plugins_loaded', 'conditional_add_actions');
