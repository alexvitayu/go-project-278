package service

import (
	"code/internal/config"
	store "code/internal/db/postgres_db"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
)

type Link struct {
	ID          int64  `json:"id"`
	OriginalUrl string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortUrl    string `json:"short_url"`
}

type CreateLinkInput struct {
	OriginalUrl string `json:"original_url"`
	ShortName   string `json:"short_name"`
}

// ErrNotFound возвращается, если запись отсутствует.
var ErrNotFound = errors.New("product not found")

type LinkServer interface {
	CreateShortLink(ctx context.Context, shortName, originalUrl string) (*Link, error)
	GetLinks(ctx context.Context, limit, offset int32) ([]*Link, int64, error)
	GetLinkByID(ctx context.Context, id int64) (*Link, error)
	UpdateLinkByID(ctx context.Context, shortName, originalUrl string, id int64) (*Link, error)
	DeleteLinkByID(ctx context.Context, id int64) (int64, error)
	GetOriginalURLByShortName(ctx context.Context, shortName string) (*Link, error)
}

// LinkService инкапсулирует работу с sqlc-запросами.
type LinkService struct {
	q   store.Querier
	cfg *config.AppConfig
}

// NewLinkService конструирует сервис поверх sqlc-слоя.
func NewLinkService(q store.Querier, config *config.AppConfig) *LinkService {
	return &LinkService{
		q:   q,
		cfg: config,
	}
}

// CreateShortLink создаёт короткий url
func (l *LinkService) CreateShortLink(ctx context.Context, shortName, originalUrl string) (*Link, error) {
	links, err := l.q.GetLinks(ctx, store.GetLinksParams{})
	if err != nil {
		return &Link{}, fmt.Errorf("createShortLink: %w", err)
	}
	for _, l := range links {
		if shortName == l.ShortName {
			return &Link{}, fmt.Errorf("short_name already exists")
		}
	}
	baseUrl := l.cfg.BaseURL
	shortUrl := baseUrl + "/" + shortName

	params := store.CreateLinkParams{
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    shortUrl,
	}

	row, err := l.q.CreateLink(ctx, params)
	if err != nil {
		return &Link{}, fmt.Errorf("createShortLink: %w", err)
	}
	out := &Link{
		ID:          row.ID,
		OriginalUrl: row.OriginalUrl,
		ShortName:   row.ShortName,
		ShortUrl:    row.ShortUrl,
	}
	return out, nil
}

// GetLinks возвращает все объекты из БД
func (l *LinkService) GetLinks(ctx context.Context, limit, offset int32) ([]*Link, int64, error) {
	_ = limit
	_ = offset

	rows, err := l.q.GetLinks(ctx, store.GetLinksParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, fmt.Errorf("getLinks: %w", ErrNotFound)
		}
		return nil, 0, fmt.Errorf("getLinks: %w", err)
	}
	out := make([]*Link, 0, len(rows))
	for _, row := range rows {
		link := &Link{
			ID:          row.ID,
			OriginalUrl: row.OriginalUrl,
			ShortName:   row.ShortName,
			ShortUrl:    row.ShortUrl,
		}
		out = append(out, link)
	}
	total, err := l.q.GetTotalLinks(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("getTotalLinks: %w", err)
	}
	return out, total, nil
}

func (l *LinkService) GetLinkByID(ctx context.Context, id int64) (*Link, error) {
	row, err := l.q.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &Link{}, ErrNotFound
		}
		return &Link{}, fmt.Errorf("getLinkByID: %w", err)
	}
	out := Link{
		ID:          row.ID,
		OriginalUrl: row.OriginalUrl,
		ShortName:   row.ShortName,
		ShortUrl:    row.ShortUrl,
	}
	return &out, nil
}

func (l *LinkService) UpdateLinkByID(ctx context.Context, shortName, originalUrl string, id int64) (*Link, error) {
	link, err := l.q.GetLinkByID(ctx, id)
	if err != nil {
		return &Link{}, fmt.Errorf("updateShortLink: %w", err)
	}
	if link.ShortName == shortName {
		return &Link{}, fmt.Errorf("short_name already exists")
	}

	baseUrl := l.cfg.BaseURL
	shortUrl := baseUrl + "/" + shortName

	params := store.UpdateLinkByIDParams{
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    shortUrl,
		ID:          id,
	}

	row, err := l.q.UpdateLinkByID(ctx, params)
	if err != nil {
		return &Link{}, fmt.Errorf("updateLinkByID: %w", err)
	}

	out := &Link{
		ID:          row.ID,
		OriginalUrl: row.OriginalUrl,
		ShortName:   row.ShortName,
		ShortUrl:    row.ShortUrl,
	}
	return out, nil
}

func (l *LinkService) GetOriginalURLByShortName(ctx context.Context, shortName string) (*Link, error) {
	link, err := l.q.GetOriginalURLByShortName(ctx, shortName)
	if err != nil {
		return &Link{}, fmt.Errorf("getOriginalURLByShortName: %w", err)
	}
	out := &Link{
		ID:          link.ID,
		OriginalUrl: link.OriginalUrl,
		ShortName:   shortName,
	}
	return out, nil
}

func (l *LinkService) DeleteLinkByID(ctx context.Context, id int64) (int64, error) {
	n, err := l.q.DeleteLinkByID(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("deleteLinkByID: %w", err)
	}

	return n, nil
}

func GenerateShortName(size int) (string, error) {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789"
	buffer := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, buffer); err != nil {
		return "", err
	}
	result := make([]byte, size)
	for i, b := range buffer {
		result[i] = chars[int(b)%len(chars)]
	}
	return string(result), nil
}
