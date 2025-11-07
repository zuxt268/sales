FROM mysql:8.0

# 日本語ロケールをインストール
RUN microdnf install -y glibc-langpack-ja

ENV LANG=ja_JP.UTF-8
ENV LC_ALL=ja_JP.UTF-8
ENV TZ=Asia/Tokyo