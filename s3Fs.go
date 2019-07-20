package aferos3

import (
	"bytes"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/afero/mem"
)

/*

I think here I should create a s3 to file system mapping?

Example:

	"private":                   600,
	"public-read":               664,
	"public-read-write":         666,
	"authenticated-read":        660,
	"bucket-owner-read":         660,
	"bucket-owner-full-control": 666,

*/

// S3Fs is the base filesystem that we will be sending back to Afero
type S3Fs struct {
	session *session.Session
	client  *s3.S3
	bname   string
}

// GetBucketFs initialises the FileSystem and then returns an afero.Fs or an error
// this is how you first initialise S3Fs
func GetBucketFs(name string, sesh *session.Session) (afero.Fs, error) {

	client := s3.New(sesh)

	// Get the ACL of the bucket so that we can test that this bucket actually exists
	_, err := client.GetBucketAcl(&s3.GetBucketAclInput{
		Bucket: aws.String(name),
	})

	if err != nil && strings.Contains(err.Error(), "NoSuchBucket") {
		fmt.Println(err.Error())
		_, err = createBucket(client, name)
	}

	// error should have been overridden if there was no such bucket
	if err != nil {
		return nil, err
	}

	return S3Fs{
		session: sesh,
		client:  client,
		bname:   name,
	}, nil
}

// If the bucket that is being passed through doesn't exist then we should be creating it
func createBucket(client *s3.S3, name string) (string, error) {
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
	})
	return name, err
}

// Name simply returns the filesystem identifier for Afero
func (S3Fs) Name() string {
	return "S3Fs"
}

// Create creates a new file object to be used, this function uses an in memory file
// at this point and does not communicate with S3 until the file has to be opened
func (S3Fs) Create(name string) (afero.File, error) {
	return mem.NewFileHandle(mem.CreateFile(name)), nil
}

// Open will read from s3, and bring down whole file? or torrent?
func (s S3Fs) Open(name string) (afero.File, error) {

	memFile, err := s.Create(getNameFromPath(name))
	if err != nil {
		return nil, err
	}

	res, err := s.client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bname),
		Key:    aws.String(name),
	})
	if err != nil {
		return memFile, err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return memFile, err
	}

	memFile.Write(b)

	return memFile, err
}

// Push executes a Put to the S3 bucket
func (s S3Fs) Push(f afero.File, path string) error {

	body, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	s.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s.bname),
		Key:    aws.String(path),
		Body:   bytes.NewReader(body),
	})

	return err
}

func getNameFromPath(fileName string) string {
	var name string
	tokens := strings.Split(fileName, ".")
	ext := tokens[len(tokens)-1]

	if len(tokens) > 2 {
		name = strings.Join(tokens[:len(tokens)-1], ".")
	} else {
		name = tokens[0]
	}

	return fmt.Sprintf("%s.%s", name, ext)
}

// S3FileInfo gives file information for a file in S3
type S3FileInfo struct {
	os.FileInfo
	file *afero.File
}

// OpenFile returns the same thing as Open for the moment, this may change later
func (s S3Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	file, err := s.Open(name)
	s.Chmod(name, perm)
	return file, err
}

// Chmod doesn't do anything at the moment as it may have unintended consequences
// however it needs to be here to comply with Afero
func (s S3Fs) Chmod(name string, mode os.FileMode) error {
	return nil
}

// Chtimes does absolutely nothing yet
func (S3Fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return nil
}

// Stat provides very light file information, this needs to be augmented later
func (s S3Fs) Stat(name string) (os.FileInfo, error) {
	f, err := s.Open(name)
	return S3FileInfo{file: &f}, err
}

// Rename creates a copy of the file with the new name and then deletes the old file
func (s S3Fs) Rename(oldname, newname string) error {
	// Create a new copy
	// Delete the old file
	_, err := s.client.CopyObject(&s3.CopyObjectInput{
		CopySource: aws.String(fmt.Sprintf("%s/%s", s.bname, oldname)),
		Bucket:     aws.String(s.bname),
		Key:        aws.String(newname),
	})
	if err != nil {
		return err
	}

	return s.Remove(oldname)
}

// Remove deletes the file from S3
func (s S3Fs) Remove(name string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bname),
		Key:    aws.String(name),
	})
	return err
}

// Mkdir does nothing at the moment
func (s S3Fs) Mkdir(name string, perm os.FileMode) error { return nil }

// MkdirAll also does nothing as S3 doesn't really do diretory struture
func (s S3Fs) MkdirAll(path string, perm os.FileMode) error { return nil }

// RemoveAll lists all files in a path and then removes them all one by one
func (s S3Fs) RemoveAll(path string) error {

	// list all of the files for this prefix
	keys := make([]string, 0)
	res, err := s.client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(s.bname),
		Prefix: aws.String(path),
	})
	if err != nil {
		return err
	}
	for _, obj := range res.Contents {
		keys = append(keys, *(obj.Key))
	}
	for res.NextMarker != nil {
		res, err = s.client.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(s.bname),
			Prefix: aws.String(path),
			Marker: res.NextMarker,
		})
		if err != nil {
			return err
		}
		for _, obj := range res.Contents {
			keys = append(keys, *(obj.Key))
		}
	}

	// delete each of them
	for _, key := range keys {
		err = s.Remove(key)
		if err != nil {
			return err
		}
	}

	return nil
}
