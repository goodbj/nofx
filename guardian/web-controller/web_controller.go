package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CommandRequest struct {
	Command string `json:"command"`
	Env     string `json:"env"`
}

type CommandResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

func executeCommandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "仅支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的JSON请求", http.StatusBadRequest)
		return
	}

	// 设置工作目录到项目根目录
	projectDir := filepath.Join(filepath.Dir(os.Args[0]), "..", "..")
	os.Chdir(projectDir)

	var fullCmd string
	switch req.Command {
	case "start":
		composeFile := "docker-compose.with-guardian.yml"
		if req.Env == "development" {
			composeFile = "docker-compose.dev.with-guardian.yml"
		}
		fullCmd = fmt.Sprintf("docker-compose -f %s up -d", composeFile)
	case "stop":
		composeFile := "docker-compose.with-guardian.yml"
		if req.Env == "development" {
			composeFile = "docker-compose.dev.with-guardian.yml"
		}
		fullCmd = fmt.Sprintf("docker-compose -f %s down", composeFile)
	case "status":
		composeFile := "docker-compose.with-guardian.yml"
		if req.Env == "development" {
			composeFile = "docker-compose.dev.with-guardian.yml"
		}
		fullCmd = fmt.Sprintf("docker-compose -f %s ps", composeFile)
	case "logs":
		serviceName := "nofx-guardian"
		if req.Env == "development" {
			serviceName = "nofx-guardian-dev"
		}
		fullCmd = fmt.Sprintf("docker logs -n 50 %s", serviceName)
	default:
		http.Error(w, "未知命令", http.StatusBadRequest)
		return
	}

	// 分割命令
	cmdParts := strings.Fields(fullCmd)
	if len(cmdParts) == 0 {
		http.Error(w, "无效命令", http.StatusBadRequest)
		return
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	output, err := cmd.CombinedOutput()

	response := CommandResponse{
		Success: err == nil,
		Output:  string(output),
	}

	if err != nil {
		response.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func serveStaticFiles(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		r.URL.Path = "/start_button.html"
	}

	// 确保请求路径在guardian目录内
	requestedPath := filepath.Join("guardian", r.URL.Path)
	absRequestedPath, _ := filepath.Abs(requestedPath)
	absGuardianDir, _ := filepath.Abs("guardian")

	if !strings.HasPrefix(absRequestedPath, absGuardianDir) {
		http.Error(w, "访问被拒绝", http.StatusForbidden)
		return
	}

	http.ServeFile(w, r, requestedPath)
}

func main() {
	port := "8085"

	// 设置路由
	http.HandleFunc("/api/command", executeCommandHandler)
	http.HandleFunc("/", serveStaticFiles)

	fmt.Printf("Guardian 控制面板服务器启动在 http://localhost:%s\n", port)
	fmt.Println("功能：")
	fmt.Println("- 启动/停止包含Guardian的Docker服务")
	fmt.Println("- 检查服务状态")
	fmt.Println("- 查看Guardian日志")
	fmt.Println("\n按 Ctrl+C 停止服务器")

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
