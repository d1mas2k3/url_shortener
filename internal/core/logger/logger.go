package core_logger

// Файл, который создает и записывает логи

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Чтобы передавать не по "log" 
type loggerContextKey struct {}

var (
    key = loggerContextKey{}
)

// Берем готовый zap.Logger, встраиваем(embedding) в свою структуру с еще одним полем file
type Logger struct {
	*zap.Logger

	file *os.File
}

func ToContext(ctx context.Context, log *Logger) context.Context {
    return context.WithValue(
        ctx,
        key,
        log,
    )
}

// Достаем необходимый контекст, который мы добавляли в других middleware
func FromContext(ctx context.Context) *Logger {
	log, ok := ctx.Value(key).(*Logger) // ctx.Value("log") возвращает any, поэтому мы ->
	if !ok {                              // -> приводим к типу Logger (type assertion)
		panic("no logger in context") // Паника может быть, если неправильно настроен middleware.go
	}
	return log
}

// Принимает конфиг, собирает логгер по шагам (папка → файл → формат → два выхода).
// Возвращает готовый инструмент для логирования. Вызывается один раз при старте
func NewLogger(config Config) (*Logger, error) {

	// Создаем объект уровня логирования. Парсит строку уровня логирования
	zapLvl := zap.NewAtomicLevel()                                     // Atomic - не будет гонок данных
	if err := zapLvl.UnmarshalText([]byte(config.Level)); err != nil { // config.Level - уровень логгирования
		return nil, fmt.Errorf("unmarshal log level: %w", err)
	} // Если уровень логирования неправильный - возвращаем ошибку

	// Создаёт папку (и все родительские папки, если их нет).
	if err := os.MkdirAll(config.Folder, 0755); err != nil { // config.Folder - путь к папке
		return nil, fmt.Errorf("mkdir log folder: %w", err) // 0755 - уровеь доступа
	} // Если создать папку не удалось - возвращаем ошибку

	// timestamp из time.Now() "правильно" берет строку времени
	// logFilePath собирает строку пути к файлу и хроанит в себе
	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
	logFilePath := filepath.Join(
		config.Folder,
		fmt.Sprintf("%s.log", timestamp),
	)

	// Ищем файл по пути logFilePath, если его нет создаем его (os.O_CREATE), открываем его
	// для записи (os.O_WRONLY); logFile хранит объект типа *os.File
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644) // 0644 - права доступа
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	// zapConfig - конфигурация формата логов (как они будут выглядеть: поля, стиль, структура)
	zapConfig := zap.NewDevelopmentEncoderConfig()
	zapConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000")

	// zapEncoder - создаёт форматировщик логов (делает их читаемым текстом для консоли)
	zapEncoder := zapcore.NewConsoleEncoder(zapConfig)

	// Записываем в два места сразу: консоль (os.Stdout) и в файл (logFile) с помощью NewTee()
	core := zapcore.NewTee( // core - это уже вся логика вывода (в консоль + файл)
		zapcore.NewCore(zapEncoder, zapcore.AddSync(os.Stdout), zapLvl),
		zapcore.NewCore(zapEncoder, zapcore.AddSync(logFile), zapLvl),
	)

	// Создает логгер из настроенного core
	zapLogger := zap.New(core, zap.AddCaller()) // AddCaller() - где именно в коде был вызов (файл + строка)

	// Возвращаем заполненную структуру
	return &Logger{
		Logger: zapLogger,
		file:   logFile,
	}, nil
}

// Переопределяем With(), чтобы он возвращал указатель не на *zap.Logger, а на *Logger (наш логгер)
func (l *Logger) With(field ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(field...),
		file:   l.file,
	}
}

// Когда приложение завершается, закрывает файл чтобы всё дозаписалось и ресурсы освободились
func (l *Logger) Close() {
	if err := l.file.Close(); err != nil {
		fmt.Println("failed to close application logger", err)
	}
}