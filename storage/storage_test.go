package storage

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

// TODO: Read doS3Tests from a configuration file
var doS3Tests = true

const TESTAPP = "stubapp"

func TestStorage(t *testing.T) {

	var rs Storage
	var err error

	// An invalid AccessMethod should fail.
	params := make(map[string]string)
	if _, err := New("nosuchaccessmethod", params); err == nil {
		t.Errorf("Storage creation succeeded with invalid access method")
	} else {
		fmt.Println(err.Error())
	}

	// Exercise the Local access method.
	params = make(map[string]string)

	// Test handling of initialization.
	if _, err := New(KST_LOCAL, params); err == nil {
		t.Errorf("%s storage initialization succeeded with missing base dir", KST_LOCAL)
	} else {
		fmt.Println(err.Error())
	}
	params["basedir"] = "../data/nosuchdir"
	if _, err := New(KST_LOCAL, params); err == nil {
		t.Errorf("%s storage initialization succeeded with nonexistent base dir", KST_LOCAL)
	} else {
		fmt.Println(err.Error())
	}

	// Set up our local base directory, and run tests.
	params["basedir"] = "../data/repository"
	if rs, err = New(KST_LOCAL, params); err != nil {
		t.Errorf("%s storage initialization failed: %s", KST_LOCAL, err.Error())
	}
	testStorage(t, KST_LOCAL, rs)

	// Exercise the S3 access method.
	params = make(map[string]string)

	// Test handling of initialization.
	if _, err := New(KST_S3, params); err == nil {
		t.Errorf("%s storage initialization succeeded with missing AWS region", KST_S3)
	} else {
		fmt.Println(err.Error())
	}
	params["awsregion"] = "nosuchregion"
	if _, err := New(KST_S3, params); err == nil {
		t.Errorf("%s storage initialization succeeded with invalid region", KST_S3)
	} else {
		fmt.Println(err.Error())
	}

	// Set up our S3 base directory, and run tests.
	params["awsregion"] = "us-east-1"
	params["bucket"] = "change-pulldeploy-test"
	params["prefix"] = "unittest"
	if rs, err = New(KST_S3, params); err != nil {
		t.Errorf("%s storage initialization failed: %s", KST_S3, err.Error())
	}
	if doS3Tests {
		testStorage(t, KST_S3, rs)
	}
}

func testStorage(t *testing.T, am AccessMethod, rs Storage) {

	sampleBytes := []byte("This is sample repository data.\n")
	sampleFilename1 := "/" + TESTAPP + "/method_a/sampledata.txt"
	sampleFilename2 := "/" + TESTAPP + "/method_b/sampledata.txt"

	// Clear out test data from previous runs.
	switch am {
	case KST_LOCAL:
		os.RemoveAll("../data/repository/" + TESTAPP)
	case KST_S3:
		stS3 := rs.(*stS3)
		stS3.bucket.Del(stS3.makeS3Path(sampleFilename1))
		stS3.bucket.Del(stS3.makeS3Path(sampleFilename2))
	}

	// Reading a nonexistent file should fail.
	if _, err := rs.Get(sampleFilename1); err == nil {
		t.Errorf("%s Get() should have failed for nonexistent file", am)
	} else {
		fmt.Println(err.Error())
	}

	// Write some data to the repo.
	if err := rs.Put(sampleFilename1, sampleBytes); err != nil {
		t.Errorf("%s Put() failed: %s", am, err.Error())
	}

	// Read back the data we wrote.
	if data, err := rs.Get(sampleFilename1); err != nil {
		t.Errorf("%s Get() failed %s", am, err.Error())
	} else {
		if bytes.Compare(sampleBytes, data) != 0 {
			t.Errorf("%s Get() error: expected %q, got %q",
				am, string(sampleBytes), string(data))
		}
	}

	// Reading a nonexistent file should fail.
	if _, err := rs.GetReader(sampleFilename2); err == nil {
		t.Errorf("%s GetReader() should have failed for nonexistent file", am)
	} else {
		fmt.Println(err.Error())
	}

	// Write some data to the repo.
	if err := rs.PutReader(
		sampleFilename2,
		ioutil.NopCloser(bytes.NewReader(sampleBytes)),
		int64(len(sampleBytes)),
	); err != nil {
		t.Errorf("%s PutReader() failed: %s", am, err.Error())
	}

	// Read back the data we wrote.
	if rdr, err := rs.GetReader(sampleFilename2); err != nil {
		t.Errorf("%s GetReader() failed %s", am, err.Error())
	} else {
		data := make([]byte, len(sampleBytes)+16)
		rdr.Read(data)
		rdr.Close()
		if bytes.Compare(sampleBytes, data[:len(sampleBytes)]) != 0 {
			t.Errorf("%s Get() error: expected %q, got %q",
				am, string(sampleBytes), string(data))
		}
	}
}
