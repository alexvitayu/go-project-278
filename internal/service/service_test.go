package service_test

import (
	"code/internal/config"
	"code/internal/db/postgres_db"
	"code/internal/service"
	"code/internal/service/mocks"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const baseUrl = "http://localhost:8081"

func TestLinkService_CreateShortLink_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	ctx := context.Background()
	m := new(mocks.MockQuerier)
	originalUrl := "https://example.com/very-very-long?query=long"
	shortName := "test"
	expectedShortUrl := baseUrl + "/" + shortName

	m.On("GetLinks", ctx).Return([]postgres_db.GetLinksRow{}, nil).Once()

	m.On("CreateLink", ctx, postgres_db.CreateLinkParams{
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    expectedShortUrl,
	}).Return(postgres_db.CreateLinkRow{
		ID:          1,
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    expectedShortUrl,
	}, nil).Once()

	s := service.NewLinkService(m, &config.AppConfig{
		BaseURL: baseUrl,
	})
	// Act
	link, err := s.CreateShortLink(ctx, shortName, originalUrl)
	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(1), link.ID)
	assert.Equal(t, expectedShortUrl, link.ShortUrl)

	m.AssertExpectations(t)
}

func TestLinkService_GetLinks(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		m := new(mocks.MockQuerier)

		mockedRows := []postgres_db.GetLinksRow{
			{
				ID:          1,
				OriginalUrl: "https://example1.com/very-very-long?query=long",
				ShortName:   "test1",
				ShortUrl:    baseUrl + "/test1",
			},
			{
				ID:          2,
				OriginalUrl: "https://example2.com/very-very-long?query=long",
				ShortName:   "test2",
				ShortUrl:    baseUrl + "/test2",
			},
		}

		m.On("GetLinks", ctx).Return(mockedRows, nil).Once()

		s := service.NewLinkService(m, &config.AppConfig{
			BaseURL: baseUrl,
		})

		links, err := s.GetLinks(ctx)
		require.NoError(t, err)
		require.Equal(t, len(mockedRows), len(links))
		assert.Equal(t, mockedRows[0].ID, links[0].ID)
		assert.Equal(t, mockedRows[1].ShortUrl, links[1].ShortUrl)
	})
	t.Run("returns empty list when no links", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		m := new(mocks.MockQuerier)

		mockedRows := []postgres_db.GetLinksRow{}

		m.On("GetLinks", ctx).Return(mockedRows, nil).Once()

		s := service.NewLinkService(m, &config.AppConfig{})

		links, err := s.GetLinks(ctx)
		require.NoError(t, err)
		require.Empty(t, links)
		m.AssertExpectations(t)
	})
}

func TestLinkService_GetLinkByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := new(mocks.MockQuerier)
	linkID := int64(10)

	mockedRow := postgres_db.GetLinkByIDRow{
		ID:          linkID,
		OriginalUrl: "https://example1.com/very-very-long?query=long",
		ShortName:   "test1",
		ShortUrl:    baseUrl + "/test1",
	}

	m.On("GetLinkByID", ctx, linkID).Return(mockedRow, nil).Once()

	s := service.NewLinkService(m, &config.AppConfig{
		BaseURL: baseUrl,
	})
	link, err := s.GetLinkByID(ctx, linkID)
	require.NoError(t, err)
	assert.Equal(t, linkID, link.ID)
	assert.Equal(t, mockedRow.ShortUrl, link.ShortUrl)
	m.AssertExpectations(t)
}

func TestLinkService_UpdateLinkByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := new(mocks.MockQuerier)
	linkID := int64(11)
	newShortName := "new_test"
	expectedNewShortUrl := baseUrl + "/" + newShortName

	oldRow := postgres_db.GetLinkByIDRow{
		ID:          linkID,
		OriginalUrl: "https://example1.com/very-very-long?query=long",
		ShortName:   "test1",
		ShortUrl:    baseUrl + "/test1",
	}

	updatedRow := postgres_db.UpdateLinkByIDRow{
		ID:          linkID,
		OriginalUrl: "https://example1.com/very-very-long?query=long",
		ShortName:   newShortName,
		ShortUrl:    baseUrl + "/" + newShortName,
	}

	m.On("GetLinkByID", ctx, linkID).Return(oldRow, nil).Once()
	m.On("UpdateLinkByID", ctx, postgres_db.UpdateLinkByIDParams{
		OriginalUrl: oldRow.OriginalUrl,
		ShortName:   newShortName,
		ShortUrl:    baseUrl + "/" + newShortName,
		ID:          linkID,
	}).Return(updatedRow, nil).Once()

	s := service.NewLinkService(m, &config.AppConfig{
		BaseURL: baseUrl,
	})

	link, err := s.UpdateLinkByID(ctx, newShortName, oldRow.OriginalUrl, linkID)

	require.NoError(t, err)
	assert.Equal(t, newShortName, link.ShortName)
	assert.Equal(t, expectedNewShortUrl, link.ShortUrl)
	m.AssertExpectations(t)

}

func TestLinkService_DeleteLinkByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := new(mocks.MockQuerier)
	linkID := int64(20)
	affectedRows := int64(1)

	m.On("DeleteLinkByID", ctx, linkID).Return(affectedRows, nil).Once()

	s := service.NewLinkService(m, &config.AppConfig{})

	deleted, err := s.DeleteLinkByID(ctx, linkID)
	require.NoError(t, err)
	assert.Equal(t, affectedRows, deleted)
	m.AssertExpectations(t)
}

func TestGenerateShortName(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		size int
		want int
	}{
		{name: "len is 4", size: 4, want: 4},
		{name: "len is 6", size: 6, want: 6},
	}
	for _, tc := range testCases {
		tc := tc // создаём копию
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			str := service.GenerateShortName(tc.size)
			assert.Equal(t, tc.want, len(str))
		})
	}

}
