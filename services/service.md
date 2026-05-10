# Services Registry

Этот файл нужен как короткая карта по `pocketbase/services`: что делает сервис и насколько он уже приведен к целевому модульному формату.

Статусы:

- `aligned` — сервис уже близок к целевому формату: есть явный `service.go`, вынесенный runtime и/или доменные подпакеты (`models/forms/apis/migrations/...`).
- `partial` — сервис уже выделен как отдельная ответственность, но еще не доведен до полного feature-module формата.
- `legacy-thin` — сервис пока больше похож на старый helper/integration package, чем на оформленный сервисный модуль.
- `infra` — это инфраструктурный слой, а не бизнес-доменный сервис; его не обязательно дотягивать до полного шаблона `forms/apis/migrations`.

## Services

| Service | Для чего нужен | Статус | Комментарий |
| --- | --- | --- | --- |
| `ai` | AI-клиенты, config resolution и provider factory/orchestration helper’ы. | `partial` | Базовый feature-module каркас выровнен; следующий шаг — постепенно вынести transport/read-side DTO и provider-specific helper code из корня в модульные подпакеты без поломки импортов. |
| `backup` | Создание, загрузка, восстановление и удаление backup-архивов. | `aligned` | Уже разделен на `service.go`/`runtime.go`, есть `forms`, `models`, `migrations`, `README`. |
| `base` | Базовый общий сервисный слой: descriptor, shared forms/models/migrations. | `infra` | Это фундамент для остальных сервисов, не стоит воспринимать как доменный feature-module. |
| `calendar` | Синхронизация календаря и хранение calendar sync/event моделей. | `partial` | Каркас приведен к общему формату, но доменные `forms/apis/queries` пока mostly пустые и часть логики живет как набор helper’ов. |
| `chat` | Домен чатов/сообщений. | `legacy-thin` | Есть признаки отдельного домена, но оформление сервиса пока минимальное. |
| `collection` | Управление коллекциями: lookup/listing/mutation/admin flows. | `aligned` | Один из самых приведенных сервисов: есть `apis/forms/models/migrations/queries/runtime/README`. |
| `container` | Работа с контейнерами и связанными query/helper потоками. | `partial` | Уже выделен как домен, но пока не доведен до полной модульной формы. |
| `credentials` | Трекинг использования credential’ов и сопутствующая runtime-логика. | `partial` | Технический сервис: оформлен каркасом, но пока больше integration/helper слой вокруг usage tracking. |
| `cron` | Пользовательские cron-задачи, scheduler и execution flow. | `aligned` | Уже получил свои `models`, `runtime`, `README`; структура понятная. |
| `feed` | Интеграции и orchestration для feed-источников и fetch/process flow. | `partial` | Каркас и README добавлены; осталось выделить доменные `queries/models` и постепенно переносить логику из helper-файлов в `service.go`. |
| `file` | Доступ к файлам, quota, zip/export, sanitization и storage helper’ы. | `partial` | Каркас и README добавлены; осталось дотянуть доменные `forms/apis/queries` и выделить миграции, если появятся file-owned схемы. |
| `health` | Легкие health-check endpoint’ы и health migrations. | `aligned` | Маленький сервис: структура приведена, `apis` и `migrations` на месте. |
| `journal` | Journal AI/reflection/query flows вокруг записей пользователя. | `legacy-thin` | Отдельный домен выглядит оправданно, но сервис почти не структурирован. |
| `log` | Логирование приложения/проектов, ingest, batching, cleanup, read-side. | `aligned` | После объединения с `logging` это канонический лог-сервис, структура уже в хорошем состоянии. |
| `mails` | Почтовые flows приложения + Gmail/mail sync/integration. | `partial` | После объединения с `mailcli` стал одним центром почтового домена, но внутри еще смешаны app-mails и user mailbox flows. |
| `memory` | Memory system: memories/entities/insights/connections/processes. | `partial` | Каркас добавлен; осталось выделить доменные `models/forms/queries` и нормализовать `query/` -> `queries/` без поломки импортов. |
| `migrate` | Migration runtime, CLI binding, template generation, automigration hooks. | `infra` | Tooling/CLI сервис: каркас приведен к общему формату, но не обязан иметь наполненные `forms/apis/queries`. |
| `newsletter` | Newsletter orchestration и runtime-интеграции. | `legacy-thin` | Похоже на нишевый домен, но модульная форма пока слабая. |
| `realtime` | Realtime subscriptions, auth sync и broadcast helpers. | `aligned` | Уже разделен на `service.go`/`runtime.go`, есть `forms` и `README`. |
| `record` | CRUD/upsert/expand/access flows для records. | `aligned` | Один из опорных сервисов: есть `apis/forms/models/migrations/queries/runtime/README`. |
| `search` | Поиск и search orchestration. | `legacy-thin` | Пока выглядит как узкий functional service; стоит решить, это отдельный домен или часть другого. |
| `settings` | Settings чтение, апдейт, admin-формы и связанные модели/миграции. | `aligned` | Уже хорошо оформлен: `service/runtime/forms/models/apis/migrations/README`. |
| `ssh` | SSH keygen + connection/session lifecycle + IO streaming домен. | `partial` | Добавлен `services/ssh` каркас; следующий шаг — вынести доменную логику из `apis/project_ssh.go` в `services/ssh/apis` и `services/ssh/runtime`. |
| `team` | Команды, membership/access control, team-specific api/form/query/model flows. | `aligned` | Один из уже приведенных сервисов. |
| `updater` | Self-update flow и orchestration релизных обновлений. | `infra` | Tooling/CLI сервис: структура приведена, но не обязан иметь наполненные `forms/apis/queries`. |
| `user` | Auth/user lifecycle, external auth, user forms/api/models/migrations. | `aligned` | Один из самых хорошо оформленных доменных сервисов. |
| `vector` | Векторные/embedding helper’ы. | `legacy-thin` | Очень похоже на технический helper, а не на самостоятельный бизнес-сервис. |
| `workflow` | Workflow engine, graph execution, nodes/connections/query layer. | `partial` | Каркас добавлен; осталось постепенно нормализовать подпакеты (`query/` -> `queries/`, runtime/state) без поломки импортов. |

