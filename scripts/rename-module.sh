#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <new_module_path> [old_module_path]"
  echo "Example: $0 github.com/you/awesome"
  exit 1
fi

NEW_MOD="$1"
NEW_APP_NAME="$(basename "$NEW_MOD")"

# 自动推断 OLD_MOD：从 go.mod 第一行读取
if [[ $# -ge 2 ]]; then
  OLD_MOD="$2"
else
  if ! grep -q '^module ' go.mod; then
    echo "go.mod not found or invalid. Run in repo root."
    exit 1
  fi
  OLD_MOD="$(grep '^module ' go.mod | awk '{print $2}')"
fi

echo "OLD_MOD: $OLD_MOD"
echo "NEW_MOD: $NEW_MOD"
echo "NEW_APP_NAME: $NEW_APP_NAME"

# 1) 修改 go.mod 的 module
go mod edit -module "$NEW_MOD"

# 2) 替换源码中的 import/引用（保守做法：仅 *.go、*.md、Makefile 等常见文本）
#   注意：GNU sed 与 BSD sed 的 -i 行为不同，下面兼容处理
FILES=$(git ls-files | grep -E '\.(go|md|yaml|yml|toml|json|mk)$|(^Makefile$)' || true)

if [[ -n "$FILES" ]]; then
  if sed --version >/dev/null 2>&1; then
    # GNU sed
    sed -i "s#${OLD_MOD}#${NEW_MOD}#g" $FILES
  else
    # BSD sed (macOS)
    sed -i '' "s#${OLD_MOD}#${NEW_MOD}#g" $FILES
  fi
fi

# 修改 Makefile 中的 APP_NAME 变量（如果存在）
if grep -q '^APP_NAME\s*=' Makefile 2>/dev/null; then
  if sed --version >/dev/null 2>&1; then
    sed -i "s#^APP_NAME\s*=.*#APP_NAME = ${NEW_APP_NAME}#g" Makefile
  else
    sed -i '' "s#^APP_NAME\s*=.*#APP_NAME = ${NEW_APP_NAME}#g" Makefile
  fi
fi

# 3) 整理依赖
go mod tidy

echo "Module renamed successfully."
echo "Try: go build ./..."
