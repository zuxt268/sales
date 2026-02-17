<?php
/*
  Plugin Name: rodut
  Description: ホムスタプラグイン。
  Version: 1.10.1
  Author: Yuki Ikezawa
  Author URI: https://github.com/IkezawaYuki/IkezawaYuki
*/

require_once(ABSPATH . 'wp-admin/includes/file.php');
require_once(ABSPATH . 'wp-admin/includes/image.php');
require_once(ABSPATH . 'wp-admin/includes/media.php');

function verify_hmac_signature($request): bool
{
    $api_key = get_config("api_key");
    if (empty($api_key)) {
        error_log("api_key not configured");
        return false;
    }
    $signature = get_request_header('X-Signature');
    $timestamp = get_request_header('X-Timestamp');
    if (empty($signature) || empty($timestamp)) return false;
    if (abs(time() - intval($timestamp)) > 300) return false; // ±5分
    // Content-Type で分岐
    $content_type = $_SERVER['CONTENT_TYPE'] ?? $_SERVER['HTTP_CONTENT_TYPE'] ?? '';
    if (stripos($content_type, 'multipart/form-data') !== false) {
        // --- upload-media 等: multipart ---
        // 署名対象は「timestamp.email.filename」
        $email = isset($_POST['email']) ? (string)$_POST['email'] : '';
        $filename = isset($_FILES['file']['name']) ? (string)$_FILES['file']['name'] : '';
        // 必須チェック（必要に応じて）
        if ($email === '' || $filename === '') return false;
        $data = $timestamp . '.' . $email . '.' . $filename;
    } else {
        // --- JSON エンドポイント ---
        $rawBody = $request->get_body();
        $data = $timestamp . '.' . $rawBody;
    }
    $expected_signature = hash_hmac('sha256', $data, $api_key);
    return hash_equals($expected_signature, $signature);
}

function get_request_header($name) {
    $nameUpper = strtoupper(str_replace('-', '_', $name));
    // 1. $_SERVER 経由
    if (isset($_SERVER['HTTP_' . $nameUpper])) {
        return $_SERVER['HTTP_' . $nameUpper];
    }
    // 2. getallheaders() がある場合
    if (function_exists('getallheaders')) {
        $headers = getallheaders();
        foreach ($headers as $k => $v) {
            if (strcasecmp($k, $name) === 0) {
                return $v;
            }
        }
    }
    // 3. apache_request_headers() がある場合
    if (function_exists('apache_request_headers')) {
        $headers = apache_request_headers();
        foreach ($headers as $k => $v) {
            if (strcasecmp($k, $name) === 0) {
                return $v;
            }
        }
    }
    return '';
}

/**
 * -------------------------
 * Utility: media URL 判定
 * -------------------------
 */
function mc_is_media_url($url) {
    if (!$url) return false;
    $path = strtolower(parse_url($url, PHP_URL_PATH) ?? '');
    if ($path === '') return false;
    $exts = [
            'jpg','jpeg','png','gif','webp','svg','avif',
            'mp4','webm','mov','m4v','ogg','ogv'
    ];
    foreach ($exts as $ext) {
        if (str_ends_with($path, '.' . $ext)) return true;
    }
    return false;
}

function mc_add_url(&$set, $url) {
    if (!$url) return;
    $url = trim(html_entity_decode($url, ENT_QUOTES));
    if ($url === '') return;
    if (str_starts_with($url, '//')) {
        $url = 'https:' . $url;
    } elseif (str_starts_with($url, '/')) {
        $url = home_url($url);
    }
    // 自ドメインのみ許可（←追加）
    $host = parse_url($url, PHP_URL_HOST);
    if ($host !== parse_url(home_url(), PHP_URL_HOST)) {
        return;
    }
    $set[$url] = true;
}

/**
 * -------------------------
 * HTML → プレーンテキスト変換
 * -------------------------
 */
