# ImageSyncer

一个用于在不同容器镜像仓库之间同步镜像的 Go 工具，支持 Harbor、Azure Container Registry (ACR) 等，具有 Zstd 压缩和数据库记录功能。

## 功能特性

- 🔄 **镜像同步**：在源 Harbor 和目标 Harbor 之间同步容器镜像
- 🏢 **多注册表支持**：支持 Harbor 和 Azure Container Registry (ACR)
- 🗜️ **Zstd 压缩**：使用 Zstd 压缩算法优化传输效率
- 🗄️ **数据库记录**：将同步记录存储到 PostgreSQL 数据库
- 🔍 **智能检查**：自动检查目标仓库是否已存在镜像，避免重复同步
- 🌐 **多平台支持**：支持 Windows、Linux、macOS
- ⚙️ **环境配置**：支持通过环境变量配置所有参数
- 🛠️ **多构建工具**：提供 Makefile、批处理脚本和 PowerShell 脚本
- 🔧 **Windows 兼容**：完美解决 Windows 上的构建问题

## 系统要求

- Go 1.21+ 
- PostgreSQL 数据库
- 网络访问权限（访问源和目标 Harbor 仓库）

## 安装

### 从源码构建

1. 克隆仓库：
```bash
git clone <repository-url>
cd imagesyncer
```

2. 安装依赖：
```bash
# Linux/macOS
make deps

# Windows (PowerShell)
.\build.ps1 deps

# Windows (CMD)
.\build.bat deps
```

3. 构建：
```bash
# Linux/macOS
make build              # 构建当前平台
make build-windows      # 构建 Windows 版本
make build-linux        # 构建 Linux 版本
make build-macos        # 构建 macOS 版本
make build-all          # 构建所有平台

# Windows (PowerShell)
.\build.ps1 build       # 构建 Windows 版本
.\build.ps1 build-linux # 构建 Linux 版本
.\build.ps1 build-macos # 构建 macOS 版本

# Windows (CMD)
.\build.bat build       # 构建 Windows 版本
.\build.bat build-linux # 构建 Linux 版本
.\build.bat build-macos # 构建 macOS 版本
```

### 构建工具使用

#### Linux/macOS - Makefile
```bash
# 查看所有可用命令
make help

# 构建并运行
make run

# 开发构建（包含竞态检测）
make build-dev

# 代码格式化和检查
make fmt vet

# 运行测试
make test

# 清理构建文件
make clean
```

#### Windows - PowerShell 脚本
```powershell
# 查看所有可用命令
.\build.ps1 help

# 构建并运行
.\build.ps1 run

# 开发构建（包含竞态检测）
.\build.ps1 build-dev

# 代码格式化和检查
.\build.ps1 fmt
.\build.ps1 vet

# 运行测试
.\build.ps1 test

# 清理构建文件
.\build.ps1 clean
```

#### Windows - 批处理脚本
```cmd
# 查看所有可用命令
.\build.bat help

# 构建并运行
.\build.bat run

# 开发构建（包含竞态检测）
.\build.bat build-dev

# 代码格式化和检查
.\build.bat fmt
.\build.bat vet

# 运行测试
.\build.bat test

# 清理构建文件
.\build.bat clean
```

## 配置

### 配置加载顺序

程序启动时会按以下顺序加载配置：

1. 当前目录下的 `.env` 文件（如存在则加载）
2. 系统环境变量
3. 程序内置默认值（见 `config.go`）

你可以直接在项目根目录放置 `.env` 文件进行配置，或使用系统环境变量覆盖。
.env可以配置如下
```
DB_HOST=
DB_PORT=5432
DB_USER=root
DB_PASSWORD=123456
DB_NAME=imagesyncer

SRC_URLBASE=
SRC_USER=appviewer
SRC_PASSWORD=
SRC_TYPE=harbor

DST_URLBASE=https://stageimage.azurecr.io
DST_USER=stageimage
DST_PASSWORD=
DST_TYPE=acr
```

### 环境变量

所有配置项都有默认值：

#### POSTGRESQL数据库配置
```bash
export DB_HOST="10.10.10.10# 数据库主机
export DB_PORT="5432"               # 数据库端口
export DB_USER="root"               # 数据库用户名
export DB_PASSWORD="123456"         # 数据库密码
export DB_NAME="imagesyncer"        # 数据库名称
```

