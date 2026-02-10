package main

import (
	"fmt"
	"nofx/mcp"
	"os"
)

func main() {
	fmt.Printf("GuardianBrowserDataDir = '%s'\n", mcp.GuardianBrowserDataDir)
	fmt.Printf("环境变量 BROWSER_DATA_DIR = '%s'\n", os.Getenv("BROWSER_DATA_DIR"))
}
