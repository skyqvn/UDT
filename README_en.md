USB Drive Thief (UDT) - Portable USB Monitoring System
======================================================
[中文](./README.md) | English

```text
 __  __  ____    ______
/\ \/\ \/\  _`\ /\__  _\
\ \ \ \ \ \ \/\ \/_/\ \/
 \ \ \ \ \ \ \ \ \ \ \ \
  \ \ \_\ \ \ \_\ \ \ \ \
   \ \_____\ \____/  \ \_\
    \/_____/\/___/    \/_/
```

## Project Introduction

A background service program based on the Windows system, used for real-time monitoring of USB storage device connection
events and automatically executing predefined file copy operations. Please use it within the scope of legal
authorization.

## System requirements

- Windows
- Set the computer to start automatically requires administrator privileges

## Core Function

✅ **Running in silent background**
✅ **Intelligent File Filtering** (Regular Expression Supported)
File Size Limit (To prevent large files from occupying)
✅ **Multisample Detection** (Prevent Repeated Execution)
✅ **Automated Permission Promotion** (UAC Automatic Handling)
✅ **Log Tracking** (Operation records stored in the system's temporary directory)

## Compile

Run build.bat to proceed.

## Installation and Deployment

### Start

Copy the folder to the specified directory and run udt.exe.

### Set to boot automatically

In the manager, select Install, or execute the following command:

```bash
# Run as administrator
manager install-current-user
# Or
manager install-all-users
```

### Disable startup auto-launch

In the manager, select uninstall, or execute the following command:

```bash
# Run as administrator
manager.exe uninstall
```

## User Instructions

### Configuration file format

> Please restart after changes to take effect

```yaml
# Whether to enable.
# Automatically exit when set to false.
#enabled: false
enabled: true

# maxSizeMB is used to specify the maximum size of files allowed to be processed, in megabytes (MB).
# When scanning files, if the file size exceeds this value, the file will be skipped and not processed.
# If set to -1, it means there is no limit on the file size.
maxSizeMB: 1024

# targetDir represents the target directory for file copying.
# When a file that matches the regular expression pattern is scanned, it will be copied to this directory.
# Please ensure that this directory exists and has write permissions.
targetDir: "./target"

# regexPatterns is a list of strings, where each string is a regular expression pattern.
# When scanning files, the file name will be regex - matched. As long as the file name meets one of the regular expressions, the file will be processed (copied to the target directory).
# File names use the Linux style, i.e., use "/" as the path separator.
# For example:
# - ".*" # Matches any file
# - ".*\.txt$" # Matches all files ending with .txt
#
# Brief explanation of regular expression syntax:
# ^ Matches the start of a line
# $ Matches the end of a line
# . Matches any single character
# \d Matches a digit, \w Matches a word character (including underscore)
# [abc] Matches any one of a, b, or c
# [^abc] Matches any character other than a, b, or c
# .* Matches 0 or more arbitrary characters (greedy mode)
# \\d{3} Exactly matches 3 digits (note the need for double backslashes)
# a+ Matches 1 or more occurrences of a
regexPatterns:
  - ".*"

# conflictStrategy is used to specify the handling strategy when a file conflict occurs.
# Available options are:
# - "timestamp": Compare based on the file's modification time. When a conflict occurs, overwrite the older file in the target directory.
# - "overwrite": Regardless of the target file's modification time, directly overwrite the target file.
# - "skip": When the target file already exists, skip the file and do not perform the copy or overwrite operation.
conflictStrategy: "timestamp"

# excludeVolumeLabels is used to specify the volume labels of USB drives to be excluded.
# USB drives with the volume labels listed here will not undergo file copy operations.
excludeVolumeLabels:
  - "SKY"

```

### Service management

Execute the relevant commands in the manager

### File monitoring

1. Insert USB storage device
2. Automatically scan files that meet the rules
3. The file is stored in the targetDir directory

> The file being transferred will be added with a .part suffix, and it will be automatically renamed to the original
> filename after the transfer is complete.

## Technical implementation

### Core mechanism

- USB device detection: Windows GetLogicalDrives API polling
- File filtering: regexp2 regular expression engine
- Global single instance control: %TEMP%\\UDT\\app.lock lock file
- Privilege Management: AdjustTokenPrivileges Authorization
- Signal processing
	- Support the following control signals:
	  > CTRL+C (SIGINT)
	  >
	  > CTRL+BREAK (SIGBREAK)
	  >
	  > System termination signal (SIGTERM)

## Caution

⚠ Legal compliance: Authorization from the device owner is required before use

⚠ Antivirus exclusion: need to add antivirus software whitelist

⚠ Storage path: It is recommended to use an independent encrypted partition

⚠ Log cleaning: Regularly clean the %TEMP%\UDT directory

## Authorization Agreement

GNU-3.0 License | Copyright © 2024 skyqvn. All rights reserved.