function mc_html_to_text($html) {
    if (!$html) return '';
    $html = apply_filters('the_content', $html);
    // videoタグ全体削除（←追加）
    $html = preg_replace('/<video\b[^>]*>.*?<\/video>/is', '', $html);
    // script/styleも削除（安全）
    $html = preg_replace('/<(script|style)\b[^>]*>.*?<\/\1>/is', '', $html);
    // <br> → 改行
    $html = preg_replace('/<br\s*\/?>/i', "\n", $html);
    // </p> → 2改行
    $html = preg_replace('/<\/p>/i', "\n\n", $html);
    $text = wp_strip_all_tags($html);
    $text = html_entity_decode($text, ENT_QUOTES | ENT_HTML5, 'UTF-8');
    $text = preg_replace("/\n{3,}/", "\n\n", $text);
    return trim($text);
}

/**
 * -------------------------
 * HTMLからメディア抽出
 * -------------------------
 */
function mc_extract_media_urls_from_html($html) {
    $urls_set = [];
    if (!$html) return [];
    if (preg_match_all('/wp-(?:image|video)-([0-9]+)/', $html, $m)) {
        foreach ($m[1] as $id) {
            $url = wp_get_attachment_url((int)$id);
            mc_add_url($urls_set, $url);
        }
    }
    if (preg_match_all('/<(img|video|source)[^>]+src=["\']([^"\']+)["\']/i', $html, $m)) {
        foreach ($m[2] as $url) mc_add_url($urls_set, $url);
    }
    if (preg_match_all('/<a[^>]+href=["\']([^"\']+)["\']/i', $html, $m)) {
        foreach ($m[1] as $url) {
            if (mc_is_media_url($url)) mc_add_url($urls_set, $url);
        }
    }
    return array_keys($urls_set);
}

/**
 * -------------------------
 * Gutenbergブロック抽出
 * -------------------------
 */
function mc_extract_media_urls_from_blocks($blocks) {
    $urls_set = [];
    if (!is_array($blocks)) return [];
    $walk = function($blk) use (&$walk, &$urls_set) {
        if (!is_array($blk)) return;
        $name = $blk['blockName'] ?? '';
        if ($name === 'core/image') {
            $id = (int)($blk['attrs']['id'] ?? 0);
            if ($id) mc_add_url($urls_set, wp_get_attachment_url($id));
            $url = $blk['attrs']['url'] ?? null;
            if ($url) mc_add_url($urls_set, $url);
        }
        if ($name === 'core/video') {
            $id = (int)($blk['attrs']['id'] ?? 0);
            if ($id) mc_add_url($urls_set, wp_get_attachment_url($id));
            $src = $blk['attrs']['src'] ?? null;
            if ($src) mc_add_url($urls_set, $src);
        }
        if ($name === 'core/embed') {
            $url = $blk['attrs']['url'] ?? null;
            if ($url) mc_add_url($urls_set, $url);
        }
        $innerHTML = $blk['innerHTML'] ?? '';
        if ($innerHTML) {
            foreach (mc_extract_media_urls_from_html($innerHTML) as $u) {
                mc_add_url($urls_set, $u);
            }
        }
        if (!empty($blk['innerBlocks'])) {
            foreach ($blk['innerBlocks'] as $child) $walk($child);
        }
    };
    foreach ($blocks as $b) $walk($b);
    return array_keys($urls_set);
}

/**
 * -------------------------
 * 最新N件取得
 * -------------------------
 */
function mc_get_latest_posts_with_media($limit = 30) {
    $posts = get_posts([
            'post_type'   => 'post',
            'post_status' => 'publish',
            'numberposts' => (int)$limit,
            'orderby'     => 'date',
            'order'       => 'DESC',
    ]);
    $results = [];
    foreach ($posts as $post) {
        $media_set = [];
        $blocks = parse_blocks($post->post_content);
        foreach (mc_extract_media_urls_from_blocks($blocks) as $u) {
            mc_add_url($media_set, $u);
        }
        foreach (mc_extract_media_urls_from_html($post->post_content) as $u) {
            mc_add_url($media_set, $u);
        }
        $thumb_id = get_post_thumbnail_id($post->ID);
        if ($thumb_id) {
            mc_add_url($media_set, wp_get_attachment_url($thumb_id));
        }
        $content = mc_html_to_text($post->post_content);
        $results[] = [
                'post_id'   => (int)$post->ID,   // ← 追加
                'content'   => $content,
                'media_urls'=> array_values(array_keys($media_set)),
        ];
    }
    return $results;
}

