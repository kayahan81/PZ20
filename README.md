---
# Практическое задание 20

## ЭФМО-02-25 

## Алиев Каяхан Командар оглы
---
# Тема работы
Настройка Prometheus + Grafana для метрик. Интеграция с приложением.

## Цели занятия
Научиться собирать и визуализировать метрики сервиса: трафик, ошибки, задержки, активные запросы.

## Структура проекта
```
C:.
│   .gitattributes
│   go.mod
│   go.sum
│   README.md
│   testdata.bat
│
├───.vs
│   │   ProjectSettings.json
│   │   slnx.sqlite
│   │   VSWorkspaceState.json
│   │
│   └───tech-ip-sem2
│       ├───FileContentIndex
│       │       2019e4a7-05d4-4380-9757-192646eef486.vsidx
│       │
│       └───v17
├───deploy
│   └───monitoring
│           docker-compose.yml
│           prometheus.yml
│
├───docs
│       pz17_api.md
│       pz17_diagram.md
│
├───img
├───proto
│       auth.proto
│
├───services
│   ├───auth
│   │   ├───cmd
│   │   │   └───auth
│   │   │           main.go
│   │   │
│   │   ├───internal
│   │   │   ├───config
│   │   │   ├───grpc
│   │   │   │       server.go
│   │   │   │
│   │   │   ├───handler
│   │   │   │       auth.go
│   │   │   │
│   │   │   ├───http
│   │   │   └───service
│   │   └───pkg
│   │       └───authpb
│   │               auth.pb.go
│   │               auth_grpc.pb.go
│   │
│   └───tasks
│       ├───cmd
│       │   └───tasks
│       │           main.go
│       │
│       └───internal
│           ├───client
│           │   ├───authclient
│           │   │       client.go
│           │   │
│           │   └───authgrpc
│           │           client.go
│           │
│           ├───handler
│           │       tasks.go
│           │
│           ├───http
│           ├───middleware
│           │       auth.go
│           │       metric.go
│           │
│           ├───models
│           │       tasks.go
│           │
│           ├───service
│           └───storage
│                   memory.go
│
└───shared
    ├───httpx
    │       client.go
    │
    ├───logger
    │       logger.go
    │
    └───middleware
            accesslog.go
            requestid.go
```

## В рамках учебного сервиса достаточно 3 групп метрик:
-	Счётчик запросов http_requests_total — counter
Labels: method, route (или path-шаблон), status
-	Длительность запросов
http_request_duration_seconds — histogram
Labels: method, route
Bucket’ы можно оставить дефолтные или задать учебные (например для API 0.01, 0.05, 0.1, 0.3, 1, 3)
-	Текущее число активных запросов
http_in_flight_requests — gauge
(либо без labels, либо с минимальными)


## Коды статуса:
-	200 OK — успешный ответ
-	201 Created — ресурс создан
-	204 No Content — успешно, без тела
-	400 Bad Request — неверные данные
-	404 Not Found — ресурс не найден
-	422 Unprocessable Entity — некорректные данные по смыслу
-	500 Internal Server Error — внутренняя ошибка

# Примечания по конфигурации и требования

Для запуска требуется:

Go: версия 1.25.1

<img width="841" height="232" alt="Установка Git и Go" src="https://github.com/user-attachments/assets/8e01d831-5a7f-4376-8348-9052b240aec9" />


# Команды запуска/сборки
## 1) Клонировать данный репозиторий в удобную для вас папку:
```Powershell
git clone https://github.com/kayahan81/pz20
```
## 2) Перейти в папку pz19:
```Powershell
cd pz20
```
## 3) Загрузка зависимостей:
```Powershell
go mod tidy
```
## 4) Команда запуска
В первом окне
```Powershell
$env:AUTH_PORT="8081"
$env:AUTH_GRPC_PORT="50051"
$env:ENV="development"
go run ./services/auth/cmd/auth
```
Во втором окне
```Powershell
$env:TASKS_PORT="8082"
$env:AUTH_GRPC_ADDR="localhost:50051"
$env:ENV="development"
go run ./services/tasks/cmd/tasks
```

# Проверка работоспособности
## Какие метрики собираются и какие у них labels? 
### Счётчик запросов
http_requests_total — counter<br/>
Labels: method, route (или path-шаблон), status
### 	Длительность запросов
http_request_duration_seconds — histogram<br/>
Labels: method, route<br/>
Bucket’ы можно оставить дефолтные или задать учебные (например для API 0.01, 0.05, 0.1, 0.3, 1, 3)
### 	Текущее число активных запросов
http_in_flight_requests — gauge<br/>
без Labels

## Конфигурация Prometheus и Docker Compose
```yml
global:
  scrape_interval: 5s  
  evaluation_interval: 5s

scrape_configs:
  - job_name: 'tasks'
    static_configs:
      - targets: ['host.docker.internal:8082'] 
    metrics_path: /metrics
```  

```yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    networks:
      - monitoring

  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_SECURITY_ADMIN_USER=admin
    volumes:
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus
    networks:
      - monitoring

networks:
  monitoring:
    driver: bridge

volumes:
  prometheus_data:
  grafana_data:
```

## Генерируем тестовую нагрузку
```bat
@echo off
echo Generating test load...

for /l %%i in (1,1,50) do (
    curl -s http://localhost:8082/v1/tasks -H "Authorization: Bearer demo-token" > nul
    echo OK: %%i
)

for /l %%i in (1,1,20) do (
    curl -s http://localhost:8082/v1/tasks -H "Authorization: Bearer wrong-token" > nul
    echo ERR: %%i
)

echo Done!
```

### Логи
<img width="1397" height="259" alt="image" src="https://github.com/user-attachments/assets/d00248d4-917f-4c5d-af64-45a90b4ff34c" />

### Вывод метрик /metrics
<img width="760" height="634" alt="image" src="https://github.com/user-attachments/assets/aa2f8520-d0a6-4e97-8ed2-61eb8e8390cb" />

## Запросы в секунду
rate(http_requests_total[5m])
<img width="1234" height="409" alt="image" src="https://github.com/user-attachments/assets/cfb16f3c-332d-4232-aa92-4efe3c92feae" />

## Ошибки
rate(http_requests_total{status=~"4.."}[5m]) 
<img width="1236" height="413" alt="image" src="https://github.com/user-attachments/assets/e767b0fc-18a6-41cb-856e-636e7811cdc8" />

## P95
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[1m]))
На этом графике видно, что в разное время 95% запросов выполняются быстрее 0.005 секунд (ось Y на графике отображает время в секундах)
<img width="1249" height="429" alt="image" src="https://github.com/user-attachments/assets/2ca6fabc-7b8d-4e4d-ae73-0d3c5e1cc94e" />

# Ответы на вопросы
1.	Чем метрики отличаются от логов и зачем нужны оба подхода?
Метрики — числовые данные для мониторинга трендов и alerting, логи — события и ошибки для дебага. Нужны оба.
2.	Чем Counter отличается от Gauge?
Counter только увеличивается, Gauge может расти и уменьшаться.
3.	Почему latency нужно измерять histogram, а не просто средним значением?
Среднее скрывает выбросы, histogram показывает распределение (p95/p99).
4.	Что такое labels и почему опасна высокая кардинальность?
Много уникальных значений labels (например, user_id) раздувает метрики и занимает много памяти.
5.	Зачем нужны p95/p99 и почему среднее может “врать”?
Они показывают реальную производительность для большинства запросов, среднее скрывает "медленные" запросы, но не врёт.