#### 源注册表配置
```bash
# Harbor 配置
export SRC_URLBASE="https://harbor.mydomain.com"  # 源 Harbor URL
export SRC_USER="app"                        # 源 Harbor 用户名
export SRC_PASSWORD="App@123"                # 源 Harbor 密码
export SRC_TYPE="harbor"                           # 注册表类型

# Azure Container Registry 配置
export SRC_URLBASE="https://myregistry.azurecr.io" # 源 ACR URL
export SRC_USER="myregistry"                       # ACR 用户名（通常是注册表名）
export SRC_PASSWORD="your-password"                # ACR 密码或访问令牌
export SRC_TYPE="acr"                              # 注册表类型
```

#### 目标注册表配置
```bash
# Harbor 配置
export DST_URLBASE="https://harbor.target.com"     # 目标 Harbor URL
export DST_USER="admin"                            # 目标 Harbor 用户名
export DST_PASSWORD="Admin@123"                    # 目标 Harbor 密码
export DST_TYPE="harbor"                           # 注册表类型

# Azure Container Registry 配置
export DST_URLBASE="https://targetregistry.azurecr.io" # 目标 ACR URL
export DST_USER="targetregistry"                   # ACR 用户名（通常是注册表名）
export DST_PASSWORD="your-password"                # ACR 密码或访问令牌
export DST_TYPE="acr"                              # 注册表类型
```

### 命令行参数

ImageSyncer 支持命令行参数来指定要同步的镜像路径：

```bash
# 同步 /library 下所有镜像（默认）
./imagesyncer

# 同步根目录下所有镜像
./imagesyncer /

# 同步指定项目下所有镜像
./imagesyncer /mywork

# 同步指定仓库下所有镜像
./imagesyncer /mywork/test

# 同步特定镜像的所有标签
./imagesyncer /mywork/test/mywork-tool

# 同步特定镜像的特定标签
./imagesyncer /mywork/test/mywork-tool:20251011-141232
```

#### 路径格式说明

- **`/`** 或 **`/library`**：同步 library 项目下所有镜像
- **`/project`**：同步指定项目下所有镜像
- **`/project/repo`**：同步指定仓库下所有镜像
- **`/project/repo/image`**：同步特定镜像的所有标签
- **`/project/repo/image:tag`**：同步特定镜像的特定标签

### Azure Container Registry 说明

#### ACR 镜像路径格式
- **Harbor**: `harbor.domain.com/project/environment/app:tag`
- **ACR**: `registry.azurecr.io/namespace/repository:tag`

#### ACR 认证方式
1. **用户名密码**：使用注册表名作为用户名，密码为访问令牌
2. **访问令牌**：通过 Azure CLI 或 Azure Portal 生成
3. **服务主体**：使用 Azure 服务主体进行认证

#### ACR 配置示例
```bash
# 使用访问令牌
export DST_URLBASE="https://myregistry.azurecr.io"
export DST_USER="myregistry"
export DST_PASSWORD="your-access-token"
export DST_TYPE="acr"

# 使用服务主体
export DST_URLBASE="https://myregistry.azurecr.io"
export DST_USER="your-service-principal-id"
export DST_PASSWORD="your-service-principal-password"
export DST_TYPE="acr"
```

## 使用方法

### 基本使用

1. 设置环境变量（可选，使用默认值）
2. 运行程序：

```bash
# 同步 /library 下所有镜像（默认）
./imagesyncer

# 同步指定路径
./imagesyncer /mywork/test

# 或使用 make
make run
```

### 程序流程

1. **解析路径参数**：根据命令行参数确定同步范围
2. **获取镜像列表**：从源注册表获取指定路径下的所有镜像
3. **批量同步**：遍历每个镜像进行同步：
   - 检查目标注册表是否已存在该镜像
   - 如果不存在，则使用 Zstd 压缩同步镜像
   - 记录同步结果到数据库
4. **完成同步**：显示同步统计信息

### 日志输出

程序提供详细的日志输出：

```
🔍 Fetching latest tag from source...
✅ Latest tag: v1.2.3
📍 Source OCI: harbor.mydomain.com/your-project/prod/your-app:v1.2.3
📍 Target OCI: harbor.target.com/your-project/prod/your-app:v1.2.3
🔍 Checking if image exists in destination...
🔄 Image not found. Syncing with Zstd compression...
✅ Image synced successfully with Zstd compression!
💾 Updating database record...
🎉 Sync job completed!
```

## 数据库结构

程序会自动创建以下数据库表：

```sql
CREATE TABLE IF NOT EXISTS image_records (
    id SERIAL PRIMARY KEY,
    image_path TEXT NOT NULL,
    src_oci TEXT NOT NULL,
    dst_oci TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (image_path)
);
```

#### 字段说明

