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


ile: ___go_build_github_com_MV7VM_url_shortener_cmd_shortener

Type: inuse_space

Time: 2026-02-10 20:52:53 MSK

Duration: 60.01s, Total samples = 2349.84kB


Showing nodes accounting for 481.36kB, 20.48% of 2349.84kB total

flat  flat%   sum%        cum   cum%

1026kB 43.66% 43.66%     1026kB 43.66%  runtime.allocm

-544.67kB 23.18% 20.48%  -544.67kB 23.18%  compress/flate.(*compressor).initDeflate (inline)

512.05kB 21.79% 42.27%   512.05kB 21.79%  internal/sync.runtime_SemacquireMutex

-512.02kB 21.79% 20.48%  -512.02kB 21.79%  reflect.(*abiSeq).assignIntN

0     0% 20.48%  -544.67kB 23.18%  compress/flate.(*compressor).init

0     0% 20.48%  -544.67kB 23.18%  compress/flate.NewWriter (inline)

0     0% 20.48%  -544.67kB 23.18%  compress/gzip.(*Writer).Write

0     0% 20.48%   512.05kB 21.79%  context.(*cancelCtx).Done

0     0% 20.48%   512.05kB 21.79%  context.(*cancelCtx).propagateCancel.func2

0     0% 20.48%  -512.02kB 21.79%  github.com/MV7VM/url-shortener/internal/app.New

0     0% 20.48%  -544.67kB 23.18%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).BatchURL

0     0% 20.48%   902.59kB 38.41%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).CreateShortURL

0     0% 20.48%  -902.59kB 38.41%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).GetUsersUrls

0     0% 20.48%  -544.67kB 23.18%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).auth

0     0% 20.48%   902.59kB 38.41%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).createController.(*Server).gzipMiddleware.func1

0     0% 20.48%  -902.59kB 38.41%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).createController.(*Server).gzipMiddleware.func11

0     0% 20.48%  -544.67kB 23.18%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).createController.(*Server).gzipMiddleware.func9

0     0% 20.48%  -544.67kB 23.18%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).createController.(*Server).withLogger.func10

0     0% 20.48%  -902.59kB 38.41%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).createController.(*Server).withLogger.func12

0     0% 20.48%   902.59kB 38.41%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*Server).createController.(*Server).withLogger.func2

0     0% 20.48%  -544.67kB 23.18%  github.com/MV7VM/url-shortener/internal/domain/url-shortener/delivery/http.(*gzipWriter).Write

0     0% 20.48% -1447.25kB 61.59%  github.com/gin-gonic/gin.(*Context).JSON (inline)

0     0% 20.48%  -544.67kB 23.18%  github.com/gin-gonic/gin.(*Context).Next

0     0% 20.48%  -544.67kB 23.18%  github.com/gin-gonic/gin.(*Context).Render

0     0% 20.48%   902.59kB 38.41%  github.com/gin-gonic/gin.(*Context).String (inline)

0     0% 20.48%  -544.67kB 23.18%  github.com/gin-gonic/gin.(*Engine).ServeHTTP

0     0% 20.48%  -544.67kB 23.18%  github.com/gin-gonic/gin.(*Engine).handleHTTPRequest

0     0% 20.48%  -544.67kB 23.18%  github.com/gin-gonic/gin.CustomRecoveryWithWriter.func1

0     0% 20.48%  -544.67kB 23.18%  github.com/gin-gonic/gin.LoggerWithConfig.func1

0     0% 20.48% -1447.25kB 61.59%  github.com/gin-gonic/gin/render.JSON.Render

0     0% 20.48%   902.59kB 38.41%  github.com/gin-gonic/gin/render.String.Render

0     0% 20.48% -1447.25kB 61.59%  github.com/gin-gonic/gin/render.WriteJSON

0     0% 20.48%   902.59kB 38.41%  github.com/gin-gonic/gin/render.WriteString

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/dig.(*Scope).Invoke

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/dig.(*constructorNode).Call

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/dig.defaultInvoker

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/dig.paramList.BuildList

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/dig.paramSingle.Build

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/fx.(*module).invoke

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/fx.(*module).invokeAll

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/fx.New

0     0% 20.48%  -512.02kB 21.79%  go.uber.org/fx.runInvoke

0     0% 20.48%   512.05kB 21.79%  internal/sync.(*Mutex).Lock (inline)

0     0% 20.48%   512.05kB 21.79%  internal/sync.(*Mutex).lockSlow

0     0% 20.48%  -512.02kB 21.79%  main.main

0     0% 20.48%  -544.67kB 23.18%  net/http.(*conn).serve

0     0% 20.48%  -544.67kB 23.18%  net/http.serverHandler.ServeHTTP

0     0% 20.48%  -512.02kB 21.79%  reflect.(*abiSeq).addArg

0     0% 20.48%  -512.02kB 21.79%  reflect.(*abiSeq).regAssign

0     0% 20.48%  -512.02kB 21.79%  reflect.Value.Call

0     0% 20.48%  -512.02kB 21.79%  reflect.Value.call

0     0% 20.48%  -512.02kB 21.79%  reflect.funcLayout

0     0% 20.48%  -512.02kB 21.79%  reflect.newAbiDesc

0     0% 20.48%  -512.02kB 21.79%  runtime.main

0     0% 20.48%      513kB 21.83%  runtime.mcall

0     0% 20.48%      513kB 21.83%  runtime.mstart

0     0% 20.48%      513kB 21.83%  runtime.mstart0

0     0% 20.48%      513kB 21.83%  runtime.mstart1

0     0% 20.48%     1026kB 43.66%  runtime.newm

0     0% 20.48%      513kB 21.83%  runtime.park_m

0     0% 20.48%     1026kB 43.66%  runtime.resetspinning

0     0% 20.48%     1026kB 43.66%  runtime.schedule

0     0% 20.48%     1026kB 43.66%  runtime.startm

0     0% 20.48%     1026kB 43.66%  runtime.wakep

0     0% 20.48%   512.05kB 21.79%  sync.(*Mutex).Lock (inline)