function get_post_list($request) {
    $limit = (int)($request->get_param('limit') ?? 30);
    if ($limit <= 0) $limit = 30;
    if ($limit > 100) $limit = 100; // safety
    return mc_get_latest_posts_with_media($limit);
}



function get_config($key) {
    // secret-config.phpのパス
    $config_path = WP_CONTENT_DIR . '/secret-config.php';

    // ファイルが存在しない場合はエラーをログに記録
    if (!file_exists($config_path)) {
        error_log("Secret config file not found: $config_path");
        return null;
    }

    // secret-config.phpを読み込む
    $config = include $config_path;

    // 指定されたキーが存在すれば返す、なければnull
    return $config[$key] ?? null;
}

function get_version(WP_REST_Request $request){
    return new WP_REST_Response(array(
            'version' => "1.9.2"
    ), 200);
}

function rodut_permission_check(WP_REST_Request $request) {
    $ip = get_client_ip($request);

    // === HMAC認証 ===
    if (!verify_hmac_signature($request)) {
        return new WP_Error('forbidden', 'Invalid signature', ['status' => 401]);
    }

    return true;
}

// POSTリクエストに対応するコールバック関数（記事投稿）
function create_post(WP_REST_Request $request) {
    $params = $request->get_json_params();

    if (!isset($params['email'])) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $email = sanitize_email($params['email']);
    $user = get_user_by('email', $email);
    if (!$user) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    // 記事のデータを取得
    $title = sanitize_text_field($params['title'] ?? '');
    $content = $params['content'];
    $media_id = isset($params['featured_media']) ? intval($params['featured_media']) : 0;

    if (empty($content)) {
        return new WP_REST_Response(array('error' => 'Content are required'), 400);
    }

    $post_date = "";
    if (isset($params['post_date'])) {
        $post_date = $params['post_date'];
    }

    // カテゴリーの処理
    $post_data = array(
            'post_title'   => $title,
            'post_content' => $content,
            'post_status'  => 'publish',
            'post_author'  => $user->ID,
            'post_date'    => $post_date,
    );

    if (isset($params['post_category']) && is_array($params['post_category']) && !empty($params['post_category'])) {
        $category_ids = array();
        foreach ($params['post_category'] as $category_name) {
            $category = get_term_by('name', $category_name, 'category');
            if ($category) {
                $category_ids[] = $category->term_id;
            }
        }
        if (!empty($category_ids)) {
            $post_data['post_category'] = $category_ids;
        }
    }

    // 記事を投稿する
    $post_id = wp_insert_post($post_data);

    if (!empty($media_id)) {
        set_post_thumbnail($post_id, $media_id);
    }

    // 記事のURLを取得
    $post_url = get_permalink($post_id);

    if (is_wp_error($post_id)) {
        return new WP_REST_Response(array('error' => 'Failed to create post'), 500);
    }

    return new WP_REST_Response(array(
            'message' => 'Post created successfully',
            'post_id' => $post_id,
            'post_url' => $post_url), 200);
}

