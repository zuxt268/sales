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

# tmpディレクトリの権限を修正
if [ -d /app/tmp ]; then
    echo "Setting permissions for /app/tmp..."
    chown -R appuser:appuser /app/tmp
    chmod 755 /app/tmp
fi

# credentialsディレクトリをappuserにコピー
if [ -d /app/credentials ]; then
    echo "Copying credentials to appuser..."
    mkdir -p /home/appuser/credentials
    cp -r /app/credentials/* /home/appuser/credentials/ 2>/dev/null || true
    chown -R appuser:appuser /home/appuser/credentials
    chmod 755 /home/appuser/credentials
    chmod 644 /home/appuser/credentials/* 2>/dev/null || true
    echo "Credentials copied successfully"
fi

# appuserとしてアプリケーションを実行
exec su-exec appuser "$@"