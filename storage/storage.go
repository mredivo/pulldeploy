// Package storage provides access to repository storage.
package storage

import (
	"fmt"
	"io"
)

// Params is our way of making Init() polymorphic.
type Params map[string]string

// Storage provides methods to set and get repository data.
type Storage interface {
	Init(params Params) error                                        // Set up access parameters
	Get(repoPath string) ([]byte, error)                             // Retrieve data from a repository file
	Put(repoPath string, data []byte) error                          // Write data to a repository file
	GetReader(repoPath string) (io.ReadCloser, error)                // Open a stream to read a repository file
	PutReader(repoPath string, rc io.ReadCloser, length int64) error // Write a stream to a repository file
}

// StorageType indicates where the repository data should be stored.
type StorageType string

// String returns a printable representation of a StorageType.
func (st StorageType) String() string {
	return string(st)
}

// NewStorage returns an instance of Storage of the requested type.
func NewStorage(st StorageType) (Storage, error) {
	switch st {
	case KST_LOCAL:
		return &stLocal{}, nil
	case KST_S3:
		return &stS3{}, nil
	default:
		return nil, fmt.Errorf("Invalid StorageType: %s", st.String())
	}
}
