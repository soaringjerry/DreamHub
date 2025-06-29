# DreamHub 发布指南

## 版本信息
- 当前版本: v0.1.0
- 发布日期: 2025-06-29
- 支持平台: Windows (amd64)

## 构建发布包

### 前置要求

1. **开发环境**
   - Go 1.24.4 或更高版本
   - 已安装所有项目依赖

2. **系统工具**
   - `zip` 命令（推荐）: `sudo apt-get install zip`
   - 或 `tar` 命令（备选）

3. **预构建文件**
   - `prebuilts/pcas.exe`: PCAS 核心的 Windows 可执行文件
   - 如果没有预构建文件，脚本会尝试从 `core/pcas.exe` 复制

### 构建步骤

1. 在项目根目录执行构建脚本：
   ```bash
   ./build_release.sh
   ```

2. 脚本将自动执行以下操作：
   - 清理旧的构建目录
   - 交叉编译 Windows 版本的 launcher
   - 创建发布目录结构
   - 复制必要文件
   - 生成 README.txt
   - 打包成压缩文件

3. 构建完成后，发布包位于：
   - ZIP 格式: `dist/DreamHub-v0.1.0-windows-amd64.zip`
   - 或 TAR.GZ 格式: `dist/DreamHub-v0.1.0-windows-amd64.zip.tar.gz`

## 发布包结构

```
DreamHub/
├── dreamhub.exe      # 主程序（启动器）
├── README.txt        # 用户说明文档
├── core/
│   └── pcas.exe      # PCAS 核心服务
└── data/             # 运行时数据目录（空）
```

## 发布流程

1. **构建**: 运行 `./build_release.sh`
2. **测试**: 在 Windows 环境中测试发布包
3. **标记**: 在 Git 中创建版本标签
   ```bash
   git tag -a v0.1.0 -m "Release version 0.1.0"
   git push origin v0.1.0
   ```
4. **上传**: 将发布包上传到发布平台
5. **公告**: 更新项目文档和发布说明

## 用户安装说明

1. 下载发布包（.zip 文件）
2. 解压到任意目录（建议：`C:\Program Files\DreamHub`）
3. 双击运行 `dreamhub.exe`
4. 程序将在系统托盘创建图标
5. 右键点击托盘图标可以管理服务

## 注意事项

1. **防病毒软件**: 某些防病毒软件可能会误报，需要添加白名单
2. **端口占用**: 默认从 50051 开始查找可用端口
3. **权限要求**: 不需要管理员权限即可运行

## 故障排除

如果用户遇到问题：

1. **无法启动**: 检查是否有其他实例在运行
2. **端口冲突**: 查看 `data/runtime.json` 中的端口信息
3. **服务未响应**: 使用命令行运行 `dreamhub.exe status` 查看状态

## 版本历史

### v0.1.0 (2025-06-29)
- 初始版本发布
- 实现基本的进程管理功能
- 支持系统托盘操作
- 实现端口动态发现
- 实现 gRPC 健康检查