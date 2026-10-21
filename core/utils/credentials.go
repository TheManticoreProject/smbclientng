package utils

import "strings"

// SplitDomainUsername splits a domain-qualified username of the form
// DOMAIN\user or DOMAIN/user into its domain and user parts. The embedded
// domain takes precedence over the one given separately (e.g. with -d). When
// the username carries no domain, or either side of the separator is empty, the
// inputs are returned unchanged.
func SplitDomainUsername(domain, username string) (string, string) {
	for _, sep := range []string{"\\", "/"} {
		if userDomain, userName, found := strings.Cut(username, sep); found && userDomain != "" && userName != "" {
			return userDomain, userName
		}
	}
	return domain, username
}
