package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type App struct{}

func NewApp() *App {
	return &App{}
}

// Invoke is a compatibility command entrypoint used by the frontend runtime shim.
// It mirrors Tauri invoke names to keep existing UI functionality intact during migration.
func (a *App) Invoke(command string, args map[string]any) (any, error) {
	switch command {
	case "platform":
		return normalizePlatform(runtime.GOOS), nil
	case "read_oauth_tokens":
		return a.readOAuthTokens()
	case "save_oauth_tokens":
		return true, a.saveOAuthTokens(args["tokens"])
	case "clear_oauth_tokens":
		return true, a.clearOAuthTokens()
	case "proxy_post":
		return a.proxyPost(asString(args, "url"), asString(args, "body"))
	case "read_access_token":
		return a.readAccessToken(), nil
	case "check_backend_running":
		return a.checkBackendRunning(asInt(args, "port", 8888))
	case "check_pocketpaw_version":
		return "", nil
	case "check_pocketpaw_installed":
		return map[string]any{
			"installed":      false,
			"has_config_dir": false,
			"has_cli":        false,
			"config_dir":     "",
		}, nil
	case "install_pocketpaw", "start_pocketpaw_backend":
		return true, nil
	case "toggle_quick_ask", "toggle_side_panel", "hide_quick_ask":
		return true, nil
	case "quickask_to_sidepanel", "detach_side_panel", "collapse_side_panel", "expand_side_panel":
		return true, nil
	case "set_attach_mode", "set_vibrancy_theme":
		return true, nil
	case "window_close", "window_start_dragging", "window_hide", "window_minimize", "window_toggle_maximize":
		return true, nil
	case "window_is_maximized":
		return false, nil
	case "updater_check":
		return nil, nil
	case "updater_download_and_install":
		return true, nil
	case "app_relaunch":
		return true, nil
	case "dialog_pick_directory":
		return nil, nil
	case "autostart_is_enabled":
		return false, nil
	case "autostart_set_enabled":
		return true, nil
	case "notifications_request_permission":
		return false, nil
	case "notifications_send":
		return true, nil
	case "log_write":
		return true, nil
	case "get_pending_quickask":
		return nil, nil
	case "is_side_panel_collapsed":
		return false, nil
	case "get_attach_mode":
		return "docked", nil
	case "get_active_context":
		return map[string]any{
			"app_name":     "",
			"window_title": "",
			"file_path":    nil,
			"icon":         "🐾",
		}, nil
	case "fs_parent_dir":
		return filepath.Dir(asString(args, "path")), nil
	case "fs_resolve_path":
		return a.fsResolvePath(asString(args, "path"), asString(args, "baseDir")), nil
	case "fs_thumbnail":
		return nil, errors.New("fs_thumbnail is not implemented in Wails migration yet")
	case "fs_read_dir":
		return a.fsReadDir(asString(args, "path"))
	case "fs_read_file_text":
		return a.fsReadFileText(asString(args, "path"))
	case "fs_write_file":
		return true, a.fsWriteFile(asString(args, "path"), asString(args, "content"))
	case "fs_delete":
		return true, a.fsDelete(asString(args, "path"), asBool(args, "recursive"))
	case "fs_rename":
		return true, os.Rename(asString(args, "oldPath"), asString(args, "newPath"))
	case "fs_stat":
		return a.fsStat(asString(args, "path"))
	case "fs_create_dir":
		return true, os.MkdirAll(asString(args, "path"), 0o755)
	case "fs_exists":
		_, err := os.Stat(asString(args, "path"))
		return err == nil, nil
	case "fs_get_default_dirs":
		home, _ := os.UserHomeDir()
		return map[string]string{
			"home":      home,
			"documents": filepath.Join(home, "Documents"),
			"downloads": filepath.Join(home, "Downloads"),
			"desktop":   filepath.Join(home, "Desktop"),
		}, nil
	case "fs_read_file_head":
		return a.fsReadFileHead(asString(args, "path"), asInt(args, "maxBytes", 2048))
	case "fs_read_file_base64":
		return a.fsReadFileBase64(asString(args, "path"))
	case "fs_copy_file":
		return true, a.fsCopyFile(asString(args, "src"), asString(args, "dest"))
	case "fs_copy_dir":
		return true, a.fsCopyDir(asString(args, "src"), asString(args, "dest"))
	case "fs_stat_extended":
		return a.fsStatExtended(asString(args, "path"))
	case "fs_open_in_terminal":
		return true, errors.New("fs_open_in_terminal is not implemented in Wails migration yet")
	case "fs_search_recursive":
		return a.fsSearchRecursive(
			asString(args, "rootPath"),
			asString(args, "query"),
			asInt(args, "maxResults", 500),
			asInt(args, "maxDepth", 10),
		)
	default:
		return nil, fmt.Errorf("wails migration: command '%s' is not implemented yet", command)
	}
}

