#!/bin/sh
set -e

# .sshディレクトリをappuserにコピー
if [ -d /root/.ssh ]; then
    echo "Copying SSH keys to appuser..."
    mkdir -p /home/appuser/.ssh
    cp -r /root/.ssh/* /home/appuser/.ssh/ 2>/dev/null || true
    chown -R appuser:appuser /home/appuser/.ssh
    chmod 700 /home/appuser/.ssh
    chmod 600 /home/appuser/.ssh/* 2>/dev/null || true
    echo "SSH keys copied successfully"
fi

# appuserとしてアプリケーションを実行
exec su-exec appuser "$@"