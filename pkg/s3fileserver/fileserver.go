package s3fileserver

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
)

// Impl - A Impl implements http.FileSystem using the minio client
type Impl struct {
	Client     *minio.Client
	Bucket     string
	BucketPath string
	Logger     *slog.Logger
}

var _ http.FileSystem = &Impl{}

func pathIsDir(ctx context.Context, s3 *Impl, name string) bool {
	name = strings.Trim(name, pathSeparator) + pathSeparator

	listCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	objCh := s3.Client.ListObjects(listCtx,
		s3.Bucket,
		minio.ListObjectsOptions{
			Prefix:  name,
			MaxKeys: 1,
		})
	for range objCh {
		cancel()
		return true
	}
	return false
}

func (s3 *Impl) Open(name string) (http.File, error) {
	name = path.Join(s3.BucketPath, name)
	if name == pathSeparator || pathIsDir(context.Background(), s3, name) {
		return &httpMinioObject{
			client: s3.Client,
			object: nil,
			isDir:  true,
			bucket: s3.Bucket,
			prefix: strings.TrimSuffix(name, pathSeparator),
		}, nil
	}

	name = strings.TrimPrefix(name, pathSeparator)
	obj, err := getObject(context.Background(), s3, name)
	if err != nil {
		return nil, os.ErrNotExist
	}

	return &httpMinioObject{
		client: s3.Client,
		object: obj,
		isDir:  false,
		bucket: s3.Bucket,
		prefix: name,
	}, nil
}

func getObject(ctx context.Context, s3 *Impl, name string) (*minio.Object, error) {
	if strings.Contains(strings.ToLower(strings.TrimSpace(name)), "soap") {
		return nil, os.ErrNotExist
	}
	names := []string{name, name + "/index.html", name + "/index.htm"}
	names = append(names, "/404.html")
	for _, n := range names {
		obj, err := s3.Client.GetObject(ctx, s3.Bucket, n, minio.GetObjectOptions{})
		if err != nil {
			s3.Logger.Info("error fetching object", "error", err)
			continue
		}

		_, err = obj.Stat()
		if err != nil {
			// do not log "file" in bucket not found errors
			if minio.ToErrorResponse(err).Code != "NoSuchKey" {
				s3.Logger.Info("error stat'ing object", "error", err)
			}
			continue
		}

		return obj, nil
	}

	return nil, os.ErrNotExist
}