- **`id`**：主键，自增
- **`image_path`**：镜像路径，如 `mywork/test/mywork-tool:20251011-141232`
- **`src_oci`**：源镜像的完整 OCI 地址
- **`dst_oci`**：目标镜像的完整 OCI 地址
- **`created_at`**：记录创建时间
- **`updated_at`**：记录更新时间

## 开发

### 项目结构

```
imagesyncer/
├── main.go          # 主程序文件
├── config.go        # 配置相关代码
├── go.mod           # Go 模块文件
├── go.sum           # 依赖校验文件
├── Makefile         # Linux/macOS 构建脚本
├── build.bat        # Windows 批处理构建脚本
├── build.ps1        # Windows PowerShell 构建脚本
└── README.md        # 项目说明
```

### 添加新功能

1. 修改 `main.go` 中的同步逻辑
2. 更新 `config.go` 中的配置结构（如需要）
3. 运行测试：`make test`
4. 格式化代码：`make fmt`

### 构建问题

如果在 Windows 上遇到 `gpgme` 相关构建错误，请使用：

```bash
go build -tags=containers_image_openpgp -o imagesyncer.exe .
```

或使用 Makefile：

```bash
make build-windows
```

## 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查数据库配置和网络连接
   - 确保 PostgreSQL 服务正在运行

2. **注册表 API 认证失败**
   - **Harbor**：验证用户名和密码，检查 Harbor URL 是否正确
   - **ACR**：验证访问令牌或服务主体凭据，确保有适当的权限
   - 检查注册表类型配置（`SRC_TYPE`/`DST_TYPE`）是否正确

3. **镜像同步失败**
   - 检查网络连接
   - 验证源镜像是否存在
   - 检查目标注册表权限
   - **ACR 特殊问题**：
     - 确保 ACR 注册表存在且可访问
     - 验证访问令牌未过期
     - 检查服务主体权限（如果使用）

4. **构建失败**
   - 确保 Go 版本 >= 1.21
   - 运行依赖更新命令：
     ```bash
     # Linux/macOS
     make deps
     
     # Windows
     .\build.ps1 deps
     # 或
     .\build.bat deps
     ```
   - 在 Windows 上使用构建标签：`-tags=containers_image_openpgp`

5. **测试失败**
   - 确保使用正确的构建标签运行测试
   - 所有构建脚本都已包含正确的构建标签
   - 手动运行：`go test -tags=containers_image_openpgp -v ./...`

6. **gpgme 构建约束错误**
   - 这是 Windows 上的已知问题
   - 所有提供的构建脚本都已解决此问题
   - 如果手动构建，请使用：`go build -tags=containers_image_openpgp -o imagesyncer.exe .`

### 调试模式

使用开发构建进行调试：

```bash
# Linux/macOS
make build-dev

# Windows (PowerShell)
.\build.ps1 build-dev

# Windows (CMD)
.\build.bat build-dev
```

## 许可证

[添加许可证信息]

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.0.3
- 🚀 **多级路径支持**：支持灵活的多级路径同步，如 `/mywork/test/mywork-tool:20251011-141232`
- 📝 **命令行参数**：添加了命令行参数支持，可以指定要同步的路径
- 🔄 **批量同步**：支持一次同步多个镜像，提高效率
- 🗄️ **简化数据库结构**：优化数据库表结构，使用 `image_path` 字段替代多个分离字段
- 📖 **完善文档**：更新了使用说明和示例，包含详细的路径格式说明

### v1.0.2
- 🏢 **支持 Azure Container Registry**：添加了对 ACR 的完整支持
- 🔍 **智能注册表检测**：根据配置自动选择 Harbor 或 ACR 的镜像存在性检查
- ⚙️ **增强配置选项**：添加了 `SRC_TYPE` 和 `DST_TYPE` 环境变量
- 📖 **完善 ACR 文档**：添加了详细的 ACR 配置和使用说明
- 🔧 **改进错误处理**：为不同注册表类型提供了更精确的错误信息

### v1.0.1
- 🔧 **修复测试问题**：修复了 Makefile 中测试命令缺少构建标签的问题
- 🛠️ **完善构建工具**：添加了 Windows 批处理脚本和 PowerShell 脚本
- 📖 **更新文档**：完善了 README.md 文档，添加了详细的构建说明
- 🐛 **解决 Windows 兼容性**：完美解决了 Windows 上的 gpgme 构建约束问题
- ✅ **验证所有功能**：确保所有构建脚本都能正常工作

### v1.0.0
- 初始版本
- 支持 Harbor 镜像同步
- 支持 Zstd 压缩
- 支持数据库记录
- 支持多平台构建
