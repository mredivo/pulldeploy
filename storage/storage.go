// Package storage provides access to repository storage.
package storage

import (
	"fmt"
	"io"
)

// Params is how storage-type-specific parameters are passed to New.
type Params map[string]string

// Storage provides methods to set and get repository data.
type Storage interface {
	init(params Params) error                                        // Set up access parameters
	Get(repoPath string) ([]byte, error)                             // Retrieve data from a repository file
	Put(repoPath string, data []byte) error                          // Write data to a repository file
	GetReader(repoPath string) (io.ReadCloser, error)                // Open a stream to read a repository file
	PutReader(repoPath string, rc io.ReadCloser, length int64) error // Write a stream to a repository file
	Delete(repoPath string) error                                    // Delete a repository file
}

// AccessMethod indicates where the repository data should be stored.
type AccessMethod string

// String returns a printable representation of an AccessMethod.
func (am AccessMethod) String() string {
	return string(am)
}

// New returns an instance of Storage of the requested type.
func New(am AccessMethod, params Params) (Storage, error) {

	var stg Storage

	switch am {
	case KST_LOCAL:
		stg = &stLocal{}
	case KST_S3:
		stg = &stS3{}
	default:
		return nil, fmt.Errorf("Invalid AccessMethod: %s", am.String())
	}

	if err := stg.init(params); err != nil {
		return nil, err
	}

	return stg, nil
}
