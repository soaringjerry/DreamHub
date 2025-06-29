#!/bin/bash

# DreamHub 发布包构建脚本
# 用于生成 Windows 平台的绿色版发布包

set -e  # 遇到错误立即退出

# ========== 变量定义 ==========
# 如果没有通过环境变量或参数提供版本号，使用默认值
VERSION="${1:-${VERSION:-0.1.0}}"
GOOS="windows"
GOARCH="amd64"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RELEASE_DIR="${PROJECT_ROOT}/release"
DIST_DIR="${PROJECT_ROOT}/dist"
BUILD_DIR="${RELEASE_DIR}/DreamHub"
ZIP_NAME="DreamHub-${VERSION}-${GOOS}-${GOARCH}.zip"
FINAL_PACKAGE=""  # 最终生成的包路径

# ========== 函数定义 ==========

# 打印带颜色的信息
print_info() {
    echo -e "\033[34m[INFO]\033[0m $1"
}

print_success() {
    echo -e "\033[32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[31m[ERROR]\033[0m $1"
}

# 清理旧的构建目录
clean_build() {
    print_info "清理旧的构建目录..."
    rm -rf "${RELEASE_DIR}"
    rm -rf "${DIST_DIR}"
    print_success "清理完成"
}

# 创建目录结构
create_directories() {
    print_info "创建发布目录结构..."
    mkdir -p "${BUILD_DIR}"
    mkdir -p "${BUILD_DIR}/core"
    mkdir -p "${BUILD_DIR}/data"
    mkdir -p "${DIST_DIR}"
    print_success "目录结构创建完成"
}

# 构建 launcher
build_launcher() {
    print_info "开始构建 DreamHub Launcher (Windows ${GOARCH})..."
    
    cd "${PROJECT_ROOT}/launcher"
    
    # 设置环境变量进行跨平台编译
    export GOOS="${GOOS}"
    export GOARCH="${GOARCH}"
    export CGO_ENABLED=0  # 禁用 CGO 以生成静态链接的可执行文件
    
    # 构建命令
    # 注意：为 Windows 构建时使用 nosystray 标签，避免 CGO 依赖
    go build -tags nosystray -ldflags="-w -s" -o "${BUILD_DIR}/dreamhub.exe" ./cmd/dreamhub
    
    if [ $? -eq 0 ]; then
        print_success "Launcher 构建成功"
    else
        print_error "Launcher 构建失败"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
}

# 复制 PCAS 核心
copy_pcas_core() {
    print_info "复制 PCAS 核心文件..."
    
    # 检查预构建的 PCAS 文件是否存在
    PCAS_PREBUILT="${PROJECT_ROOT}/prebuilts/pcas.exe"
    
    if [ -f "${PCAS_PREBUILT}" ]; then
        cp "${PCAS_PREBUILT}" "${BUILD_DIR}/core/pcas.exe"
        print_success "PCAS 核心文件复制成功"
    else
        # 如果预构建文件不存在，尝试从 core 目录查找
        if [ -f "${PROJECT_ROOT}/core/pcas.exe" ]; then
            cp "${PROJECT_ROOT}/core/pcas.exe" "${BUILD_DIR}/core/pcas.exe"
            print_success "PCAS 核心文件复制成功 (从 core 目录)"
        else
            print_error "找不到 PCAS 核心文件 (pcas.exe)"
            print_info "请确保以下位置之一存在 pcas.exe:"
            print_info "  - ${PCAS_PREBUILT}"
            print_info "  - ${PROJECT_ROOT}/core/pcas.exe"
            exit 1
        fi
    fi
}

# 创建启动脚本
create_launcher() {
    print_info "创建启动脚本..."
    
    # 创建批处理文件
    cat > "${BUILD_DIR}/启动DreamHub.bat" << 'EOF'
@echo off
title DreamHub Launcher
echo 正在启动 DreamHub...
dreamhub.exe
EOF
    
    print_success "启动脚本创建成功"
}

