package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// HeavyCall - это симуляция вашего медленного источника данных (API/БД).
// Его реализация уже готова ниже, менять её не нужно.
type HeavyCall func(ctx context.Context, key string) (string, error)

type request struct {
	wg  sync.WaitGroup
	val string
	err error
}

type Group struct {
	mu      sync.Mutex
	request map[string]*request
}

func NewGroup() *Group {
	return &Group{
		request: make(map[string]*request),
	}
}

// Do принимает ключ и функцию, которая выполняет тяжелый запрос.
func (g *Group) Do(ctx context.Context, key string, fn HeavyCall) (string, error) {

	g.mu.Lock()
	req, ok := g.request[key]
	if ok {
		g.mu.Unlock()
		req.wg.Wait()

		return req.val, req.err
	}

	req = &request{}
	req.wg.Add(1)
	g.request[key] = req
	g.mu.Unlock()
	req.val, req.err = fn(ctx, key)

	g.mu.Lock()
	delete(g.request, key)
	g.mu.Unlock()

	return req.val, req.err
}

// --- Симуляция окружения для проверки (не менять) ---
func fakeBackend(ctx context.Context, key string) (string, error) {
	fmt.Printf("[Backend] Стартовал реальный запрос для ключа: %s\n", key)
	select {
	case <-time.After(1 * time.Second):
		return "data_for_" + key, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func main() {
	g := NewGroup()
	ctx := context.Background()

	res, _ := g.Do(ctx, "user_1", fakeBackend)
	fmt.Println("Результат:", res)
}
