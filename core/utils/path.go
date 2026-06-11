package utils

import "strings"

// NormalizeRemotePath cleans an SMB path expressed with backslash separators
// relative to a share root. It resolves "." and ".." segments, collapses
// repeated separators, and returns a path WITHOUT a leading or trailing
// backslash. The share root is represented by the empty string.
func NormalizeRemotePath(path string) string {
	// Treat forward slashes as backslashes for convenience.
	path = strings.ReplaceAll(path, "/", "\\")

	segments := []string{}
	for _, segment := range strings.Split(path, "\\") {
		switch segment {
		case "", ".":
			// Skip empty segments (from leading/duplicate separators) and ".".
			continue
		case "..":
			if len(segments) > 0 {
				segments = segments[:len(segments)-1]
			}
		default:
			segments = append(segments, segment)
		}
	}

	return strings.Join(segments, "\\")
}

// JoinRemotePath resolves target against base. If target is absolute (starts
// with a separator), base is ignored. The result is normalized and contains no
// leading or trailing backslash.
func JoinRemotePath(base, target string) string {
	target = strings.ReplaceAll(target, "/", "\\")

	if strings.HasPrefix(target, "\\") {
		// Absolute path relative to the share root.
		return NormalizeRemotePath(target)
	}

	return NormalizeRemotePath(base + "\\" + target)
}

// RemoteBase returns the last segment of a remote (backslash-separated) path,
// e.g. "a\\b\\c.txt" -> "c.txt". An empty or root path yields "".
func RemoteBase(path string) string {
	normalized := NormalizeRemotePath(path)
	if normalized == "" {
		return ""
	}
	if idx := strings.LastIndex(normalized, "\\"); idx >= 0 {
		return normalized[idx+1:]
	}
	return normalized
}

// RemoteDir returns everything but the last segment of a remote path,
// normalized, e.g. "a\\b\\c.txt" -> "a\\b". A single-segment or empty path
// yields "".
func RemoteDir(path string) string {
	normalized := NormalizeRemotePath(path)
	if idx := strings.LastIndex(normalized, "\\"); idx >= 0 {
		return normalized[:idx]
	}
	return ""
}

// ToWirePath converts a normalized share-relative path into the leading-
// backslash form expected by the SMB client (e.g. "" -> "\", "a\b" -> "\a\b").
func ToWirePath(path string) string {
	if path == "" {
		return "\\"
	}
	return "\\" + path
}