# 创建 README 文件
create_readme() {
    print_info "创建 README.txt..."
    
    cat > "${BUILD_DIR}/README.txt" << 'EOF'
欢迎使用 DreamHub v0.3.0！

【快速启动】
双击 "启动DreamHub.bat" 文件即可启动。

⚠️ 如果直接运行 dreamhub.exe 出现错误提示，请使用批处理文件启动！

【如果遇到安全警告】
Windows Defender 可能会误报。这是因为程序未签名。
处理方法：

1. 如果下载时被阻止：
   - 在浏览器下载列表中找到文件
   - 点击"保留" -> "仍要保留"

2. 如果解压或运行时被阻止：
   - 右键点击 dreamhub.exe
   - 选择"属性"
   - 勾选"解除锁定"
   - 点击"确定"

3. 如果 Windows Defender 删除了文件：
   - 打开 Windows 安全中心
   - 点击"病毒和威胁防护"
   - 点击"保护历史记录"
   - 找到被隔离的文件，选择"还原"

【使用说明】
程序启动后会在系统托盘（屏幕右下角）创建图标。
右键点击托盘图标可以：
- 启动/停止 PCAS 服务
- 查看服务状态
- 退出程序

【命令行模式】
如需使用命令行功能，打开 cmd 并运行：
- dreamhub.exe status  查看 PCAS 状态

【文件说明】
- 启动DreamHub.bat: 推荐的启动方式
- dreamhub.exe: 主程序（通过批处理文件运行）
- core/pcas.exe: 核心服务（由主程序自动管理，请勿直接运行）

感谢您的使用！
EOF
    
    print_success "README.txt 创建成功"
}

# 打包成 ZIP
create_zip() {
    print_info "创建 ZIP 压缩包..."
    
    cd "${RELEASE_DIR}"
    
    # 使用 zip 命令创建压缩包
    if command -v zip > /dev/null 2>&1; then
        zip -r "${DIST_DIR}/${ZIP_NAME}" "DreamHub"
        FINAL_PACKAGE="${DIST_DIR}/${ZIP_NAME}"
    else
        print_error "未找到 zip 命令"
        # 在 CI 环境中，如果运行在 root 或有 sudo 权限，尝试安装 zip
        if [ "${CI}" = "true" ] && command -v apt-get > /dev/null 2>&1; then
            print_info "检测到 CI 环境，尝试安装 zip..."
            if [ "$EUID" -eq 0 ] || sudo -n true 2>/dev/null; then
                apt-get update -qq && apt-get install -qq -y zip
                if command -v zip > /dev/null 2>&1; then
                    print_success "zip 已成功安装"
                    zip -r "${DIST_DIR}/${ZIP_NAME}" "DreamHub"
                    FINAL_PACKAGE="${DIST_DIR}/${ZIP_NAME}"
                    cd "${PROJECT_ROOT}"
                    return
                fi
            fi
        fi
        
        print_info "使用 tar 作为备选..."
        tar -czf "${DIST_DIR}/${ZIP_NAME}.tar.gz" "DreamHub"
        FINAL_PACKAGE="${DIST_DIR}/${ZIP_NAME}.tar.gz"
        print_info "注意：生成的是 tar.gz 格式，而非标准的 zip 格式"
    fi
    
    cd "${PROJECT_ROOT}"
    
    if [ -f "${FINAL_PACKAGE}" ]; then
        print_success "压缩包创建成功"
    else
        print_error "压缩包创建失败"
        exit 1
    fi
}

# 显示构建结果
show_results() {
    print_info "========== 构建完成 =========="
    print_info "版本: v${VERSION}"
    print_info "目标平台: ${GOOS}/${GOARCH}"
    
    if [ -f "${FINAL_PACKAGE}" ]; then
        print_info "发布包: ${FINAL_PACKAGE}"
        print_info "文件大小: $(du -h "${FINAL_PACKAGE}" | cut -f1)"
    fi
    
    print_info ""
    print_info "发布包内容:"
    if [[ "${FINAL_PACKAGE}" == *.zip ]]; then
        unzip -l "${FINAL_PACKAGE}" 2>/dev/null || echo "（需要 unzip 工具查看内容）"
    else
        tar -tzf "${FINAL_PACKAGE}" 2>/dev/null | head -10
    fi
    
    print_success "构建成功！"
    print_info ""
    print_info "下一步："
    print_info "1. 将 ${FINAL_PACKAGE} 发送给 Windows 用户"
    print_info "2. 用户解压后双击 dreamhub.exe 即可使用"
}

# ========== 主流程 ==========
main() {
    print_info "开始构建 DreamHub v${VERSION} Windows 发布包..."
    
    clean_build
    create_directories
    build_launcher
    copy_pcas_core
    create_launcher
    create_readme
    create_zip
    show_results
}

# 执行主流程
main