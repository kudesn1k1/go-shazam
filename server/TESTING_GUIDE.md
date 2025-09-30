# Руководство по тестированию в Go

## Основные принципы тестирования в Go

### Структура тестов

1. **Файлы тестов** должны заканчиваться на `_test.go`
2. **Пакет** должен быть таким же, как у тестируемого кода
3. **Функции тестов** должны начинаться с `Test` и принимать параметр `*testing.T`

### Стандартные библиотеки для тестирования

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)
```

### Структура теста (Arrange-Act-Assert)

```go
func TestFunctionName_Scenario(t *testing.T) {
    // Arrange - подготовка данных и моков
    mockService := new(MockService)
    mockService.On("Method", args...).Return(expectedResult, nil)

    service := NewService(mockService)

    // Act - выполнение тестируемого кода
    result, err := service.Method(context.Background(), input)

    // Assert - проверка результатов
    assert.NoError(t, err)
    assert.Equal(t, expectedResult, result)
    mockService.AssertExpectations(t)
}
```

## Виды тестов

### 1. Unit тесты
Тестируют отдельные функции/методы в изоляции с использованием моков.

### 2. Integration тесты
Тестируют взаимодействие между компонентами (например, HTTP handlers с сервисами).

### 3. End-to-End тесты
Тестируют полную функциональность приложения (редко используются в Go).

## Mocking в Go

### testify/mock

```go
// Определение интерфейса
type ServiceInterface interface {
    Method(ctx context.Context, input string) (string, error)
}

// Создание мока
type MockService struct {
    mock.Mock
}

func (m *MockService) Method(ctx context.Context, input string) (string, error) {
    args := m.Called(ctx, input)
    return args.String(0), args.Error(1)
}

// Использование в тесте
func TestHandler_Method(t *testing.T) {
    mockService := new(MockService)
    mockService.On("Method", mock.Anything, "input").Return("output", nil)

    handler := NewHandler(mockService)

    // ... тестирование handler'а

    mockService.AssertExpectations(t)
}
```

## Тестирование HTTP handlers

```go
func TestHandler_CreateUser(t *testing.T) {
    gin.SetMode(gin.TestMode)

    // Создание моков
    mockService := new(MockUserService)

    // Настройка роутера
    router := gin.New()
    handler := NewUserHandler(router, mockService)

    // Создание тестового запроса
    requestBody := CreateUserRequest{Name: "John"}
    jsonBody, _ := json.Marshal(requestBody)

    req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")

    // Создание ResponseRecorder
    w := httptest.NewRecorder()

    // Выполнение запроса
    router.ServeHTTP(w, req)

    // Проверка ответа
    assert.Equal(t, http.StatusCreated, w.Code)

    var response User
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "John", response.Name)
}
```

## Тестирование конфигурации

```go
func TestLoadConfig_Success(t *testing.T) {
    // Установка переменных окружения
    os.Setenv("DATABASE_URL", "test-url")
    defer os.Unsetenv("DATABASE_URL")

    config := LoadConfig()

    assert.Equal(t, "test-url", config.DatabaseURL)
}
```

## Лучшие практики

1. **Изоляция тестов** - каждый тест должен быть независимым
2. **Описательные названия** - `TestFunctionName_Scenario_ExpectedResult`
3. **Моки внешних зависимостей** - базы данных, HTTP клиентов, файловой системы
4. **Тестирование ошибок** - проверка обработки ошибочных ситуаций
5. **Покрытие edge-кейсов** - граничные значения, пустые данные, неверные типы
6. **Чистые тесты** - отсутствие побочных эффектов между тестами

## Запуск тестов

```bash
# Запуск всех тестов
go test ./...

# Запуск тестов в конкретном пакете
go test ./internal/song

# Запуск с покрытием
go test -cover ./...

# Запуск с verbose выводом
go test -v ./...

# Запуск конкретного теста
go test -run TestFunctionName ./internal/song
```

## Покрытие кода

```bash
# Генерация отчета о покрытии
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Распространенные ошибки

1. **Зависимости между тестами** - использование глобальных переменных
2. **Отсутствие очистки** - не удаление созданных файлов/данных
3. **Слишком сложные тесты** - тестирующие слишком много за раз
4. **Отсутствие моков** - тестирование внешних сервисов в unit тестах
5. **Игнорирование ошибок** - не проверка всех возможных ошибок

## Пример структуры тестового файла

```go
package mypackage

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// MockService - мок для внешней зависимости
type MockService struct {
    mock.Mock
}

func (m *MockService) ExternalMethod(ctx context.Context, input string) (string, error) {
    args := m.Called(ctx, input)
    return args.String(0), args.Error(1)
}

func TestMyService_Method_Success(t *testing.T) {
    // Arrange
    mockService := new(MockService)
    mockService.On("ExternalMethod", mock.Anything, "input").Return("output", nil)

    service := NewMyService(mockService)

    // Act
    result, err := service.Method(context.Background(), "input")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "output", result)
    mockService.AssertExpectations(t)
}

func TestMyService_Method_Error(t *testing.T) {
    // Arrange
    mockService := new(MockService)
    expectedError := errors.New("external error")
    mockService.On("ExternalMethod", mock.Anything, "input").Return("", expectedError)

    service := NewMyService(mockService)

    // Act
    result, err := service.Method(context.Background(), "input")

    // Assert
    assert.Error(t, err)
    assert.Equal(t, expectedError, err)
    assert.Empty(t, result)
    mockService.AssertExpectations(t)
}
```