function download_media(WP_REST_Request $request)
{
    $params = $request->get_json_params();

    if (!isset($params['email'])) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $email = sanitize_email($params['email']);
    $user = get_user_by('email', $email);
    if (!$user) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    // メディアライブラリ内のすべての添付ファイルを取得
    $args = array(
            'post_type'      => 'attachment',
            'post_status'    => 'inherit', // 添付ファイルは 'inherit' ステータスです
            'posts_per_page' => -1, // 全ファイルを取得、特定の数に制限する場合は数字を指定
    );

    $query = new WP_Query($args);

    // メディア情報をリスト化
    $image_urls = array();

    if ($query->have_posts()) {
        while ($query->have_posts()) {
            $query->the_post();
            // 添付ファイルのURLを取得
            $image_title = get_the_title();
            $image_url = wp_get_attachment_url(get_the_ID());
            $ext = pathinfo($image_url, PATHINFO_EXTENSION);
            $allowed_extensions = array('jpg', 'jpeg', 'png');
            if (in_array(strtolower($ext), $allowed_extensions)) {
                $image_urls[] = array(
                        'url' => $image_url,
                        'title' => $image_title,
                );
            }
        }
    }

    return new WP_REST_Response($image_urls, 200);
}

function upload_media(WP_REST_Request $request)
{
    $email = sanitize_email($_POST['email'] ?? '');

    if (empty($email)) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $user = get_user_by('email', $email);
    if (!$user) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    if (!isset($_FILES['file'])) {
        return new WP_REST_Response(['error' => 'No file uploaded'], 400);
    }

    // ファイル情報の取得とアップロード処理
    $file = $_FILES['file'];

    if ($_FILES['file']['size'] > 1073741824) { // 1GBのファイルサイズ制限
        return new WP_REST_Response(['error' => 'File size exceeds limit'], 400);
    }

    $allowed_types = ['image/jpeg', 'image/png', 'video/mp4'];
    if (!in_array($file['type'], $allowed_types)) {
        return new WP_REST_Response(['error' => 'Unsupported file type'], 400);
    }

    $finfo = finfo_open(FILEINFO_MIME_TYPE);
    $mime_type = finfo_file($finfo, $file['tmp_name']);
    finfo_close($finfo);

    if (!in_array($mime_type, $allowed_types)) {
        return new WP_REST_Response(['error' => 'Unsupported file type'], 400);
    }

    $uploaded_file = wp_handle_upload($file, ['test_form' => false]);
    if (isset($uploaded_file['error'])) {
        return new WP_REST_Response(['error' => $uploaded_file['error']], 500);
    }

    $filename = $uploaded_file['file'];
    $filetype = wp_check_filetype($filename, null);

    $attachment = array(
            'post_mime_type' => $filetype['type'],
            'post_title'     => sanitize_file_name(basename($filename)),
            'post_content'   => '',
            'post_status'    => 'inherit'
    );

    // メディアライブラリに添付ファイルとして登録
    $attach_id = wp_insert_attachment($attachment, $filename);

    // 添付ファイルのメタデータを生成して保存
    $attach_data = wp_generate_attachment_metadata($attach_id, $filename);
    wp_update_attachment_metadata($attach_id, $attach_data);

    // アップロード結果をレスポンスとして返す
    return new WP_REST_Response([
            'id' => $attach_id,
            'source_url' => wp_get_attachment_url($attach_id),
            'mime_type' => $mime_type,
    ], 201);
}

// GETリクエストに対応するコールバック関数
function get_usage() {
    global $wpdb;

    // データベースサイズを取得
    $tables = $wpdb->get_results("SHOW TABLE STATUS", ARRAY_A);
    $db_size = 0;

    foreach ($tables as $table) {
        $db_size += $table['Data_length'] + $table['Index_length'];
    }
    $db_size = $db_size / 1024 / 1024; // バイトからメガバイトに変換

    // ファイルシステムのサイズを取得
    $upload_dir = wp_upload_dir();
    $upload_size = folder_size($upload_dir['basedir']);
    $upload_size = $upload_size / 1024 / 1024;  // バイトからメガバイトに変換

    $amount = $db_size+$upload_size;

    $data = array(
            "db_usage" => $db_size,
            "folder_usage" => $upload_size,
            "amount" => $amount
    );

    return new WP_REST_Response($data, 200);
}

