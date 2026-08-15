package i18n

import "fmt"

type Language int

const (
	EN Language = iota
	RU
)

var translations = map[Language]map[string]string{
	EN: {
		"windowTitle": "Memory/CPU Stress Tester", "testingTab": "Testing", "advancedTab": "Advanced", "logTab": "Log", "settingsTab": "Settings",
		"systemStatus": "System Status:", "totalMemory": "Total Memory: %d MB", "freeMemory": "Free Memory: %d MB", "testSettings": "Test Settings:",
		"memoryMB": "Memory (MB)", "cpuThreads": "CPU Threads", "alertOnError": "Alert on error", "stopOnError": "Stop on error", "testStatus": "Test Status:",
		"time": "Time: %s", "coverage": "Coverage: %s", "errors": "Errors: %d", "start": "Start", "stop": "Stop", "enableCache": "Enable CPU Cache",
		"rng": "RNG:", "experimental": "Experimental:", "stressFPU": "Stress FPU (Floating Point Unit)", "language": "Language", "invalidNumbers": "please enter valid positive numbers for memory and threads",
		"memExceed": "requested memory (%d MB) exceeds available memory (%d MB)",
	},
	RU: {
		"windowTitle": "Тестер нагрузки памяти/процессора", "testingTab": "Тестирование", "advancedTab": "Дополнительно", "logTab": "Лог", "settingsTab": "Настройки",
		"systemStatus": "Состояние системы:", "totalMemory": "Всего памяти: %d МБ", "freeMemory": "Свободно памяти: %d МБ", "testSettings": "Настройки теста:",
		"memoryMB": "Память (МБ)", "cpuThreads": "Потоки CPU", "alertOnError": "Оповещать при ошибке", "stopOnError": "Остановить при ошибке", "testStatus": "Статус теста:",
		"time": "Время: %s", "coverage": "Покрытие: %s", "errors": "Ошибки: %d", "start": "Старт", "stop": "Стоп", "enableCache": "Использовать кэш CPU",
		"rng": "ГПСЧ:", "experimental": "Экспериментально:", "stressFPU": "Нагрузка на FPU", "language": "Язык", "invalidNumbers": "введите корректные положительные значения памяти и потоков",
		"memExceed": "запрошенная память (%d МБ) превышает доступную (%d МБ)",
	},
}

type Translator struct{ lang Language }

func New(lang Language) *Translator             { return &Translator{lang: lang} }
func (t *Translator) SetLanguage(lang Language) { t.lang = lang }
func (t *Translator) Language() Language        { return t.lang }
func (t *Translator) T(key string, args ...interface{}) string {
	text := translations[t.lang][key]
	if text == "" {
		text = translations[EN][key]
	}
	if len(args) > 0 {
		return fmt.Sprintf(text, args...)
	}
	return text
}
