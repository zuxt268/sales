FROM mysql:8.0

# 日本語ロケールをインストール
RUN apt-get update && \
    apt-get install -y locales && \
    locale-gen ja_JP.UTF-8 && \
    update-locale LANG=ja_JP.UTF-8

ENV LANG=ja_JP.UTF-8
ENV LC_ALL=ja_JP.UTF-8
ENV TZ=Asia/Tokyo