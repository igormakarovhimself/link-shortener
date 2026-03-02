# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**


## Профилирование

Бенчмарки добавлены в `internal/service/shortener_service_bench_test.go` и `internal/handler/handler_bench_test.go`.

В ходе профилирования (`pprof top/list`) под нагрузкой через `hey` (-n 10000 -c 100) обнаружил лишнюю heap-аллокацию в `generateShortURL` — `sha256.New()` создаёт объект на куче. Заменил на `sha256.Sum256()`. До исправления: 4 allocs/op, 176 B/op. После: 3 allocs/op, 144 B/op.

Результат:

```
Type: alloc_space
Showing nodes accounting for -57228069B, 0.68% of 8387274267B total

      flat  flat%   sum%        cum   cum%
-76712598B  0.91%  0.91% -44597364B  0.53%  compress/flate.NewWriter
 -6306268B 0.075%  0.56%  -6306268B 0.075%  sync.(*Pool).pinSlow
 -6293376B 0.075%  0.63%  -6293376B 0.075%  net/http.(*Request).WithContext
 -4197009B  0.05%  0.68%  -4197009B  0.05%  compress/flate.newHuffmanEncoder
 -3146160B 0.038%  0.71%  -3146160B 0.038%  net/url.parse
 -2097248B 0.025%  0.76%  -2097248B 0.025%  context.WithValue
 -1048640B 0.013%  0.71%  -1048640B 0.013%  crypto/internal/fips140/hmac.New (sha256.New)
   524312B 0.0063%  0.7%   1048624B 0.013%  link-shortener/internal/service.(*ShortenerServiceImpl).generateShortURL
```

Отрицательные значения — аллокации уменьшились. Итого -57 МБ за прогон.
