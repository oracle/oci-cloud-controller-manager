// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package main

import (
	"fmt"
	"log"
	"os"

	baremetal "github.com/oracle/bmcs-go-sdk"
	tt "github.com/oracle/bmcs-go-sdk/test/shared"
)

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
	if client, err = baremetal.NewFromKeyPath(userOCID, tenancyOCID, fingerprint, keyPath, password); err != nil {
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

	// objectID := "test-object"

	// tt.PrintTestHeader("PutObject")
	// var object *baremetal.Object
	// if object, err = client.PutObject(*namespace, bucket2Name, objectID, []byte("test-object-content"), &baremetal.PutObjectOptions{}); err != nil {
	// 	log.Println(err)
	// 	os.Exit(tt.ERR)
	// }
	// fmt.Printf("%+v\n", object)
	// tt.PrintTestFooter()

	// tt.PrintTestHeader("ListObjects")

	// var objects *baremetal.ListObjects
	// if objects, err = client.ListObjects(*namespace, bucket2Name, &baremetal.ListObjectsOptions{}); err != nil {
	// 	log.Println(err)
	// 	os.Exit(tt.ERR)
	// }
	// fmt.Printf("%+v\n", objects)
	// tt.PrintTestFooter()

	// tt.PrintTestHeader("GetObject")
	// var object2 *baremetal.Object
	// if object2, err = client.GetObject(*namespace, bucket2Name, objectID, &baremetal.GetObjectOptions{}); err != nil {
	// 	log.Println(err)
	// 	os.Exit(tt.ERR)
	// }
	// fmt.Printf("%+v\n", object2)
	// tt.PrintTestFooter()

	// tt.PrintTestHeader("DeleteObject")
	// var delObj *baremetal.DeleteObject
	// if delObj, err = client.DeleteObject(*namespace, bucket2Name, objectID, &baremetal.DeleteObjectOptions{}); err != nil {
	// 	log.Println(err)
	// 	os.Exit(tt.ERR)
	// }
	// fmt.Printf("%+v\n", delObj)
	// tt.PrintTestFooter()

	tt.PrintTestHeader("DeleteBucket")
	if err = client.DeleteBucket(bucketName, *namespace, &baremetal.IfMatchOptions{IfMatch: bucket2.ETag}); err != nil {
		log.Println(err)
		os.Exit(tt.ERR)
	}
	tt.PrintTestFooter()

	fmt.Printf("PASS\n")
}
