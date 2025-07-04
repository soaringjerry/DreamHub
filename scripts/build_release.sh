#!/bin/bash

# DreamHub 发布包构建脚本
# 用于生成 Windows 平台的绿色版发布包

set -e  # 遇到错误立即退出

# ========== 变量定义 ==========
# 如果没有通过环境变量或参数提供版本号，使用默认值
VERSION="${1:-${VERSION:-0.3.4}}"
GOOS="windows"
GOARCH="amd64"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RELEASE_DIR="${PROJECT_ROOT}/release"
DIST_DIR="${PROJECT_ROOT}/dist"
BUILD_DIR="${RELEASE_DIR}/DreamHub"
ZIP_NAME="DreamHub-v${VERSION}-${GOOS}-${GOARCH}.zip"
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
    mkdir -p "${BUILD_DIR}/bin"
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
    go build -tags nosystray -ldflags="-w -s" -o "${BUILD_DIR}/bin/dreamhub.exe" ./cmd/dreamhub
    
    if [ $? -eq 0 ]; then
        print_success "Launcher 构建成功"
    else
        print_error "Launcher 构建失败"
        exit 1
    fi
    
    cd "${PROJECT_ROOT}"
}

# 复制可执行文件到 bin 目录
copy_executables() {
    print_info "复制可执行文件到 bin 目录..."
    
    # 复制 PCAS 核心
    PCAS_PREBUILT="${PROJECT_ROOT}/prebuilts/pcas.exe"
    if [ -f "${PCAS_PREBUILT}" ]; then
        cp "${PCAS_PREBUILT}" "${BUILD_DIR}/bin/pcas.exe"
        print_success "PCAS 核心文件复制成功"
    else
        print_error "找不到 PCAS 核心文件: ${PCAS_PREBUILT}"
        exit 1
    fi
    
    # 复制 pcasctl
    PCASCTL_PREBUILT="${PROJECT_ROOT}/prebuilts/pcasctl.exe"
    if [ -f "${PCASCTL_PREBUILT}" ]; then
        cp "${PCASCTL_PREBUILT}" "${BUILD_DIR}/bin/pcasctl.exe"
        print_success "pcasctl 文件复制成功"
    else
        print_error "找不到 pcasctl 文件: ${PCASCTL_PREBUILT}"
        exit 1
    fi
    
    # 复制 policy.yaml 配置文件
    print_info "复制 policy.yaml 配置文件..."
    if [ -f "${PROJECT_ROOT}/core/policy.yaml" ]; then
        cp "${PROJECT_ROOT}/core/policy.yaml" "${BUILD_DIR}/core/policy.yaml"
        print_success "policy.yaml 配置文件复制成功"
    else
        print_error "找不到 policy.yaml 配置文件"
        print_info "请确保 ${PROJECT_ROOT}/core/policy.yaml 存在"
        exit 1
    fi
    
    # 复制 pcasctl.exe 命令行工具
    print_info "复制 pcasctl.exe 命令行工具..."
    if [ -f "${PROJECT_ROOT}/prebuilts/pcasctl.exe" ]; then
        cp "${PROJECT_ROOT}/prebuilts/pcasctl.exe" "${BUILD_DIR}/pcasctl.exe"
        print_success "pcasctl.exe 命令行工具复制成功"
    else
        print_error "找不到 pcasctl.exe 命令行工具"
        print_info "请确保 ${PROJECT_ROOT}/prebuilts/pcasctl.exe 存在"
        exit 1
    fi
}

# 创建启动脚本
create_launcher() {
    print_info "创建启动脚本..."
    
    # 创建批处理文件（使用 ASCII 避免编码问题）
    cat > "${BUILD_DIR}/StartDreamHub.bat" << 'EOF'
@echo off
cd /d %~dp0
start "" "bin\dreamhub.exe"
EOF
    
    print_success "启动脚本创建成功"
}

# 复制 policy.yaml 文件
copy_policy_file() {
    print_info "复制默认 policy.yaml 配置文件..."
    
    # 检查 policy.yaml 文件是否存在
    POLICY_FILE="${PROJECT_ROOT}/policy.yaml"
    
    if [ -f "${POLICY_FILE}" ]; then
        cp "${POLICY_FILE}" "${BUILD_DIR}/policy.yaml"
        print_success "policy.yaml 配置文件复制成功"
    else
        print_error "找不到 policy.yaml 配置文件"
        print_info "期望位置: ${POLICY_FILE}"
        exit 1
    fi
}

# 创建 README 文件
create_readme() {
    print_info "创建 README.txt..."
    
    cat > "${BUILD_DIR}/README.txt" << 'EOF'
欢迎使用 DreamHub v0.3.4！

请直接双击运行 StartDreamHub.bat 文件来启动程序。

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
    copy_executables
    copy_policy_file
    create_launcher
    create_readme
    create_zip
    show_results
}

# 执行主流程
main