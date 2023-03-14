package dwds

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/abdusco/alman/internal/http"
)

func TestDwds_Find(t *testing.T) {
	d := Dwds{
		fetcher: http.ReqFetcher,
	}
	res, err := d.Find(context.Background(), "zuschütten")
	assert.NoError(t, err)
	t.Log(res.String())
}
