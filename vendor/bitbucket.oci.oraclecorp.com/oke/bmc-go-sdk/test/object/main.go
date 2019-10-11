// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	baremetal "github.com/MustWin/baremetal-sdk-go"
	tt "github.com/MustWin/baremetal-sdk-go/test/shared"
)

func streamCompareAndClose(f1, f2 io.ReadCloser) bool {
	defer f1.Close()
	defer f2.Close()

	b1, _ := ioutil.ReadAll(f1)
	b2, _ := ioutil.ReadAll(f2)
	if !bytes.Equal(b1, b2) {
		return false
	}
	return true
}

func fileOpen(fileName string) *os.File {
	f, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	return f
}

func fileSize(fileName string) uint64 {
	f, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	size, err1 := f.Seek(0, io.SeekEnd)
	if err1 != nil {
		log.Println(err1)
		os.Exit(tt.ERR)
	}
	return uint64(size)
}

func main() {
	log.SetFlags(log.Lshortfile)
	var err error

	keyPath := tt.TestVals["BAREMETAL_PRIVATE_KEY_PATH"]
	tenancyOCID := tt.TestVals["BAREMETAL_TENANCY_OCID"]
	userOCID := tt.TestVals["BAREMETAL_USER_OCID"]
	fingerprint := tt.TestVals["BAREMETAL_FINGERPRINT"]
	password := tt.TestVals["BAREMETAL_KEY_PASSWORD"]
	compartmentID := tt.TestVals["TEST_COMPARTMENT_ID"]

	var client *baremetal.Client
	if client, err = baremetal.NewClient(userOCID, tenancyOCID, fingerprint, baremetal.PrivateKeyFilePath(keyPath), baremetal.PrivateKeyPassword(password)); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}

	tt.PrintTestHeader("GetNamespace")
	var namespace *baremetal.Namespace
	if namespace, err = client.GetNamespace(); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", *namespace)
	tt.PrintTestFooter()

	// var ns baremetal.Namespace = baremetal.Namespace("test-compartment")
	// namespace := &ns

	tt.PrintTestHeader("CreateBucket")
	var bucketName = "test-bucket"
	var bucketMetadata = map[string]string{"foo": "bar"}
	var bucket *baremetal.Bucket
	if bucket, err = client.CreateBucket(compartmentID, bucketName, *namespace, &baremetal.CreateBucketOptions{}); err != nil {
		log.Println(err)
		// os.Exit(tt.ERR)
	} else {
		fmt.Printf("%+v\n", bucket)
	}
	tt.PrintTestFooter()

	tt.PrintTestHeader("ListBuckets")
	var buckets *baremetal.ListBuckets
	if buckets, err = client.ListBuckets(compartmentID, *namespace, &baremetal.ListBucketsOptions{}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", buckets)
	tt.PrintTestFooter()

	tt.PrintTestHeader("GetBucket")
	var bucket2 *baremetal.Bucket
	if bucket2, err = client.GetBucket(bucketName, *namespace); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", bucket2)
	tt.PrintTestFooter()

	tt.PrintTestHeader("UpdateBucket")
	if bucket2, err = client.UpdateBucket(compartmentID, bucketName, *namespace, &baremetal.UpdateBucketOptions{
		IfMatchOptions: baremetal.IfMatchOptions{IfMatch: bucket2.ETag},
		Name:           bucketName,
		Namespace:      *namespace,
		Metadata:       bucketMetadata,
	}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", bucket2)
	tt.PrintTestFooter()

	objectID := [...]string{"test-object", "README.md"}

	tt.PrintTestHeader("PutObject for " + objectID[0])
	var object *baremetal.Object
	testObjectByte := []byte("test-object-content")
	if object, err = client.PutObject(*namespace, bucketName, objectID[0], testObjectByte, &baremetal.PutObjectOptions{}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", object)
	tt.PrintTestFooter()


	tt.PrintTestHeader("PutObjectAsStream for " + objectID[1])
	if object, err = client.PutObjectAsStream(*namespace, bucketName, objectID[1], fileOpen(objectID[1]), &baremetal.PutObjectOptions{}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", object)
	tt.PrintTestFooter()

	tt.PrintTestHeader("ListObjects")
	var objects *baremetal.ListObjects
	if objects, err = client.ListObjects(*namespace, bucketName, &baremetal.ListObjectsOptions{Fields:"name,size,timeCreated,md5"}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	// Validate ListObjects
	objectList := objects.Objects
	var found bool
	for i := range objectID {
		found = false
		for i1 := range objectList {
			if objectList[i1].Name == objectID[i] {
				found = true
				switch i {
				case 0:   // []byte
					oLen := uint64(len(testObjectByte))
					if objectList[i1].Size != uint64(oLen) {
						log.Println(errors.New(fmt.Sprintf("Invalid size for %s, %d instead of %d", objectID[i], objectList[i1].Size, oLen)))
						os.Exit(tt.ERR)
					}
				case 1:	  // io.ReadCloser
					fLen := fileSize(objectID[i])
					if objectList[i1].Size != fLen {
						log.Println(errors.New(fmt.Sprintf("Invalid size for %s, %d instead of %d", objectID[i], objectList[i1].Size, fLen)))
						os.Exit(tt.ERR)
					}
				}
				break
			}
		}
		if !found {
			log.Println(errors.New("Unable to find %s from List of Objects (+%v)"), objectID[i], objects)
			os.Exit(tt.ERR)
		}
	}
	fmt.Printf("%+v\n", objects)
	tt.PrintTestFooter()

	tt.PrintTestHeader("GetObject " + objectID[0])
	var object2 *baremetal.Object
	if object2, err = client.GetObject(*namespace, bucketName, objectID[0], &baremetal.GetObjectOptions{}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	if !bytes.Equal(testObjectByte, object2.Body) {
		log.Println(errors.New(fmt.Sprintf("PutObject() of %s is not the same as GetObject() of %s!",  testObjectByte, object2.Body)))
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", object2)
	tt.PrintTestFooter()

	tt.PrintTestHeader("GetObjectAsStream " + objectID[1])
	if object2, err = client.GetObjectAsStream(*namespace, bucketName, objectID[1], &baremetal.GetObjectOptions{}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	if !streamCompareAndClose(fileOpen(objectID[1]), object2.BodyAsStream) {
		log.Println(errors.New("PutObjectStream() and GetObjectStream() for " + objectID[1] + " did not yield the same data."))
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", object2)
	tt.PrintTestFooter()

	tt.PrintTestHeader("DeleteObject " + objectID[0])
	var delObj *baremetal.DeleteObject
	if delObj, err = client.DeleteObject(*namespace, bucketName, objectID[0], &baremetal.DeleteObjectOptions{}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", delObj)
	tt.PrintTestFooter()

	tt.PrintTestHeader("DeleteObject " + objectID[1])
	var delObj2 *baremetal.DeleteObject
	if delObj2, err = client.DeleteObject(*namespace, bucketName, objectID[1], &baremetal.DeleteObjectOptions{}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	fmt.Printf("%+v\n", delObj2)
	tt.PrintTestFooter()

	tt.PrintTestHeader("DeleteBucket")
	if err = client.DeleteBucket(bucketName, *namespace, &baremetal.IfMatchOptions{IfMatch: bucket2.ETag}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	tt.PrintTestFooter()

	fmt.Printf("PASS\n")
}
