USB Drive Thief (UDT) - 便携式USB监控系统
======================================================
中文 | [English](./README_en.md)

```text
 __  __  ____    ______
/\ \/\ \/\  _`\ /\__  _\
\ \ \ \ \ \ \/\ \/_/\ \/
 \ \ \ \ \ \ \ \ \ \ \ \
  \ \ \_\ \ \ \_\ \ \ \ \
   \ \_____\ \____/  \ \_\
    \/_____/\/___/    \/_/
```

## 项目简介

一款基于Windows系统的后台服务程序，用于实时监控USB存储设备接入事件，并自动执行预定义的文件拷贝操作。请务必在合法授权范围内使用。

## 系统要求

- Windows
- 设置开机自启动需管理员权限

## 核心功能

✅ **静默后台运行**  
✅ **智能文件过滤**（正则表达式支持）  
✅ **文件大小限制**（防止大文件占用）  
✅ **多实例检测**（防止重复运行）  
✅ **权限自动提权**（UAC自动处理）  
✅ **日志跟踪**（操作记录存储于系统临时目录）

## 编译

运行build.bat即可。

## 安装与部署

### 启动

复制文件夹到指定目录 ，运行udt.exe即可。

### 设置为开机自启动

在manager中选择安装，或执行下面的指令：

```bash
# 以管理员权限运行
manager install-current-user
# 或
manager install-all-users
```

### 取消开机自启动

在manager中选择卸载，或执行下面的指令：

```bash
# 以管理员权限运行
manager.exe uninstall
```

## 使用说明

### 配置文件格式

> 注意：更改后请重启以生效

> yaml中，字符串如果没有空格或特殊字符，则无需使用引号
>
> 若使用单引号，则不适用反斜线（\）转义，单引号用两个单引号（''）表示
>
> 若使用双引号，则可以使用反斜线（\）进行转义，双引号用反斜线加双引号（\"）表示
>
> 例如：
>
> Hello! 可直接书写
>
> 单引号：'Don''t'被视为 Don't
>
> 双引号："\"Good!\""被视为 "Good!"
>
> 同时双引号内可以使用\n换行等转义字符
>
> 正则表达式请使用单引号避免转义
>
> 其他用法自行搜索

```yaml
# 是否启用。
# 为false时自动退出
#enabled: false
enabled: true

# maxSizeMB 用于指定允许处理的文件的最大大小，单位为兆字节（MB）。
# 当扫描文件时，如果文件大小超过该值，将跳过该文件不进行处理。
# 若设置为 -1，则表示不限制文件大小。
maxSizeMB: 1024

# targetDir 表示文件复制的目标目录。
# 当扫描到符合正则表达式模式的文件时，会将这些文件复制到该目录下。
# 请确保该目录存在且有写入权限。
targetDir: ./target

# regexPatterns 是一个字符串列表，其中每个字符串都是一个正则表达式模式。
# 在扫描文件时，会对文件名进行正则匹配，只要文件名满足其中一个正则表达式，该文件就会被处理（复制到目标目录）。
# 文件名使用Linux风格，即使用“/”作为路径分隔符。
# 比如：
# - '.*' # 匹配任意文件
# - '.*\.txt$' # 匹配所有以 .txt 结尾的文件
#
# 正则表达式语法简要说明：
# ^ 匹配行首
# $ 匹配行尾
# . 匹配任意单个字符
# \d 匹配数字，\w 匹配单词字符（含下划线）
# [abc] 匹配 a/b/c 中的任意一个字符
# [^abc] 匹配非 a/b/c 的任意字符
# .* 匹配0或多个任意字符（贪婪模式）
# \\d{3} 精确匹配3个数字（注意需要双反斜杠）
# a+ 匹配1个及以上数量的a
regexPatterns:
  - '.*'

# conflictStrategy 用于指定文件冲突时的处理策略。
# 可选择的值有：
# - "timestamp": 根据文件的修改时间进行比较，冲突时覆盖目标目录中较旧的文件。
# - "overwrite": 无论目标文件的修改时间如何，都直接覆盖目标文件。
# - "skip": 当发现目标文件已存在时，跳过该文件，不进行复制或覆盖操作。
conflictStrategy: timestamp

# excludeVolumeLabels 用于指定需要排除的U盘卷标
# 在此列出的卷标对应的U盘将不会进行文件复制操作
excludeVolumeLabels:
  - SKY

```

### 服务管理

在manager中执行相关指令即可

### 文件监控

1. 插入USB存储设备
2. 自动扫描符合规则的文件
3. 文件存储至 targetDir 目录

> 正在传输的文件会添加 .part 后缀
> 传输完成后自动重命名为原始文件名

## 技术实现

### 核心机制

- USB设备检测：Windows GetLogicalDrives API轮询
- 文件过滤：regexp 正则引擎
- 全局单实例控制：%TEMP%\UDT\app.lock 锁文件
- 权限管理：AdjustTokenPrivileges 提权
- 信号处理
	- 支持以下控制信号：
	  > CTRL+C (SIGINT)
	  >
	  >   CTRL+BREAK (SIGBREAK)
	  >
	  >   系统终止信号 (SIGTERM)

## 注意事项

⚠ 法律合规：使用前必须获得设备所有者授权

⚠ 防病毒排除：需添加杀毒软件白名单

⚠ 存储路径：建议使用独立加密分区

⚠ 日志清理：定期清理 %TEMP%\UDT 目录

## 授权协议

GNU-3.0 License | Copyright © 2024 skyqvn. 保留所有权利