func (a *App) readOAuthTokens() (any, error) {
	path := a.oauthTokensPath()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (a *App) saveOAuthTokens(tokens any) error {
	if tokens == nil {
		return errors.New("missing tokens payload")
	}
	if err := os.MkdirAll(filepath.Dir(a.oauthTokensPath()), 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(tokens)
	if err != nil {
		return err
	}
	return os.WriteFile(a.oauthTokensPath(), b, 0o600)
}

func (a *App) clearOAuthTokens() error {
	if err := os.Remove(a.oauthTokensPath()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (a *App) readAccessToken() string {
	v, err := a.readOAuthTokens()
	if err != nil || v == nil {
		return ""
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	return asString(obj, "access_token")
}

func (a *App) oauthTokensPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pocketpaw", "oauth_tokens.json")
}

func (a *App) proxyPost(url, body string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("http %d: %s", res.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func (a *App) checkBackendRunning(port int) (bool, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 800*time.Millisecond)
	if err != nil {
		return false, nil
	}
	_ = conn.Close()
	return true, nil
}

type fileEntry struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDir     bool   `json:"is_dir"`
	Size      int64  `json:"size"`
	Modified  int64  `json:"modified"`
	Extension string `json:"extension"`
}

func (a *App) fsReadDir(path string) ([]fileEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]fileEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fullPath := filepath.Join(path, entry.Name())
		result = append(result, toFileEntry(fullPath, info))
	}
	return result, nil
}

func (a *App) fsReadFileText(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (a *App) fsWriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

func (a *App) fsDelete(path string, recursive bool) error {
	if recursive {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

func (a *App) fsStat(path string) (fileEntry, error) {
	info, err := os.Stat(path)
	if err != nil {
		return fileEntry{}, err
	}
	return toFileEntry(path, info), nil
}

func (a *App) fsReadFileHead(path string, maxBytes int) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(b) > maxBytes {
		b = b[:maxBytes]
	}
	return string(b), nil
}

func (a *App) fsReadFileBase64(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return "data:application/octet-stream;base64," + base64.StdEncoding.EncodeToString(b), nil
}

func (a *App) fsCopyFile(src, dest string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, b, 0o644)
}

func (a *App) fsCopyDir(src, dest string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return a.fsCopyFile(path, target)
	})
}

func (a *App) fsStatExtended(path string) (map[string]any, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"name":       info.Name(),
		"path":       path,
		"is_dir":     info.IsDir(),
		"size":       info.Size(),
		"modified":   info.ModTime().UnixMilli(),
		"created":    info.ModTime().UnixMilli(),
		"extension":  strings.TrimPrefix(filepath.Ext(info.Name()), "."),
		"readonly":   info.Mode().Perm()&0o222 == 0,
		"is_symlink": info.Mode()&os.ModeSymlink != 0,
	}, nil
}

func (a *App) fsSearchRecursive(rootPath, query string, maxResults, maxDepth int) (map[string]any, error) {
	needle := strings.ToLower(query)
	var scanned int
	result := make([]fileEntry, 0, maxResults)
	truncated := false

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		rel, relErr := filepath.Rel(rootPath, path)
		if relErr == nil && rel != "." {
			depth := len(strings.Split(rel, string(os.PathSeparator)))
			if depth > maxDepth && d.IsDir() {
				return filepath.SkipDir
			}
		}

		scanned++
		if len(result) >= maxResults {
			truncated = true
			return filepath.SkipDir
		}

		name := strings.ToLower(d.Name())
		if needle == "" || strings.Contains(name, needle) {
			info, infoErr := d.Info()
			if infoErr == nil {
				result = append(result, toFileEntry(path, info))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"entries":       result,
		"total_scanned": scanned,
		"truncated":     truncated,
	}, nil
}

func (a *App) fsResolvePath(path, baseDir string) string {
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	if baseDir == "" {
		baseDir, _ = os.Getwd()
	}
	return filepath.Join(baseDir, path)
}

func toFileEntry(path string, info os.FileInfo) fileEntry {
	return fileEntry{
		Name:      info.Name(),
		Path:      path,
		IsDir:     info.IsDir(),
		Size:      info.Size(),
		Modified:  info.ModTime().UnixMilli(),
		Extension: strings.TrimPrefix(filepath.Ext(info.Name()), "."),
	}
}

func asString(args map[string]any, key string) string {
	if args == nil {
		return ""
	}
	v, ok := args[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		b, _ := json.Marshal(t)
		return strings.Trim(string(b), "\"")
	}
}

func asBool(args map[string]any, key string) bool {
	if args == nil {
		return false
	}
	v, ok := args[key]
	if !ok {
		return false
	}
	b, _ := v.(bool)
	return b
}

func asInt(args map[string]any, key string, def int) int {
	if args == nil {
		return def
	}
	v, ok := args[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case float32:
		return int(n)
	default:
		return def
	}
}

func normalizePlatform(goos string) string {
	switch goos {
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	default:
		return "unknown"
	}
}
