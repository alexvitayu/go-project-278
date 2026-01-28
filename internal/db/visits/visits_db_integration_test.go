package visits_test

import (
	"code/internal/db/visits"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateVisit(t *testing.T) {
	t.Parallel()
	withTx(t, func(ctx context.Context, q *visits.Queries) {
		visits := CreateTestVisits(t)
		err := q.CreateVisit(ctx, *visits[0])
		require.NoError(t, err)
	})
}

func Test_GetTotalVisits(t *testing.T) {
	t.Parallel()
	withTx(t, func(ctx context.Context, q *visits.Queries) {
		expectedTotal := int64(3)
		visits := CreateTestVisits(t)
		for _, v := range visits {
			err := q.CreateVisit(ctx, *v)
			require.NoError(t, err)
		}
		total, err := q.GetTotalVisits(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedTotal, total)
	})
}

func Test_GetVisits(t *testing.T) {
	t.Parallel()
	params := visits.GetVisitsParams{
		Limit:  2,
		Offset: 0,
	}
	withTx(t, func(ctx context.Context, q *visits.Queries) {
		visits := CreateTestVisits(t)
		expectedIP := "192.168.34.189"
		expectedUserAgent := " Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 YaBrowser/24.1.0.0 Safari/537.36"
		for _, v := range visits {
			err := q.CreateVisit(ctx, *v)
			require.NoError(t, err)
		}
		rows, err := q.GetVisits(ctx, params)
		require.NoError(t, err)
		assert.Equal(t, expectedUserAgent, rows[0].UserAgent)
		assert.Equal(t, expectedIP, rows[1].Ip)
	})
}
