# aicommit CLI

## 安装钩子

安装 git prepare-commit-msg 钩子到当前的 Git 仓库。

### 语法

`aicommit install`

### 例子

`aicommit install`

## 卸载钩子

卸载 git prepare-commit-msg 钩子。

### 语法

`aicommit uninstall`

### 例子

`aicommit uninstall`

## 配置设置

读取和修改 aicommit 的配置选项。

### 语法

`aicommit config`

### 例子

`aicommit config`

## 列出配置键

列出所有配置键。

### 语法

`aicommit keys`

### 例子

`aicommit keys`

## 列出配置值

列出所有配置值。

### 语法

`aicommit list`

### 例子

`aicommit list`

## 读取配置值

读取特定的配置值。

### 语法

`aicommit get <key>`

### 例子

`aicommit get api_key`

## 设置配置值

设置特定的配置值。

### 语法

`aicommit set <key> <value>`

### 例子

`aicommit set api_key "your_api_key"`

## 清除配置值

清除特定的配置值。

### 语法

`aicommit delete <key>`

### 例子

`aicommit delete api_key`

## 手动运行钩子

手动运行准备提交消息的钩子。

### 语法

`aicommit prepare-commit-msg`

### 例子

`aicommit prepare-commit-msg`

## 帮助信息

打印帮助信息或指定子命令的帮助。

### 语法

`aicommit -h` 或 `aicommit help`

### 例子

`aicommit -h` 或 `aicommit help`

## 版本信息

打印 aicommit 的版本信息。

### 语法

`aicommit -V` 或 `aicommit version`

### 例子

`aicommit -V` 或 `aicommit version`

## 详细日志

在执行 aicommit 命令时启用详细日志输出。

### 语法

`aicommit --verbose`

### 例子

`aicommit --verbose`
