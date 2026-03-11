// Package main реализует multichecker для статического анализа кода проекта.
//
// Запуск:
//
//	go run ./cmd/staticlint ./...
//
// Включает:
//   - стандартные анализаторы из golang.org/x/tools/go/analysis/passes
//   - все SA-анализаторы и анализаторы класса S1 из staticcheck.io
//   - errcheck — поиск необработанных ошибок
//   - ineffassign — поиск неэффективных присваиваний
//   - osexit — запрещает прямой вызов os.Exit в функции main пакета main
package main
