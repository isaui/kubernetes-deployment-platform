package utils

import (
	"path"
	"strings"
)

// ExtractRepoName mengekstrak nama repository dari URL GitHub
func ExtractRepoName(repoURL string) string {
	// Hapus .git di akhir URL jika ada
	repoURL = strings.TrimSuffix(repoURL, ".git")
	
	// Ambil bagian terakhir dari path URL
	url := strings.TrimSuffix(strings.TrimPrefix(repoURL, "https://"), "/")
	parts := strings.Split(url, "/")
	
	if len(parts) < 2 {
		return "unknown-repo"
	}
	
	// Format: username/repo atau organization/repo
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}
	
	// Fallback: return just the repo name
	return path.Base(repoURL)
}

// SanitizeDirName membersihkan nama direktori dari karakter yang tidak diinginkan
func SanitizeDirName(name string) string {
	// Ganti karakter yang tidak diinginkan dengan underscore
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "*", "_")
	name = strings.ReplaceAll(name, "?", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "<", "_")
	name = strings.ReplaceAll(name, ">", "_")
	name = strings.ReplaceAll(name, "|", "_")
	return name
}
