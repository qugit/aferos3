package aferos3

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/afero"
)

const BucketName = "testbucket.s3fs"
const TestKey = "a/test/key.ext"
const TestObject = "This is some random text to go in here"

func TestMain(m *testing.M) {
	sesh := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	client := s3.New(sesh)

	// Create the bucket
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(BucketName),
	})

	if err != nil {
		os.Exit(1)
	}

	// seed the bucket with the test object
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(TestKey),
		Body:   bytes.NewReader([]byte(TestObject)),
	})

	m.Run()

	// remove the seeded test object
	client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(BucketName),
		Key:    aws.String(TestKey),
	})

	// Delete the bucket
	client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(BucketName),
	})
}

func TestOpenFile(t *testing.T) {

	sesh := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))
	sfs, err := GetBucketFs(BucketName, sesh)
	if err != nil {
		t.Error(err)
	}

	file, err := sfs.Open(TestKey)

	if err != nil {
		t.Error(err)
	}

	if e, _ := file.Stat(); e == nil {
		t.Error("Corrupted file read")
	}

	var OsFs afero.Fs = afero.OsFs{}
	newFile, err := OsFs.Create("output.jpg")
	io.Copy(newFile, file)
}

func TestNewFile(t *testing.T) {
	// create a new file through the system
	// check that it's there
	// delete the file to clean it up
}

func TestRenameFile(t *testing.T) {
	// create a file
	// make sure that it's there
	// rename the file
	// check that the new file is there
	// delete the new filename to make sure it's clean
}

func TestRemoveFile(t *testing.T) {
	// create a file
	// delete the file
	// remove the file
}

func TestRemoveAll(t *testing.T) {
	// seed a multitude of files in a multitude of paths
	// check that they're there
	// then try deleting them with the remove all function
}
