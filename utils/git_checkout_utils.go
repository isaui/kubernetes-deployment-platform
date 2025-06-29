package utils

import "fmt"

// Helper to get the appropriate git checkout command
func getCheckoutCommand(commitSHA string) string {
    if commitSHA == "" {
        return ""
    }
    return fmt.Sprintf("&& echo 'Checking out commit %s...' && git fetch origin %s && git checkout %s && echo 'Commit checkout completed'", commitSHA, commitSHA, commitSHA)
}