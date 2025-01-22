FROM openresty/openresty:alpine

# Установка необходимых пакетов
RUN apk add --no-cache \
    lua5.1 \
    lua5.1-dev \
    gcc \
    libc-dev \
    make \
    git \
    curl \
    luarocks5.1 \
    && luarocks-5.1 install lua-resty-jwt \
    && apk del \
    lua5.1-dev \
    gcc \
    libc-dev \
    make \
    git

# Создание директории для конфигурации
RUN mkdir -p /etc/nginx/conf.d

# Создание структуры директорий
RUN mkdir -p /etc/nginx/conf.d/locations

# Копирование конфигурационных файлов
COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY conf.d/ /etc/nginx/conf.d/
COPY jwt_secret.key /etc/nginx/jwt_secret.key


# Открываем порт
EXPOSE 80

# Запуск nginx
CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]
