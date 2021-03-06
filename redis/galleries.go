package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"strconv"
	"time"

	"github.com/jgsheppa/golang_website/models"
)

func (c *Client) GetGalleryID(ctx context.Context, nconst string) (*models.Gallery, error) {
	cmd := c.client.Get(ctx, nconst)

	cmdb, err := cmd.Bytes()
	if err != nil {
		return &models.Gallery{}, err
	}

	b := bytes.NewReader(cmdb)

	var res models.Gallery

	if err := gob.NewDecoder(b).Decode(&res); err != nil {
		return &models.Gallery{}, err
	}

	return &res, nil
}

func (c *Client) SetGalleryID(ctx context.Context, n *models.Gallery) error {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(n); err != nil {
		return err
	}

	id := strconv.FormatUint(uint64(n.ID), 10)

	return c.client.Set(ctx, id, b.Bytes(), 25*time.Second).Err()
}
