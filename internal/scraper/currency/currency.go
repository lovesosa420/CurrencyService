package currency

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
)

var (
	ErrNoSuchCurrency = errors.New("no such currency")
	ErrLongOperation  = errors.New("operation too long")
)

func GetCurrencyCourse(ctx context.Context, currency string) (string, error) {
	c := colly.NewCollector()

	var course string
	c.OnHTML("table.data tbody tr", func(e *colly.HTMLElement) {
		code := e.ChildText("td:nth-child(2)")
		if code == currency {
			course = e.ChildText("td:nth-child(5)")
		}
		return
	})

	result := make(chan error)
	go func() {
		err := c.Visit("https://www.cbr.ru/currency_base/daily/")
		result <- err
	}()

	select {
	case err := <-result:
		if err != nil {
			return "", err
		}
	case <-ctx.Done():
		return "", fmt.Errorf("%w: %w", ErrLongOperation, ctx.Err())
	}

	if course == "" {
		return "", ErrNoSuchCurrency
	}
	return course, nil
}
