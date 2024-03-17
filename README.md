# coursesGo
Разработчик: Кайсин Алишер 21B030833

Описание проекта: Курсовой сайт

Проект представляет собой платформу для обмена курсами, где пользователи могут создавать свои курсы, проходить курсы других пользователей, а также обмениваться опытом и знаниями. Сайт предоставляет возможность регистрации пользователей, аутентификации через токены и обеспечивает функционал API для взаимодействия с данными.

Здоровье приложения:

GET /v1/healthcheck - Проверка состояния здоровья приложения.
Курсы:

GET /v1/courses - Получение списка всех курсов.
POST /v1/courses - Создание нового курса.
GET /v1/courses/:id - Получение информации о конкретном курсе.
PATCH /v1/courses/:id - Обновление информации о курсе.
DELETE /v1/courses/:id - Удаление курса.
Пользователи:

POST /v1/users - Регистрация нового пользователя.
PUT /v1/users/activated - Активация пользователя.
Токены аутентификации:

POST /v1/tokens/authentication - Создание токена аутентификации.

Структура сущностей в базе данных:
Table courses {
  id bigserial [pk]
  created_at timestamp(0) with time zone [not null, default: `NOW()`]
  updated_at timestamp(0) with time zone [not null, default: `NOW()`]
  title text [not null]
  year integer [not null]
  runtime integer [not null]
  subjects text[] [not null]
  version integer [not null, default: 1]
}

Table users {
  id bigserial [pk]
  created_at timestamp(0) with time zone [not null, default: `NOW()`]
  name text [not null]
  email citext [unique, not null]
  password_hash bytea [not null]
  activated bool [not null]
  version integer [not null, default: 1]
}

Table tokens {
  hash bytea [pk]
  user_id bigint [not null, ref: > users.id]
  expiry timestamp(0) with time zone [not null]
  scope text [not null]
}

Table permissions {
  id bigserial [pk]
  code text [not null]
}

Table users_permissions {
  user_id bigint [ref: > users.id]
  permission_id bigint [ref: > permissions.id]
  
  index [pk]
}
