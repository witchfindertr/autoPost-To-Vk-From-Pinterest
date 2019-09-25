# autoPostToVkFromPinterest
autoPost To Vk groups from Pinterest boards

Create tables in youre MySQL database

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
