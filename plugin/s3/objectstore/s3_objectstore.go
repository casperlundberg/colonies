package objectstore

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/colonyos/colonies/pkg/fs"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3ObjectStore implements fs.ObjectStore backed by an S3-compatible service.
//
// Storage layout inside the bucket:
//
//	{colonyName}/objects/{objectName}
//	{colonyName}/.staging/{objectName}/{chunkIndex}
type S3ObjectStore struct {
	mc     *minio.Client
	bucket string
}

// NewS3ObjectStore creates an S3ObjectStore from explicit configuration.
func NewS3ObjectStore(cfg fs.ObjectStoreConfig) (*S3ObjectStore, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: false}

	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:    cfg.TLS,
		Transport: transport,
	})
	if err != nil {
		return nil, fmt.Errorf("s3: failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := mc.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("s3: failed to check bucket: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("s3: failed to create bucket: %w", err)
		}
	}

	return &S3ObjectStore{mc: mc, bucket: cfg.Bucket}, nil
}

func (s *S3ObjectStore) objectKey(colonyName, objectName string) string {
	return colonyName + "/objects/" + objectName
}

func (s *S3ObjectStore) stagingKey(colonyName, objectName string, chunkIndex int) string {
	return colonyName + "/.staging/" + objectName + "/" + strconv.Itoa(chunkIndex)
}

func (s *S3ObjectStore) stagingPrefix(colonyName, objectName string) string {
	return colonyName + "/.staging/" + objectName + "/"
}

func (s *S3ObjectStore) Put(colonyName, objectName string, reader io.Reader, size int64) error {
	key := s.objectKey(colonyName, objectName)
	_, err := s.mc.PutObject(context.Background(), s.bucket, key, reader, size,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return fmt.Errorf("s3: put failed: %w", err)
	}
	return nil
}

func (s *S3ObjectStore) PutChunk(colonyName, objectName string, chunkIndex int, reader io.Reader, size int64) error {
	key := s.stagingKey(colonyName, objectName, chunkIndex)
	_, err := s.mc.PutObject(context.Background(), s.bucket, key, reader, size,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return fmt.Errorf("s3: put chunk failed: %w", err)
	}
	return nil
}

func (s *S3ObjectStore) AssembleChunks(colonyName, objectName string, totalChunks int) error {
	ctx := context.Background()

	// Read all chunks into a buffer, then write as the final object.
	var buf bytes.Buffer
	for i := 0; i < totalChunks; i++ {
		key := s.stagingKey(colonyName, objectName, i)
		obj, err := s.mc.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("s3: assemble failed reading chunk %d: %w", i, err)
		}
		if _, err := io.Copy(&buf, obj); err != nil {
			obj.Close()
			return fmt.Errorf("s3: assemble failed copying chunk %d: %w", i, err)
		}
		obj.Close()
	}

	finalKey := s.objectKey(colonyName, objectName)
	_, err := s.mc.PutObject(ctx, s.bucket, finalKey, &buf, int64(buf.Len()),
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return fmt.Errorf("s3: assemble put failed: %w", err)
	}

	// Clean up staging chunks.
	for i := 0; i < totalChunks; i++ {
		key := s.stagingKey(colonyName, objectName, i)
		_ = s.mc.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	}

	return nil
}

func (s *S3ObjectStore) GetChunkStatus(colonyName, objectName string) ([]int, error) {
	prefix := s.stagingPrefix(colonyName, objectName)
	ctx := context.Background()

	var indices []int
	for obj := range s.mc.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("s3: list chunks failed: %w", obj.Err)
		}
		name := strings.TrimPrefix(obj.Key, prefix)
		idx, err := strconv.Atoi(name)
		if err != nil {
			continue
		}
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	return indices, nil
}

func (s *S3ObjectStore) Get(colonyName, objectName string) (io.ReadCloser, int64, error) {
	key := s.objectKey(colonyName, objectName)
	ctx := context.Background()

	info, err := s.mc.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("s3: stat failed: %w", err)
	}

	obj, err := s.mc.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("s3: get failed: %w", err)
	}

	return obj, info.Size, nil
}

func (s *S3ObjectStore) GetRange(colonyName, objectName string, offset, length int64) (io.ReadCloser, error) {
	key := s.objectKey(colonyName, objectName)
	opts := minio.GetObjectOptions{}
	if err := opts.SetRange(offset, offset+length-1); err != nil {
		return nil, fmt.Errorf("s3: invalid range: %w", err)
	}

	obj, err := s.mc.GetObject(context.Background(), s.bucket, key, opts)
	if err != nil {
		return nil, fmt.Errorf("s3: get range failed: %w", err)
	}
	return obj, nil
}

func (s *S3ObjectStore) Remove(colonyName, objectName string) error {
	key := s.objectKey(colonyName, objectName)
	return s.mc.RemoveObject(context.Background(), s.bucket, key, minio.RemoveObjectOptions{})
}

func (s *S3ObjectStore) RemoveStaging(colonyName, objectName string) error {
	prefix := s.stagingPrefix(colonyName, objectName)
	ctx := context.Background()
	for obj := range s.mc.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return fmt.Errorf("s3: list staging failed: %w", obj.Err)
		}
		if err := s.mc.RemoveObject(ctx, s.bucket, obj.Key, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("s3: remove staging object failed: %w", err)
		}
	}
	return nil
}

func (s *S3ObjectStore) Exists(colonyName, objectName string) bool {
	key := s.objectKey(colonyName, objectName)
	_, err := s.mc.StatObject(context.Background(), s.bucket, key, minio.StatObjectOptions{})
	return err == nil
}

func (s *S3ObjectStore) Stat(colonyName, objectName string) (fs.ObjectInfo, error) {
	key := s.objectKey(colonyName, objectName)
	info, err := s.mc.StatObject(context.Background(), s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return fs.ObjectInfo{}, fmt.Errorf("s3: stat failed: %w", err)
	}
	return fs.ObjectInfo{Size: info.Size}, nil
}

func (s *S3ObjectStore) List(colonyName string) ([]string, error) {
	prefix := colonyName + "/objects/"
	ctx := context.Background()

	var names []string
	for obj := range s.mc.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return nil, fmt.Errorf("s3: list failed: %w", obj.Err)
		}
		name := strings.TrimPrefix(obj.Key, prefix)
		if name != "" {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names, nil
}

func (s *S3ObjectStore) RemoveAll(colonyName string) error {
	prefix := colonyName + "/"
	ctx := context.Background()
	for obj := range s.mc.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return fmt.Errorf("s3: list for remove all failed: %w", obj.Err)
		}
		if err := s.mc.RemoveObject(ctx, s.bucket, obj.Key, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("s3: remove all object failed: %w", err)
		}
	}
	return nil
}

func (s *S3ObjectStore) DiskUsage(colonyName string) (int64, error) {
	prefix := colonyName + "/"
	ctx := context.Background()

	var total int64
	for obj := range s.mc.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
		if obj.Err != nil {
			return 0, fmt.Errorf("s3: list for disk usage failed: %w", obj.Err)
		}
		total += obj.Size
	}
	return total, nil
}
