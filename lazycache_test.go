package lazycache

import (
	"errors"
	"testing"
	"time"
)

func TestFetchesWhenNotCached(t *testing.T) {
	count := 0
	cache := New(countFetcher(&count), time.Second, 1)
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
}

func TestDoesNotFetchWhenCached(t *testing.T) {
	count := 0
	cache := New(countFetcher(&count), time.Second, 1)
	cache.Get("Hi")
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
}

func TestReturnsCachedAndFetchesLazilyAfterTtl(t *testing.T) {
	count := 0
	cache := New(countFetcher(&count), time.Microsecond, 1)

	cache.Get("Hi")

	// Second get, returns old value (1) and fetches on the background.
	time.Sleep(5 * time.Microsecond)
	v2, _ := cache.Get("Hi")

	if v2.(int) != 1 {
		t.Errorf("expected %+v to equal 1", v2.(int))
	}

	time.Sleep(2 * time.Microsecond)

	if count != 2 {
		t.Errorf("expected %+v to equal 2", count)
	}

	v3, _ := cache.Get("Hi")

	if v3.(int) != 2 {
		t.Errorf("expected %+v to equal 2", v3.(int))
	}
}

func TestDoesNotFetchErrorsUntilExpire(t *testing.T) {
	count := 0
	cache := New(nilFetcher(&count), time.Second, 1)
	cache.Get("Hi")
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
}

func TestFirstFetchOfNilSavesTheObject(t *testing.T) {
	count := 0
	cache := New(nilFetcher(&count), time.Minute, 1)
	obj, exists := cache.Get("Hi") // flush it
	if exists {
		t.Errorf("item should not exist")
	}
	if obj != nil {
		t.Errorf("item should be nil")
	}
}

func TestFetchingNilErasesExistingValue(t *testing.T) {
	count := 0
	cache := New(nilFetcher(&count), time.Minute, 1)
	cache.items["Hi"] = &Item{object: 99, expires: time.Now().Add(-time.Minute)}
	cache.Get("Hi") // flush it
	time.Sleep(2 * time.Microsecond)
	obj, exists := cache.Get("Hi")
	if exists {
		t.Errorf("item should not exist")
	}
	if obj != nil {
		t.Errorf("item should be nil")
	}
}

func TestErrorOnFetchKeepsOldValue(t *testing.T) {
	count := 0
	cache := New(errorFetcher(&count), time.Microsecond, 1)
	cache.items["paul"] = &Item{object: 99, expires: time.Now().Add(-time.Hour)}
	v1, _ := cache.Get("paul")
	if v1.(int) != 99 {
		t.Errorf("expected %+v to equal 99", v1.(int))
	}
}

func countFetcher(count *int) Fetcher {
	return func(id string) (interface{}, error) {
		*count += 1
		return *count, nil
	}
}

func nilFetcher(count *int) Fetcher {
	return func(id string) (interface{}, error) {
		*count += 1
		return nil, nil
	}
}

func slowNilFetcher(count *int) Fetcher {
	return func(id string) (interface{}, error) {
		time.Sleep(10 * time.Microsecond)
		*count += 1
		return nil, nil
	}
}

func errorFetcher(count *int) Fetcher {
	return func(id string) (interface{}, error) {
		return nil, errors.New("oops")
	}
}

func TestGroupStoresMultipleValues(t *testing.T) {
	count := 0
	cache := NewGroup(groupStoreFetcher(&count), time.Second, 1)
	res, _ := cache.Get("Hi")
	if res != count {
		t.Errorf("expected %+v to equal 1", count)
	}
	res, _ = cache.Get("Bye")
	if res != count {
		t.Errorf("expected %+v to equal 1", count)
	}
}

func TestGroupFetchesWhenNotCached(t *testing.T) {
	count := 0
	cache := NewGroup(groupCountFetcher(&count), time.Second, 1)
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
}

func TestGroupDoesNotFetchWhenCached(t *testing.T) {
	count := 0
	cache := NewGroup(groupCountFetcher(&count), time.Second, 1)
	cache.Get("Hi")
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
}

func TestGroupReturnsCachedAndFetchesLazilyAfterTtl(t *testing.T) {
	count := 0
	cache := NewGroup(groupCountFetcher(&count), time.Microsecond, 1)

	cache.Get("Hi")

	// Second get, returns old value (1) and fetches on the background.
	time.Sleep(5 * time.Microsecond)
	v2, _ := cache.Get("Hi")

	if v2.(int) != 1 {
		t.Errorf("expected %+v to equal 1", v2.(int))
	}

	time.Sleep(2 * time.Microsecond)

	if count != 2 {
		t.Errorf("expected %+v to equal 2", count)
	}

	v3, _ := cache.Get("Hi")

	if v3.(int) != 2 {
		t.Errorf("expected %+v to equal 2", v3.(int))
	}
}

func TestGroupFirstFetchOfNilSavesTheObject(t *testing.T) {
	count := 0
	cache := NewGroup(groupNilFetcher(&count), time.Minute, 1)
	obj, exists := cache.Get("Hi") // flush it
	if exists {
		t.Errorf("item should not exist")
	}
	if obj != nil {
		t.Errorf("item should be nil")
	}
}

func TestGroupFetchingNilDoesNotSave(t *testing.T) {
	count := 0
	cache := NewGroup(groupNilFetcher(&count), time.Minute, 1)
	cache.items["Hi"] = &Item{object: 99, expires: time.Now().Add(-time.Minute)}
	cache.Get("Hi") // flush it
	time.Sleep(10 * time.Microsecond)
	obj, exists := cache.Get("Hi")
	if !exists {
		t.Errorf("item should exist")
	}
	if obj == nil {
		t.Errorf("item should not be nil")
	}
}

func TestGroupErrorOnFetchKeepsOldValue(t *testing.T) {
	count := 0
	cache := NewGroup(groupErrorFetcher(&count), time.Microsecond, 1)
	cache.items["paul"] = &Item{object: 99, expires: time.Now().Add(-time.Hour)}
	v1, _ := cache.Get("paul")
	if v1.(int) != 99 {
		t.Errorf("expected %+v to equal 99", v1.(int))
	}
}

func TestSwapGroupOnFetchKeepsOldValue(t *testing.T) {
	count := 0
	cache := New(countFetcher(&count), time.Second, 1)
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
	cache = cache.SwapGroup(groupCountFetcher(&count))
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
}

func TestSwapSingleOnFetchKeepsOldValue(t *testing.T) {
	count := 0
	cache := NewGroup(groupCountFetcher(&count), time.Second, 1)
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
	cache = cache.SwapSingle(countFetcher(&count))
	cache.Get("Hi")
	if count != 1 {
		t.Errorf("expected %+v to equal 1", count)
	}
}

func groupStoreFetcher(count *int) GroupFetcher {
	return func() (*map[string]interface{}, error) {
		group := map[string]interface{}{
			"Hi":  *count,
			"Bye": *count,
		}
		return &group, nil
	}
}

func groupCountFetcher(count *int) GroupFetcher {
	return func() (*map[string]interface{}, error) {
		*count += 1
		group := map[string]interface{}{
			"Hi": *count,
		}
		return &group, nil
	}
}

func groupNilFetcher(count *int) GroupFetcher {
	return func() (*map[string]interface{}, error) {
		*count += 1
		return nil, nil
	}
}

func groupErrorFetcher(count *int) GroupFetcher {
	return func() (*map[string]interface{}, error) {
		return nil, errors.New("oops")
	}
}
