# afero-s3

Afero S3 is a Afero FS interface for Amazon s3

## Install

	go get github.com/spf13/afero
	go get github.com/qugit/aferos3

## How to use

	AferoS3 works as a filesystem for afero to work on top of, all of the standard functions are accounted for.
	Though it is worth noting that S3 does not have a directory file structure, so mkdir etc. have no effect.

#### Initialising the library

	AferoS3 uses the aws sdk for authentication and S3 access. Therefore in order to initialise the filesystem
	you only need to pass through an aws-sdk session and the bucket name that you would like to act as the base
	of the filesystem

	```go
	// import afero, aferos3, and the aws sdk
	import (
		"github.com/aws/aws-sdk-go/aws"
		"github.com/aws/aws-sdk-go/aws/session"
		"github.com/spf13/afero"
		"github.com/qugit/aferos3"
	)

	// Generate your aws sdk session
	sesh := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")
	}))

	// Create the bucket file system, if the bucket doesn't exist this will create it
	// This will also error out if you don't have read permissions on the bucket
	afero.Fs, err := aferos3.GetBucketFs("bucket_name", sesh)
	if err != nil {
		fmt.Println(err)
		return
	}

	```

## TODO

- Write tests for more than just, 'can this open a file'
- Write a more comprehensive Readme so that this is accessible
- Consider what a multi-bucket filesystem would look like
- Implement multi-part uploads and s3manager in order to make this more robust

## Contributing

Please feel free to open a pull request, I would be very happy to listen to any
thoughts on features or how to make the current code better. I'd also love any
help with documentation!

## Acknowledgement

This was forked from @chonthu's git repo from four years ago. While much of the
original source code has been replaced, I would really like to thank them for 
their hard work and the outline that it gave me to get this out relatively
quickly.