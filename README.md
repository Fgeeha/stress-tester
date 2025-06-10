# stress-tester

Кроссплатформенное приложение для стресс-тестирования CPU и памяти.

## Сборка и запуск

```bash
cd $GOPATH/src/github.com/fgeeha/stress-tester
go mod tidy
go run ./cmd/stress-tester
```

## Функции
- CPU-тест: нагрузка на процессор с простыми вычислениями.
- Memory-тест: записи/чтение различных паттернов.
- Логирование в HTML-отчет.
- Простой GUI на Fyne.
