FROM php:7.2-fpm

RUN pecl install redis mongodb \
		&& docker-php-ext-enable redis mongodb \
		&& docker-php-ext-install mysqli pdo pdo_mysql exif json pcntl

COPY ./etc/php-fpm.conf /usr/local/etc/php-fpm.conf
COPY ./etc/php.ini /usr/local/etc/php/php.ini

WORKDIR /data

EXPOSE 9000

CMD ["php-fpm"]