## Что уже убрали

- `logging` удален, канонический сервис теперь `log`.
- `filemanager` удален, канонический сервис теперь `file`.
- `mailcli` удален, канонический сервис теперь `mails`.
- `fold` удален.
- `music` удален.
- `oauth` и `weather` сейчас сознательно убраны из текущего плана.

## Где есть смысл еще подумать про границы

### 1. Явно оставить как отдельные домены

- `collection`
- `record`
- `user`
- `team`
- `settings`
- `cron`
- `backup`
- `realtime`
- `log`
- `workflow`
- `memory`

Это либо core-домены проекта, либо понятные platform services. Их логично дальше дотягивать до общего формата, а не сливать.

### 2. Похоже на технические сервисы, а не на продуктовые домены

- `base`
- `ai`
- `credentials`
- `search`
- `vector`
- `migrate`
- `updater`

Для них не обязательно насильно строить полный шаблон `forms/apis/migrations`. Можно держать компактнее и считать infra/tooling-слоем.

### 3. Кандидаты на дополнительное укрупнение или пересмотр

- `vector` + `search` + часть `memory`
  Сейчас это близкие технические зоны. Возможно удобнее оформить как один knowledge/indexing слой, если они почти всегда ходят вместе.
- `mails` + `calendar`
  Не обязательно сливать в один сервис, но точно стоит вынести общий Google-account/OAuth/token sync helper, если эти интеграции будут жить долго.
- `journal` + `memory`
  Если journal mostly feeds memory/AI analysis, можно подумать, не является ли `journal` просто public flow поверх memory-domain.
- `feed` + `newsletter`
  Если newsletter в основном потребляет/рассылает feed-like контент, можно позже посмотреть на общий content-delivery слой.
- `chat`
  Надо понять, это core-domain или побочный integration flow. От этого зависит, развивать его как полноценный сервис или оставить компактным.
- `container`
  Если контейнерный слой важен продукту, его стоит довести до формата. Если это внутренний helper, можно оставить более узким.
- `ssh`
  Сейчас это заметный домен прямо в `apis`, но без собственного `services/ssh`. Либо стоит выделить отдельный сервис, либо честно считать это временным API-only куском.

## Практический вывод

Если продолжать рефакторинг последовательно, я бы дальше шел так:

1. Дотянуть крупные core-domain сервисы, где уже есть инерция:
   `workflow`, `memory`, `calendar`, `file`, `feed`.
2. После этого отдельно принять архитектурные решения по техническим/сомнительным сервисам:
   `ai`, `vector`, `search`, `newsletter`, `chat`.
3. Не пытаться одинаково "упаковать" infra-сервисы и feature-domain сервисы:
   `base`, `migrate`, `updater`, `credentials` могут жить легче.
