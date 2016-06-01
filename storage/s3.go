package storage

import (
	"fmt"
	"io"
	"path"

	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/s3"
)

/*
Repository data is stored in Amazon S3.

Valid Params for KST_S3:

	* "awsregion"   The code for the AWS Region, for example "us-east-1"
	* "bucket"      The name of the AWS bucket
	* "prefix"      An optional prefix for the bucket contents, for example "pulldeploy"
*/
const KST_S3 AccessMethod = "s3"

// stS3 is used for PullDeploy repositories in Amazon S3.
type stS3 struct {
	regionName string     // Name of the AWS Region with our bucket
	bucketName string     // Name of the S3 bucket
	pathPrefix string     // Optional prefix to namespace our bucket
	bucket     *s3.Bucket // Handle to the S3 bucket
}

// Initialize the repository object.
func (st *stS3) init(params Params) error {

	// Extract the AWS region name.
	if regionName, ok := params["awsregion"]; ok {
		st.regionName = regionName
	}

	// Extract the AWS bucket name.
	if bucketName, ok := params["bucket"]; ok {
		st.bucketName = bucketName
	}

	// Extract the optional prefix for our paths.
	if pathPrefix, ok := params["prefix"]; ok {
		st.pathPrefix = pathPrefix
	}

	// Validate the region.
	region, ok := aws.Regions[st.regionName]
	if !ok {
		validSet := ""
		for k := range aws.Regions {
			validSet += " " + k
		}
		return fmt.Errorf("Invalid AWS region name: '%s' Valid values:%s",
			st.regionName, validSet)
	}

	// Pull AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY out of the environment.
	auth, err := aws.EnvAuth()
	if err != nil {
		return err
	}

	// Open a handle to the bucket.
	s := s3.New(auth, region)
	st.bucket = s.Bucket(st.bucketName)

	return nil
}

// Get fetches the contents of a repository file into a byte array.
func (st *stS3) Get(repoPath string) ([]byte, error) {
	return st.bucket.Get(st.makeS3Path(repoPath))
}

// Put writes the contents of a byte array into a repository file.
func (st *stS3) Put(repoPath string, data []byte) error {
	options := s3.Options{}
	return st.bucket.Put(
		st.makeS3Path(repoPath),
		data,
		"application/octet-stream",
		"authenticated-read",
		options,
	)
}

// GetReader returns a stream handle for reading a repository file.
func (st *stS3) GetReader(repoPath string) (io.ReadCloser, error) {
	return st.bucket.GetReader(st.makeS3Path(repoPath))
}

// PutReader writes a stream to a repository file.
func (st *stS3) PutReader(repoPath string, rc io.ReadCloser, length int64) error {
	options := s3.Options{}
	return st.bucket.PutReader(
		st.makeS3Path(repoPath),
		rc,
		length,
		"application/octet-stream",
		"authenticated-read",
		options,
	)
}

// Delete removes a repository file.
func (st *stS3) Delete(repoPath string) error {
	return st.bucket.Del(st.makeS3Path(repoPath))
}

// Utility helper to generate a full S3 repository path.
func (st *stS3) makeS3Path(repoPath string) string {
	if st.pathPrefix == "" {
		return repoPath
	}
	return path.Join(st.pathPrefix, repoPath)
}
