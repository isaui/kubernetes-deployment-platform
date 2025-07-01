package utils

import "fmt"

func getCheckoutCommand(commitSHA string) string {
    if commitSHA == "" {
        return ""
    }
    return fmt.Sprintf("&& cd /workspace && echo 'Checking out commit %s...' && git fetch origin %s && git checkout %s && echo 'Commit checkout completed'", commitSHA, commitSHA, commitSHA)
}