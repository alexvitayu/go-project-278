// это тесты, которые проверяют непосредственно SQL запросы, сгенерированные sqlc
// работают ли они корректно с реальной БД
package db_test

import (
	"code/internal/db/postgres_db"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_CreateLink(t *testing.T) {
	t.Parallel()
	withTx(t, func(ctx context.Context, q *postgres_db.Queries) {
		links, err := CreateTestLinks(t, ctx, q, BASE_URL)
		require.NoError(t, err)

		assert.Equal(t, BASE_URL+"/test-short1", links[0].ShortUrl)

		getLink, err := q.GetLinkByID(ctx, links[0].ID)
		require.NoError(t, err)
		assert.Equal(t, getLink.ID, links[0].ID)
	})
}

func Test_DeleteLinkByID(t *testing.T) {
	t.Parallel()
	withTx(t, func(ctx context.Context, q *postgres_db.Queries) {
		links, err := CreateTestLinks(t, ctx, q, BASE_URL)
		require.NoError(t, err)

		n, err := q.DeleteLinkByID(ctx, links[0].ID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), n)

		_, err = q.GetLinkByID(ctx, links[0].ID)
		require.Error(t, err)
	})
}

func Test_GetLinkByID(t *testing.T) {
	t.Parallel()
	withTx(t, func(ctx context.Context, q *postgres_db.Queries) {
		links, err := CreateTestLinks(t, ctx, q, BASE_URL)
		require.NoError(t, err)
		got, err := q.GetLinkByID(ctx, links[0].ID)
		require.NoError(t, err)
		assert.Equal(t, links[0].ID, got.ID)
	})
}

func Test_GetLinks(t *testing.T) {
	t.Parallel()
	withTx(t, func(ctx context.Context, q *postgres_db.Queries) {
		links, err := CreateTestLinks(t, ctx, q, BASE_URL)
		require.NoError(t, err)
		got, err := q.GetLinks(ctx, postgres_db.GetLinksParams{
			Limit:  3,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Len(t, got, len(links))
	})
}

func Test_UpdateLinkByID(t *testing.T) {
	t.Parallel()
	withTx(t, func(ctx context.Context, q *postgres_db.Queries) {
		links, err := CreateTestLinks(t, ctx, q, BASE_URL)
		require.NoError(t, err)

		updateParams := postgres_db.UpdateLinkByIDParams{
			ID:          links[1].ID,
			OriginalUrl: "https://example2.net/very-very-long-short-name?with=queries",
			ShortName:   "new_short_name2",
			ShortUrl:    BASE_URL + "/new_short_name2",
		}
		got, err := q.UpdateLinkByID(ctx, updateParams)
		require.NoError(t, err)

		link, err := q.GetLinkByID(ctx, got.ID)
		require.NoError(t, err)
		assert.Equal(t, BASE_URL+"/new_short_name2", got.ShortUrl)
		assert.Equal(t, link.ShortUrl, got.ShortUrl)
		assert.NotEqual(t, links[got.ID].ShortUrl, link.ShortUrl)
	})
}

func Test_GetOriginalURLByShortName(t *testing.T) {
	t.Parallel()
	withTx(t, func(ctx context.Context, q *postgres_db.Queries) {
		_, err := CreateTestLinks(t, ctx, q, BASE_URL)
		require.NoError(t, err)
		shortName := "test-short3"
		expectedOriginalURL := "https://example3.net/very-very-long-short-name?with=queries"
		got, err := q.GetOriginalURLByShortName(ctx, shortName)
		require.NoError(t, err)

		assert.Equal(t, expectedOriginalURL, got.OriginalUrl)
	})
}
