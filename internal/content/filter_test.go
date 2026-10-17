package content

import (
	"context"
	"strings"
	"testing"
)

func TestListRejectsInvalidLimit(t *testing.T) {
	service := NewService(nil)

	_, err := service.List(context.Background(), ListFilter{
		Limit: 101,
	})
	if err == nil {
		t.Fatal("expected error for limit greater than 100")
	}
}

func TestListRejectsNegativeOffset(t *testing.T) {
	service := NewService(nil)

	_, err := service.List(context.Background(), ListFilter{
		Limit:  20,
		Offset: -1,
	})
	if err == nil {
		t.Fatal("expected error for negative offset")
	}
}

func TestListRejectsInvalidSort(t *testing.T) {
	service := NewService(nil)

	_, err := service.List(context.Background(), ListFilter{
		Limit: 20,
		Sort:  "DROP TABLE content",
	})
	if err == nil {
		t.Fatal("expected error for invalid sort")
	}
}

func TestListRejectsInvalidOrder(t *testing.T) {
	service := NewService(nil)

	_, err := service.List(context.Background(), ListFilter{
		Limit: 20,
		Order: "random",
	})
	if err == nil {
		t.Fatal("expected error for invalid order")
	}
}

func TestListRejectsLongSearch(t *testing.T) {
	service := NewService(nil)

	_, err := service.List(context.Background(), ListFilter{
		Search: strings.Repeat("a", 101),
		Limit:  20,
	})
	if err == nil {
		t.Fatal("expected error for search longer than 100 characters")
	}
}
