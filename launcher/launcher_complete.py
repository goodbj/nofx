import tkinter as tk
from tkinter import scrolledtext, END, INSERT, messagebox, Menu
import subprocess
import threading
import sys
import os
from pathlib import Path
import time
import queue

class NOFXLauncher:
    def __init__(self, root):
        self.root = root
        self.root.title("NOFX 开发版启动器 v1.0")
        self.root.geometry("900x700")
        
        # 设置窗口图标（如果有的话）
        # self.root.iconbitmap('icon.ico')  # 可选
        
        # 创建菜单栏
        menubar = Menu(root)
        root.config(menu=menubar)
        
        help_menu = Menu(menubar, tearoff=0)
        menubar.add_cascade(label="帮助", menu=help_menu)
        help_menu.add_command(label="关于", command=self.show_about)
        
        # 创建主框架
        main_frame = tk.Frame(root)
        main_frame.pack(fill=tk.BOTH, expand=True, padx=10, pady=10)
        
        # 创建标题
        title_label = tk.Label(
            main_frame,
            text="NOFX AI 交易系统开发版启动器",
            font=("Arial", 16, "bold"),
            fg="#2c3e50"
        )
        title_label.pack(pady=(0, 10))
        
        # 创建按钮框架
        button_frame = tk.Frame(main_frame)
        button_frame.pack(fill=tk.X, pady=(0, 10))
        
        # 启动后端按钮
        self.start_backend_btn = tk.Button(
            button_frame, 
            text="🚀 启动后端服务 (端口 8888)", 
            command=self.start_backend,
            bg="#4CAF50",
            fg="white",
            font=("Arial", 10, "bold"),
            height=2,
            width=20
        )
        self.start_backend_btn.pack(side=tk.LEFT, padx=(0, 5))
        
        # 启动前端按钮
        self.start_frontend_btn = tk.Button(
            button_frame, 
            text="🌐 启动前端服务 (端口 3300)", 
            command=self.start_frontend,
            bg="#2196F3",
            fg="white",
            font=("Arial", 10, "bold"),
            height=2,
            width=20
        )
        self.start_frontend_btn.pack(side=tk.LEFT, padx=5)
        
        # 启动全部服务按钮
        self.start_all_btn = tk.Button(
            button_frame,
            text="⚡ 启动全部服务",
            command=self.start_all,
            bg="#FF9800",
            fg="white",
            font=("Arial", 10, "bold"),
            height=2,
            width=15
        )
        self.start_all_btn.pack(side=tk.LEFT, padx=5)
        
        # 停止所有服务按钮
        self.stop_all_btn = tk.Button(
            button_frame,
            text="🛑 停止所有服务",
            command=self.stop_all,
            bg="#f44336",
            fg="white",
            font=("Arial", 10, "bold"),
            height=2,
            width=15
        )
        self.stop_all_btn.pack(side=tk.LEFT, padx=5)
        
        # 状态标签
        self.status_label = tk.Label(
            main_frame,
            text="状态: 待命 | 后端: ❌ | 前端: ❌",
            font=("Arial", 10),
            fg="#7f8c8d"
        )
        self.status_label.pack(pady=(5, 10))
        
        # 日志文本框
        self.log_text = scrolledtext.ScrolledText(
            main_frame,
            wrap=tk.WORD,
            width=80,
            height=30,
            state='normal',
            font=("Consolas", 9)
        )
        self.log_text.pack(fill=tk.BOTH, expand=True)
        
        # 服务进程存储
        self.backend_process = None
        self.frontend_process = None
        
        # 服务状态
        self.backend_running = False
        self.frontend_running = False
        
        # 输出队列
        self.output_queue = queue.Queue()
        self.running = True
        
        # 启动队列处理器
        self.process_queue()
        
        # 添加日志标题
        self.log_message("=== NOFX 开发版启动器 ===\n")
        self.log_message("开发版后端端口: 8888 | 前端端口: 3300\n")
        self.log_message("提示：请确保已安装 Docker Desktop\n")
        self.log_message("- 点击 '启动后端服务' 启动后端 API 服务 (http://localhost:8888)\n")
        self.log_message("- 点击 '启动前端服务' 启动前端 Web 服务 (http://localhost:3300)\n")
        self.log_message("- 点击 '启动全部服务' 同时启动前后端服务\n")
        self.log_message("- 点击 '停止所有服务' 停止所有正在运行的服务\n")
        self.log_message("----------------------------------------\n")
    
    def log_message(self, message):
        """向日志框添加消息"""
        self.output_queue.put(message)
    
    def process_queue(self):
        """处理输出队列"""
        try:
            while True:
                message = self.output_queue.get_nowait()
                self.log_text.insert(END, message)
                self.log_text.see(END)  # 滚动到底部
                self.root.update_idletasks()  # 立即更新界面
        except queue.Empty:
            pass
        if self.running:
            self.root.after(50, self.process_queue)  # 每50毫秒检查一次
    
    def check_docker(self):
        """检查 Docker 是否可用"""
        try:
            result = subprocess.run(["docker", "--version"], 
                                  capture_output=True, text=True, timeout=5)
            return result.returncode == 0
        except:
            return False
    
    def start_backend(self):
        """启动后端服务"""
        if not self.check_docker():
            messagebox.showerror("错误", "未找到 Docker，请确保已安装 Docker Desktop")
            return
            
        if self.backend_running:
            self.log_message("后端服务已在运行中...\n")
            return
            
        self.log_message("正在启动后端服务 (端口 8888)...\n")
        
        try:
            # 使用 docker-compose.dev.watch.yml 启动开发版后端
            self.backend_process = subprocess.Popen(
                ["docker", "compose", "-f", "docker-compose.dev.watch.yml", "up", "nofx-dev-watch"],
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                bufsize=1,
                cwd=os.getcwd()  # 设置工作目录为当前目录
            )
            
            self.backend_running = True
            self.update_status()
            
            # 启动线程读取输出
            threading.Thread(target=self.read_backend_output, daemon=True).start()
            
            self.log_message("✅ 后端服务启动成功！\n")
            self.log_message("后端服务将在 http://localhost:8888 运行\n")
            self.log_message("API 文档: http://localhost:8888/swagger\n\n")
            
        except FileNotFoundError:
            self.log_message("❌ 错误：未找到 docker 命令，请确保已安装 Docker Desktop\n\n")
        except Exception as e:
            self.log_message(f"❌ 启动后端服务时出错：{str(e)}\n\n")
    
    def read_backend_output(self):
        """读取后端服务输出"""
        if self.backend_process:
            for line in iter(self.backend_process.stdout.readline, ''):
                if not line.strip():
                    continue
                self.log_message(f"[BACKEND] {line}")
            self.backend_process.wait()
            self.backend_running = False
            self.update_status()
    
    def start_frontend(self):
        """启动前端服务"""
        if not self.check_docker():
            messagebox.showerror("错误", "未找到 Docker，请确保已安装 Docker Desktop")
            return
            
        if self.frontend_running:
            self.log_message("前端服务已在运行中...\n")
            return
            
        self.log_message("正在启动前端服务 (端口 3300)...\n")
        
        try:
            # 使用 docker-compose.dev.watch.yml 启动开发版前端
            self.frontend_process = subprocess.Popen(
                ["docker", "compose", "-f", "docker-compose.dev.watch.yml", "up", "nofx-frontend-dev-watch"],
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                bufsize=1,
                cwd=os.getcwd()  # 设置工作目录为当前目录
            )
            
            self.frontend_running = True
            self.update_status()
            
            # 启动线程读取输出
            threading.Thread(target=self.read_frontend_output, daemon=True).start()
            
            self.log_message("✅ 前端服务启动成功！\n")
            self.log_message("前端服务将在 http://localhost:3300 运行\n")
            self.log_message("请在浏览器中打开 http://localhost:3300 访问系统\n\n")
            
        except FileNotFoundError:
            self.log_message("❌ 错误：未找到 docker 命令，请确保已安装 Docker Desktop\n\n")
        except Exception as e:
            self.log_message(f"❌ 启动前端服务时出错：{str(e)}\n\n")
    
    def read_frontend_output(self):
        """读取前端服务输出"""
        if self.frontend_process:
            for line in iter(self.frontend_process.stdout.readline, ''):
                if not line.strip():
                    continue
                self.log_message(f"[FRONTEND] {line}")
            self.frontend_process.wait()
            self.frontend_running = False
            self.update_status()
    
    def start_all(self):
        """启动所有服务"""
        self.log_message("=== 开始启动全部服务 ===\n")
        
        # 先启动后端
        self.start_backend()
        
        # 等待几秒让后端启动
        threading.Thread(target=self.delayed_start_frontend, daemon=True).start()
    
    def delayed_start_frontend(self):
        """延时启动前端"""
        time.sleep(5)  # 等待后端启动
        self.start_frontend()
    
    def stop_all(self):
        """停止所有服务"""
        self.log_message("正在停止所有服务...\n")
        
        # 停止后端服务
        if self.backend_process and self.backend_running:
            try:
                self.backend_process.terminate()
                self.backend_process.wait(timeout=5)
                self.log_message("✅ 后端服务已停止\n")
            except subprocess.TimeoutExpired:
                self.backend_process.kill()
                self.log_message("✅ 后端服务强制停止\n")
            except Exception as e:
                self.log_message(f"❌ 停止后端服务时出错：{str(e)}\n")
        
        # 停止前端服务
        if self.frontend_process and self.frontend_running:
            try:
                self.frontend_process.terminate()
                self.frontend_process.wait(timeout=5)
                self.log_message("✅ 前端服务已停止\n")
            except subprocess.TimeoutExpired:
                self.frontend_process.kill()
                self.log_message("✅ 前端服务强制停止\n")
            except Exception as e:
                self.log_message(f"❌ 停止前端服务时出错：{str(e)}\n")
        
        # 使用 docker compose 命令停止服务
        try:
            subprocess.run(["docker", "compose", "-f", "docker-compose.dev.watch.yml", "down"], 
                         capture_output=True, timeout=10, cwd=os.getcwd())
        except:
            pass  # 忽略错误，因为服务可能已经停止
        
        self.backend_running = False
        self.frontend_running = False
        self.update_status()
        
        self.log_message("✅ 所有服务已停止\n\n")
    
    def update_status(self):
        """更新状态标签"""
        backend_status = "✅" if self.backend_running else "❌"
        frontend_status = "✅" if self.frontend_running else "❌"
        status_text = f"状态: "
        if self.backend_running and self.frontend_running:
            status_text += "全部运行"
        elif self.backend_running:
            status_text += "仅后端运行"
        elif self.frontend_running:
            status_text += "仅前端运行"
        else:
            status_text += "待命"
        status_text += f" | 后端: {backend_status} | 前端: {frontend_status}"
        
        self.status_label.config(text=status_text)
    
    def show_about(self):
        """显示关于对话框"""
        messagebox.showinfo("关于", "NOFX 开发版启动器 v1.0\n\n用于便捷启动和管理 NOFX AI 交易系统的前后端服务\n\n开发版后端端口: 8888\n开发版前端端口: 3300")

def main():
    root = tk.Tk()
    app = NOFXLauncher(root)
    
    def on_closing():
        """关闭窗口时停止所有服务"""
        app.running = False
        app.stop_all()
        root.destroy()
    
    root.protocol("WM_DELETE_WINDOW", on_closing)
    root.mainloop()

if __name__ == "__main__":
    main()