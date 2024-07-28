// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http).

package litestream

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/benbjohnson/litestream"
	lss3 "github.com/benbjohnson/litestream/s3"
)

func Replicate(ctx context.Context, dsn, replicaURL string) (*litestream.DB, error) {
	if i := strings.Index(dsn, "?"); i > 0 {
		dsn = dsn[0:i]
	}
	dsn = strings.TrimPrefix(dsn, "file:")

	lsdb := litestream.NewDB(dsn)

	u, err := url.Parse(replicaURL)
	if err != nil {
		return nil, err
	}

	host := u.Host
	path := strings.TrimPrefix(path.Clean(u.Path), "/")
	bucket, region, endpoint, forcePathStyle := lss3.ParseHost(host)

	if e := os.Getenv("AWS_ENDPOINT_URL_S3"); e != "" {
		endpoint = e
	}

	if r := os.Getenv("AWS_REGION"); r != "" {
		region = r
	}

	if b := os.Getenv("BUCKET_NAME"); b != "" {
		bucket = b
	}

	if fps := os.Getenv("LITESTREAM_FORCE_PATH_STYLE"); fps != "" {
		if b, err := strconv.ParseBool(fps); err != nil {
			return nil, fmt.Errorf("invalid LITESTREAM_FORCE_PATH_STYLE value: %q", fps)
		} else {
			forcePathStyle = b
		}
	}

	client := lss3.NewReplicaClient()
	client.Bucket = bucket
	client.Path = path
	client.Region = region
	client.Endpoint = endpoint
	client.ForcePathStyle = forcePathStyle
	client.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	client.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	replica := litestream.NewReplica(lsdb, lss3.ReplicaClientType)
	replica.Client = client

	lsdb.Replicas = append(lsdb.Replicas, replica)

	if err := restore(ctx, replica); err != nil {
		return nil, err
	}

	if err := lsdb.Open(); err != nil {
		return nil, err
	}

	if err := lsdb.Sync(ctx); err != nil {
		return nil, err
	}

	return lsdb, nil
}

func restore(ctx context.Context, replica *litestream.Replica) error {
	if _, err := os.Stat(replica.DB().Path()); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	opt := litestream.NewRestoreOptions()
	opt.OutputPath = replica.DB().Path()

	var err error
	if opt.Generation, _, err = replica.CalcRestoreTarget(ctx, opt); err != nil {
		return err
	}

	if opt.Generation == "" {
		return nil
	}

	if err := replica.Restore(ctx, opt); err != nil {
		return err
	}
	return nil
}