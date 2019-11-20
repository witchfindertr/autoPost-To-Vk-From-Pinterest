# Автоматический постинг в группы VK.com из Pinterest
скрипты сканируют специально подготовленные доски в Pinterest, а также плейлист на Youtube, и заносят контент в базу данных

затем автоматически публикуют в группы VK.com

перед запуском необходимо заполнить ini-файл

компиляция приложения: go build autoPost.go

запуск бинарного файла лучше реализовать через Cron


# Функции приложения:

```
./autoPost pinPicToDB - парсит контент с картинками с уже заполненой доски Pinterest

./autoPost downloadPic - загружает картинку во временную папку на сервере для дальнейшей публикации

./autoPost postPicToVk - публикует картинку на стену группы, и в альбом группы


./autoPost pinVidToDB - парсит контент с youtube-видеороликами с уже заполненой доски Pinterest

./autoPost postVidToVkGroupWall - публикует видео на стену группы

./autoPost postVidToVkAlbum - публикует видео в альбом группы
```



Automatic posting to vk.com groups from Pinterest boards

First Create tables in youre MySQL database

```go
create table curs
(
  id       int auto_increment
    primary key,
  `group`  varchar(10) not null,
  `cursor` text        not null,
  writes   tinyint     not null
);

create table pinTut
(
  id         int auto_increment
    primary key,
  media      varchar(10)       not null,
  link       text              not null,
  H          smallint(6)       not null,
  W          smallint(6)       not null,
  originlink text              not null,
  note       text              not null,
  public     tinyint default 0 not null,
  idPin      bigint            not null,
  timestamp  text              null
);

create table pintArtPic
(
  id         int auto_increment
    primary key,
  media      varchar(5)        not null,
  link       text              not null,
  H          smallint(6)       not null,
  W          smallint(6)       not null,
  originlink text              not null,
  note       text              not null,
  public     tinyint default 0 not null,
  idPin      bigint            not null
);
```