function folder_size($dir) {
    $size = 0;
    try {
        if (!is_dir($dir)) {
            return 0;
        }
        foreach (new RecursiveIteratorIterator(new RecursiveDirectoryIterator($dir)) as $file) {
            $size += $file->getSize();
        }
    } catch (Exception $e) {
        error_log("folder_size error: " . $e->getMessage());
        return 0;
    }
    return $size;
}

function rodut_enqueue_single_post_styles() {
    // シングル投稿ページの場合にのみスタイルを読み込む
    wp_enqueue_style(
            'rodut-single-post-style', // スタイルのハンドル名
            plugin_dir_url(__FILE__) . 'rodut-style.css', // CSSファイルへのパス
            array(), // 依存関係のスタイル（ない場合は空配列）
            '1.0.0' // バージョン番号
    );
}

function rodut_enqueue_slick_carousel() {
    if (is_singular('post')) {
        // SlickのCSSを読み込む
        wp_enqueue_style('slick-css', 'https://cdnjs.cloudflare.com/ajax/libs/slick-carousel/1.9.0/slick.css', array(), '1.9.0');
        wp_enqueue_style('slick-theme-css', 'https://cdnjs.cloudflare.com/ajax/libs/slick-carousel/1.9.0/slick-theme.css', array(), '1.9.0');

        // SlickのJavaScriptを読み込む
        wp_enqueue_script('slick-js', 'https://cdnjs.cloudflare.com/ajax/libs/slick-carousel/1.9.0/slick.min.js', array('jquery'), '1.9.0', true);

        // カスタムJavaScriptをインラインで追加（Slickの初期化処理）
        wp_add_inline_script('slick-js', "
        jQuery(document).ready(function(){
            jQuery('.a-root-wordpress-instagram-slider').slick({
                dots: true
            });
        });
    ");
    }
}

function rodut_ping(WP_REST_Request $request) {
    return new WP_REST_Response("ok", 200);
}

function remove_media(WP_REST_Request $request){
    $params = $request->get_json_params();

    if (!isset($params['email'])) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $email = sanitize_email($params['email']);
    $user = get_user_by('email', $email);
    if (!$user) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    // ファイル名の存在チェック
    if (!isset($params['file_names'])) {
        return new WP_REST_Response(array('error' => 'No file names provided'), 400);
    }

    $file_names = $params['file_names'];
    $deleted_files = [];

    // メディアライブラリ内のすべての添付ファイルを取得
    $args = array(
            'post_type'      => 'attachment',
            'post_status'    => 'inherit', // 添付ファイルは 'inherit' ステータスです
            'posts_per_page' => -1, // 全ファイルを取得、特定の数に制限する場合は数字を指定
    );

    $query = new WP_Query($args);

    // メディア情報をリスト化
    if ($query->have_posts()) {
        foreach ($query->posts as $post) {
            if (in_array($post->post_title, $file_names)) {
                $result = wp_delete_attachment($post->ID, true);
                if ($result) {
                    $deleted_files[] = $post->post_title;
                }
            }
        }
    }

    return new WP_REST_Response($deleted_files, 200);
}


function get_title(WP_REST_Request $request): WP_REST_Response {
    $site_title = get_bloginfo('name');
    $ary = [
            "title" => $site_title,
    ];
    return new WP_REST_Response($ary, 200);
}

function rodut_add_user(WP_REST_Request $request): WP_REST_Response {
    $params = $request->get_json_params();

    if (!isset($params['email'])) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $email = sanitize_email($params['email']);
    $user = get_user_by('email', $email);
    if (!$user) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $user_name = sanitize_text_field($params['user_name'] ?? '');
    $user_email = sanitize_email($params['user_email'] ?? '');
    $user_password = $params['user_password'] ?? '';
    $first_name = sanitize_text_field($params['first_name'] ?? '');
    $last_name = sanitize_text_field($params['last_name'] ?? '');

    if (empty($user_name) || empty($user_email) || empty($user_password)) {
        return new WP_REST_Response(array('error' => 'All fields are required'), 400);
    }

    if (email_exists($user_email)) {
        return new WP_REST_Response("ok", 200);
    }

    $user_id = wp_create_user($user_name, $user_password, $user_email);
    if (is_wp_error($user_id)) {
        $error_message = $user_id->get_error_message();  // エラーメッセージを取得
        return new WP_REST_Response(array('error' => 'ユーザーデータのinsertに失敗: ' . $error_message), 500);
    }

    // 権限を「管理者」に設定
    $user = new WP_User($user_id);

    $role = "administrator";
    if (isset($params['role']) && trim($params['role']) == "original") {
        $role = "original";
    }
    $user->set_role($role);

    update_user_meta($user_id, 'first_name', $first_name);
    update_user_meta($user_id, 'last_name', $last_name);

    return new WP_REST_Response("ok", 200);
}

function rodut_public_site(WP_REST_Request $request): WP_REST_Response {
    $params = $request->get_json_params();

    if (!isset($params['email'])) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $email = sanitize_email($params['email']);
    $user = get_user_by('email', $email);
    if (!$user) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    // 「ウェブサイトを公開しました」「ウェブサイトをリニューアルしました」の投稿を検索（完全一致）
    $posts = get_posts(array(
            'post_type' => 'post',
            'post_status' => 'publish',
            'numberposts' => -1,
            'meta_query' => array()
    ));

    $target_post = null;
    foreach ($posts as $post) {
        if ($post->post_title === 'ウェブサイトを公開しました') {
            $target_post = $post;
            break;
        }
        if ($post->post_title === 'ウェブサイトをリニューアルしました') {
            $target_post = $post;
            break;
        }
    }

    if (!$target_post) {
        return new WP_REST_Response(array('error' => 'Target post not found'), 404);
    }

    // 公開日を現在日時に更新
    $updated_post = array(
            'ID' => $target_post->ID,
            'post_date' => current_time('mysql'),
            'post_date_gmt' => current_time('mysql', 1)
    );

    $result = wp_update_post($updated_post);

    if (is_wp_error($result) || $result === 0) {
        return new WP_REST_Response(array('error' => 'Failed to update post'), 500);
    }

    return new WP_REST_Response(array(
            'message' => 'Post updated successfully',
            'post_id' => $target_post->ID,
            'new_date' => current_time('mysql')
    ), 200);
}

function rodut_toggle_lightstart(WP_REST_Request $request): WP_REST_Response {
    $params = $request->get_json_params();

    if (!isset($params['email'])) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $email = sanitize_email($params['email']);
    $user = get_user_by('email', $email);
    if (!$user) {
        return new WP_REST_Response(array('error' => 'Invalid email'), 400);
    }

    $enable = isset($params['enable']) ? (bool)$params['enable'] : false;

    $settings = get_option('wpmm_settings', array());

    if (empty($settings)) {
        $settings = array(
                'general' => array(
                        'status' => 0,
                        'status_date' => date('Y-m-d H:i:s')
                )
        );
    }

    if (!isset($settings['general'])) {
        $settings['general'] = array();
    }

    $settings['general']['status'] = $enable ? 1 : 0;
    if ($enable) {
        $settings['general']['status_date'] = date('Y-m-d H:i:s');
    }

    $result = update_option('wpmm_settings', $settings);

    if (!$result && get_option('wpmm_settings') !== $settings) {
        return new WP_REST_Response(array('error' => 'Failed to update LightStart settings'), 500);
    }

    if (function_exists('wpmm_delete_cache')) {
        wpmm_delete_cache();
    }

    return new WP_REST_Response(array(
            'message' => $enable ? 'LightStart maintenance mode enabled' : 'LightStart maintenance mode disabled',
            'lightstart_status' => $enable
    ), 200);
}


function send_slack_message($webhook_url, $message) {
    $payload = json_encode(['text' => $message]);

    $args = [
            'body' => $payload,
            'headers' => [
                    'Content-Type' => 'application/json',
            ],
    ];

    $response = wp_remote_post($webhook_url, $args);

    // レスポンスがエラーの場合はfalseを返す
    if (is_wp_error($response) || wp_remote_retrieve_response_code($response) !== 200) {
        return false;
    }

    return true;
}

function get_client_ip(WP_REST_Request $request) {
    // クライアントのIPアドレスを取得
    if (!empty($_SERVER['HTTP_CLIENT_IP'])) {
        // プロキシ経由のIPアドレス
        $ip = $_SERVER['HTTP_CLIENT_IP'];
    } elseif (!empty($_SERVER['HTTP_X_FORWARDED_FOR'])) {
        // フォワーディングされたIPアドレス
        $ip = explode(',', $_SERVER['HTTP_X_FORWARDED_FOR'])[0];
    } else {
        // リモートアドレス
        $ip = $_SERVER['REMOTE_ADDR'];
    }
    return $ip;
}


// 管理画面の設定メニュー
add_action('admin_menu', function() {
    add_options_page(
            '通知設定',
            '通知設定',
            'manage_options',
            'cf7_notifier_settings',
            'cf7_notifier_settings_page'
    );
});

// 設定ページのUI
function cf7_notifier_settings_page() {
    ?>
    <div class="wrap">
        <h1>CF7 通知設定</h1>
        <form method="post" action="options.php">
            <?php
            settings_fields('cf7_notifier_group');
            do_settings_sections('cf7_notifier_settings');
            submit_button();
            ?>
        </form>
    </div>
    <?php
}

// 設定項目の登録
add_action('admin_init', function() {
    register_setting('cf7_notifier_group', 'cf7_slack_webhook_url');
    register_setting('cf7_notifier_group', 'cf7_line_channel_token');
    register_setting('cf7_notifier_group', 'cf7_line_user_id');

    add_settings_section('cf7_notifier_section', 'Slack & LINE 通知設定', null, 'cf7_notifier_settings');

    add_settings_field('cf7_slack_webhook_url', 'Slack Webhook URL', function() {
        $val = esc_attr(get_option('cf7_slack_webhook_url', ''));
        echo "<input type='text' name='cf7_slack_webhook_url' value='$val' size='60'>";
    }, 'cf7_notifier_settings', 'cf7_notifier_section');

    add_settings_field('cf7_line_channel_token', 'LINE チャネルアクセストークン', function() {
        $val = esc_attr(get_option('cf7_line_channel_token', ''));
        echo "<input type='text' name='cf7_line_channel_token' value='$val' size='60'>";
    }, 'cf7_notifier_settings', 'cf7_notifier_section');

    add_settings_field('cf7_line_user_id', 'LINE ユーザーID（to）', function() {
        $val = esc_attr(get_option('cf7_line_user_id', ''));
        echo "<input type='text' name='cf7_line_user_id' value='$val' size='60'>";
    }, 'cf7_notifier_settings', 'cf7_notifier_section');
});

// CF7 送信前フックで通知
add_action('wpcf7_before_send_mail', function($cf7) {
    if (!class_exists('WPCF7_Submission')) {
        error_log('WPCF7_Submission クラスが存在しません');
        return;
    }

    $submission = WPCF7_Submission::get_instance();
    if (!$submission) return;

    $data = $submission->get_posted_data();

    $message = "📩 新しいお問い合わせ\n"
            . "🧑 名前: " . ($data['your-name'] ?? '未入力') . "\n"
            . "✉️ メール: " . ($data['your-email'] ?? '未入力') . "\n"
            . "💬 内容:\n" . ($data['your-message'] ?? '未入力');

    // Slack通知
    $slack_url = get_option('cf7_slack_webhook_url');
    if (!empty($slack_url)) {
        wp_remote_post($slack_url, [
                'headers' => ['Content-Type' => 'application/json'],
                'body' => json_encode(['text' => $message]),
        ]);
    }

    // LINE通知
    $channel_token = get_option('cf7_line_channel_token');
    $to_user_id = get_option('cf7_line_user_id');
    if (!empty($channel_token) && !empty($to_user_id)) {
        wp_remote_post('https://api.line.me/v2/bot/message/push', [
                'headers' => [
                        'Authorization' => 'Bearer ' . $channel_token,
                        'Content-Type' => 'application/json',
                ],
                'body' => json_encode([
                        'to' => $to_user_id,
                        'messages' => [[
                                'type' => 'text',
                                'text' => $message,
                        ]],
                ]),
        ]);
    }
});


// REST APIエンドポイントを登録するフック
add_action('rest_api_init', function() {

    register_rest_route('rodut/v1', '/version', array(
            'methods' => 'GET',
            'callback' => 'get_version',
            'show_in_index' => false,
            'permission_callback' => '__return_true',
    ));

    // healthcheck
    register_rest_route('rodut/v1', '/ping', array(
            'methods' => 'POST',
            'callback' => 'rodut_ping',
            'show_in_index' => false,
            'permission_callback' => '__return_true',
    ));

    // サーバーの容量取得のエンドポイント
    register_rest_route('rodut/v1', '/server', array(
            'methods' => 'GET',
            'callback' => 'get_usage',
            'show_in_index' => false,
            'permission_callback' => '__return_true',
    ));

    // サイトのタイトル取得、A-Rootにて使用する
    register_rest_route('rodut/v1', '/title', array(
            'methods' => 'GET',
            'callback' => 'get_title',
            'show_in_index' => false,
            'permission_callback' => '__return_true',
    ));

    register_rest_route('rodut/v1', '/posts', array(
            'methods' => 'GET',
            'callback' => 'get_post_list',
            'show_in_index' => false,
            'permission_callback' => '__return_true',
    ));

    // 記事投稿のエンドポイント
    register_rest_route('rodut/v1', '/create-post', array(
            'methods' => 'POST',
            'callback' => 'create_post',
            'show_in_index' => false,
            'permission_callback' => 'rodut_permission_check',
    ));

    // メディアダウンロードのエンドポイント
    register_rest_route('rodut/v1', '/download-media', array(
            'methods' => 'POST',
            'callback' => 'download_media',
            'show_in_index' => false,
            'permission_callback' => 'rodut_permission_check',
    ));

    // メディアアップロードのエンドポイント
    register_rest_route('rodut/v1', '/upload-media', array(
            'methods' => 'POST',
            'callback' => 'upload_media',
            'show_in_index' => false,
            'permission_callback' => 'rodut_permission_check',
    ));

    // メディアの削除エンドポイント
    register_rest_route('rodut/v1', '/remove-media', array(
            'methods' => 'POST',
            'callback' => 'remove_media',
            'show_in_index' => false,
            'permission_callback' => 'rodut_permission_check',
    ));

    // ユーザーの追加エンドポイント
    register_rest_route('rodut/v1', '/add-user', array(
            'methods' => 'POST',
            'callback' => 'rodut_add_user',
            'show_in_index' => false,
            'permission_callback' => 'rodut_permission_check',
    ));

    // 公開サイト用エンドポイント
    register_rest_route('rodut/v1', '/public_site', array(
            'methods' => 'POST',
            'callback' => 'rodut_public_site',
            'show_in_index' => false,
            'permission_callback' => 'rodut_permission_check',
    ));

    // LightStart切り替えエンドポイント
    register_rest_route('rodut/v1', '/lightstart', array(
            'methods' => 'POST',
            'callback' => 'rodut_toggle_lightstart',
            'show_in_index' => false,
            'permission_callback' => 'rodut_permission_check',
    ));
});

add_action('wp_enqueue_scripts', 'rodut_enqueue_slick_carousel');
add_action('wp_enqueue_scripts', 'rodut_enqueue_single_post_styles');

add_action( 'do_faviconico', function() {
    exit;
});
