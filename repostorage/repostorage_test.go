package repostorage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

// TODO: Read doS3Tests from a configuration file
var doS3Tests bool = true

const TESTAPP = "stubapp"

func TestRepoStorage(t *testing.T) {

	// An invalid StorageType should fail.
	if _, err := NewRepoStorage("nosuchstoragetype"); err == nil {
		t.Errorf("Storage creation succeeded with invalid storage type")
	} else {
		fmt.Println(err.Error())
	}

	// Exercise the Local storage type.
	if rs, err := NewRepoStorage(KST_LOCAL); err == nil {

		// Test handling of initialization.
		params := make(map[string]string)
		if err := rs.Init(params); err == nil {
			t.Errorf("%s storage initialization succeeded with missing base dir", KST_LOCAL)
		} else {
			fmt.Println(err.Error())
		}
		params["basedir"] = "../data/nosuchdir"
		if err := rs.Init(params); err == nil {
			t.Errorf("%s storage initialization succeeded with nonexistent base dir", KST_LOCAL)
		} else {
			fmt.Println(err.Error())
		}

		// Set up our local base directory, and run tests.
		params["basedir"] = "../data/repository"
		if err := rs.Init(params); err != nil {
			t.Errorf("%s storage initialization failed: %s", KST_LOCAL, err.Error())
		}
		testRepoStorage(t, KST_LOCAL, rs)

	} else {
		t.Errorf("%s storage creation failed: %s", KST_LOCAL, err.Error())
	}

	// Exercise the S3 storage type.
	if rs, err := NewRepoStorage(KST_S3); err == nil {

		// Test handling of initialization.
		params := make(map[string]string)
		if err := rs.Init(params); err == nil {
			t.Errorf("%s storage initialization succeeded with missing AWS region", KST_S3)
		} else {
			fmt.Println(err.Error())
		}
		params["awsregion"] = "nosuchregion"
		if err := rs.Init(params); err == nil {
			t.Errorf("%s storage initialization succeeded with invalid region", KST_S3)
		} else {
			fmt.Println(err.Error())
		}

		// Set up our S3 base directory, and run tests.
		params["awsregion"] = "us-east-1"
		params["bucket"] = "change-pulldeploy-test"
		params["prefix"] = "unittest"
		if err := rs.Init(params); err != nil {
			t.Errorf("%s storage initialization failed: %s", KST_S3, err.Error())
		}
		if doS3Tests {
			testRepoStorage(t, KST_S3, rs)
		}

	} else {
		t.Errorf("%s storage creation failed: %s", KST_S3, err.Error())
	}
}

func testRepoStorage(t *testing.T, st StorageType, rs RepoStorage) {

	sampleBytes := []byte("This is sample repository data.\n")
	sampleFilename1 := "/" + TESTAPP + "/method_a/sampledata.txt"
	sampleFilename2 := "/" + TESTAPP + "/method_b/sampledata.txt"

	// Clear out test data from previous runs.
	switch st {
	case KST_LOCAL:
		os.RemoveAll("../data/repository/" + TESTAPP)
	case KST_S3:
		stS3 := rs.(*stS3)
		stS3.bucket.Del(stS3.makeS3Path(sampleFilename1))
		stS3.bucket.Del(stS3.makeS3Path(sampleFilename2))
	}

	// Reading a nonexistent file should fail.
	if _, err := rs.Get(sampleFilename1); err == nil {
		t.Errorf("%s Get() should have failed for nonexistent file", st)
	} else {
		fmt.Println(err.Error())
	}

	// Write some data to the repo.
	if err := rs.Put(sampleFilename1, sampleBytes); err != nil {
		t.Errorf("%s Put() failed: %s", st, err.Error())
	}

	// Read back the data we wrote.
	if data, err := rs.Get(sampleFilename1); err != nil {
		t.Errorf("%s Get() failed %s", st, err.Error())
	} else {
		if bytes.Compare(sampleBytes, data) != 0 {
			t.Errorf("%s Get() error: expected %q, got %q",
				st, string(sampleBytes), string(data))
		}
	}

	// Reading a nonexistent file should fail.
	if _, err := rs.GetReader(sampleFilename2); err == nil {
		t.Errorf("%s GetReader() should have failed for nonexistent file", st)
	} else {
		fmt.Println(err.Error())
	}

	// Write some data to the repo.
	if err := rs.PutReader(
		sampleFilename2,
		ioutil.NopCloser(bytes.NewReader(sampleBytes)),
		int64(len(sampleBytes)),
	); err != nil {
		t.Errorf("%s PutReader() failed: %s", st, err.Error())
	}

	// Read back the data we wrote.
	if rdr, err := rs.GetReader(sampleFilename2); err != nil {
		t.Errorf("%s GetReader() failed %s", st, err.Error())
	} else {
		data := make([]byte, len(sampleBytes)+16)
		rdr.Read(data)
		rdr.Close()
		if bytes.Compare(sampleBytes, data[:len(sampleBytes)]) != 0 {
			t.Errorf("%s Get() error: expected %q, got %q",
				st, string(sampleBytes), string(data))
		}
	}
}
