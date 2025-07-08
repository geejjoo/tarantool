package domain

import "time"

type KV struct {
	Key       string     `json:"key"`
	Value     string     `json:"value"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	IsDeleted bool       `json:"is_deleted,omitempty"`
}

type CreateKVRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type UpdateKVRequest struct {
	Value string `json:"value" binding:"required"`
}

type DeleteKVRequest struct {
	SoftDelete bool `json:"soft_delete"`
}

type ListKVResponse struct {
	Items  []*KV `json:"items"`
	Total  int   `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type RestoreKVRequest struct {
	Key string `json:"key" binding:"required"`
}

type KVResponse struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListKVRequest struct {
	Limit  int `json:"limit" validate:"min=1,max=100"`
	Offset int `json:"offset" validate:"min=0"`
}
