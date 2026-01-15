import tkinter as tk
from tkinter import scrolledtext, END, INSERT
import subprocess
import threading
import sys
import os
from pathlib import Path

class NOFXLauncher:
    def __init__(self, root):
        self.root = root
        self.root.title("NOFX 启动器")
        self.root.geometry("800x600")
        
        # 设置窗口图标（如果有的话）
        # self.root.iconbitmap('icon.ico')  # 可选
        
        # 创建主框架
        main_frame = tk.Frame(root)
        main_frame.pack(fill=tk.BOTH, expand=True, padx=10, pady=10)
        
        # 创建按钮框架
        button_frame = tk.Frame(main_frame)
        button_frame.pack(fill=tk.X, pady=(0, 10))
        
        # 启动后端按钮
        self.start_backend_btn = tk.Button(
            button_frame, 
            text="启动后端服务", 
            command=self.start_backend,
            bg="#4CAF50",
            fg="white",
            font=("Arial", 10, "bold"),
            height=2
        )
        self.start_backend_btn.pack(side=tk.LEFT, padx=(0, 5))
        
        # 启动前端按钮
        self.start_frontend_btn = tk.Button(
            button_frame, 
            text="启动前端服务", 
            command=self.start_frontend,
            bg="#2196F3",
            fg="white",
            font=("Arial", 10, "bold"),
            height=2
        )
        self.start_frontend_btn.pack(side=tk.LEFT, padx=5)
        
        # 停止所有服务按钮
        self.stop_all_btn = tk.Button(
            button_frame,
            text="停止所有服务",
            command=self.stop_all,
            bg="#f44336",
            fg="white",
            font=("Arial", 10, "bold"),
            height=2
        )
        self.stop_all_btn.pack(side=tk.LEFT, padx=5)
        
        # 日志文本框
        self.log_text = scrolledtext.ScrolledText(
            main_frame,
            wrap=tk.WORD,
            width=80,
            height=30,
            state='normal'
        )
        self.log_text.pack(fill=tk.BOTH, expand=True)
        
        # 服务进程存储
        self.backend_process = None
        self.frontend_process = None
        
        # 添加日志标题
        self.log_message("=== NOFX 启动器 ===\n")
        self.log_message("提示：请确保已安装 Docker 和 Docker Compose\n")
        self.log_message("- 点击 '启动后端服务' 启动后端 API 服务\n")
        self.log_message("- 点击 '启动前端服务' 启动前端 Web 服务\n")
        self.log_message("- 点击 '停止所有服务' 停止所有正在运行的服务\n")
        
    def log_message(self, message):
        """向日志框添加消息"""
        self.log_text.insert(END, message)
        self.log_text.see(END)  # 滚动到底部
        self.root.update_idletasks()  # 立即更新界面
        
    def start_backend(self):
        """启动后端服务"""
        if self.backend_process and self.backend_process.poll() is None:
            self.log_message("后端服务已在运行中...\n")
            return
            
        self.log_message("正在启动后端服务...\n")
        
        try:
            # 使用 docker-compose.dev.watch.yml 启动开发版后端
            self.backend_process = subprocess.Popen(
                ["docker", "compose", "-f", "docker-compose.dev.watch.yml", "up", "nofx-dev-watch"],
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                bufsize=1
            )
            
            # 启动线程读取输出
            threading.Thread(target=self.read_backend_output, daemon=True).start()
            
            self.log_message("后端服务启动成功！\n")
            self.log_message("后端服务将在 http://localhost:8888 运行\n\n")
            
        except FileNotFoundError:
            self.log_message("错误：未找到 docker 命令，请确保已安装 Docker Desktop\n\n")
        except Exception as e:
            self.log_message(f"启动后端服务时出错：{str(e)}\n\n")
    
    def read_backend_output(self):
        """读取后端服务输出"""
        if self.backend_process:
            for line in iter(self.backend_process.stdout.readline, ''):
                self.log_message(f"[BACKEND] {line}")
            self.backend_process.wait()
    
    def start_frontend(self):
        """启动前端服务"""
        if self.frontend_process and self.frontend_process.poll() is None:
            self.log_message("前端服务已在运行中...\n")
            return
            
        self.log_message("正在启动前端服务...\n")
        
        try:
            # 使用 docker-compose.dev.watch.yml 启动开发版前端
            self.frontend_process = subprocess.Popen(
                ["docker", "compose", "-f", "docker-compose.dev.watch.yml", "up", "nofx-frontend-dev-watch"],
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                bufsize=1
            )
            
            # 启动线程读取输出
            threading.Thread(target=self.read_frontend_output, daemon=True).start()
            
            self.log_message("前端服务启动成功！\n")
            self.log_message("前端服务将在 http://localhost:3300 运行\n\n")
            
        except FileNotFoundError:
            self.log_message("错误：未找到 docker 命令，请确保已安装 Docker Desktop\n\n")
        except Exception as e:
            self.log_message(f"启动前端服务时出错：{str(e)}\n\n")
    
    def read_frontend_output(self):
        """读取前端服务输出"""
        if self.frontend_process:
            for line in iter(self.frontend_process.stdout.readline, ''):
                self.log_message(f"[FRONTEND] {line}")
            self.frontend_process.wait()
    
    def stop_all(self):
        """停止所有服务"""
        self.log_message("正在停止所有服务...\n")
        
        # 停止后端服务
        if self.backend_process and self.backend_process.poll() is None:
            try:
                self.backend_process.terminate()
                self.backend_process.wait(timeout=5)
                self.log_message("后端服务已停止\n")
            except subprocess.TimeoutExpired:
                self.backend_process.kill()
                self.log_message("后端服务强制停止\n")
            except Exception as e:
                self.log_message(f"停止后端服务时出错：{str(e)}\n")
        
        # 停止前端服务
        if self.frontend_process and self.frontend_process.poll() is None:
            try:
                self.frontend_process.terminate()
                self.frontend_process.wait(timeout=5)
                self.log_message("前端服务已停止\n")
            except subprocess.TimeoutExpired:
                self.frontend_process.kill()
                self.log_message("前端服务强制停止\n")
            except Exception as e:
                self.log_message(f"停止前端服务时出错：{str(e)}\n")
        
        # 使用 docker compose 命令停止服务
        try:
            subprocess.run(["docker", "compose", "-f", "docker-compose.dev.watch.yml", "down"], 
                         capture_output=True, timeout=10)
        except:
            pass  # 忽略错误，因为服务可能已经停止
            
        self.log_message("所有服务已停止\n\n")

def main():
    root = tk.Tk()
    app = NOFXLauncher(root)
    root.mainloop()

if __name__ == "__main__":
    main()