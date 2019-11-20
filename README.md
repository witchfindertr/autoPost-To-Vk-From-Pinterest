# Автоматический постинг в группы VK.com из Pinterest
скрипты сканируют специально подготовленные доски в Pinterest, заносят контент в базу данных

затем автоматически публикуют в группы VK.com

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
