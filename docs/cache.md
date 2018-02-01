Build Cache
===========
Hydroponics includes a build cache backed by S3 called `s3cache`. It is
intended to run inside of or as a sidecar to a build container. It supports the
[Bazel REST protocol][api].

Cache Behavior
--------------
The `s3cache` tool uses S3 similarly to an LRU cache. The S3 bucket is
configured with a lifecycle rule which deletes items which have reached a
certain age. As `s3cache` performs `GetObject` requests it will refresh the
accessed object to reset its expiration time. This allows items which are
accessed frequently to stay in the cache.

Bazel uses two types of caches: a content-addressable store (CAS) and an action
cache (AC). The `s3cache` can be configured to store these in independent S3
buckets or in the same bucket. Their key prefixes are also configurable.

The `s3cache` uploads and downloads S3 objects in parallel. This allows
`s3cache` to be highly performant When deployed in AWS. The primary bottleneck
lies between Bazel and `s3cache`. This is why `s3cache` is intended to be run
on the same physical instance as Bazel.

Setting Up S3
-------------
This configuration will create a single bucket with a 7 day expiration
lifecycle. The lifecycle will also remove failed multipart uploads. Cached
objects will be kept under the `/cache/` prefix. We will use the AWS CLI to
interact with S3.

The lifecycle we will create has the following configuration:

	{
		"Rules": [
			{
				"ID": "multipart", 
				"Prefix": "/",
				"Status": "Enabled",
				"AbortIncompleteMultipartUpload": {
					"DaysAfterInitiation": 1
				}
			}, 
			{
				"ID": "cache",
				"Prefix": "/cache/",
				"Status": "Enabled",
				"Expiration": {
					"Days": 7
				} 
			}
		]
	}

Create the bucket:

    $ aws s3 mb s3://s3cache.example.com
    make_bucket: s3cache.example.com

Create the lifecycle:

	# aws s3api put-bucket-lifecycle \
		--bucket s3cache.example.com \
		--lifecycle-configuration '{"Rules":[{"ID":"multipart","Prefix":"/","Status":"Enabled","AbortIncompleteMultipartUpload":{"DaysAfterInitiation":1}},{"ID":"cache","Prefix":"/cache/","Status":"Enabled","Expiration":{"Days":7}}]}'

Now the bucket is ready to store cahced items.

Configure `s3cache`
-------------------
The `s3cache` is configured using environment variables. The following
variablea are recognized:

| Variable     | Description |
| ------------ | ------------------------------------------------------------------- |
| `CAS_BUCKET` | Name of the S3 bucekt for CAS objects. Required.                    |
| `CAS_PREFIX` | Key prefix for CAS cache objects. Defaults to "".                   |
| `AC_BUCKET`  | Name of the S3 bucket for AC objects. Required.                     |
| `AC_PREFIX`  | Key prefix for AC cache objects. Defaults to "".                    |
| `S3_TIMEOUT` | Time after which an S3 request time out. Defaults to 0s (disabled). |
| `LISTEN`     | The `host:port` to listen on. Defaults to `:80`.                    |
| `LOG_LEVEL`  | Log level. Valid values are `info` and `debug`. Defaults to `info`. |

The `s3cache` uses the AWS SDK internally. This allows it to seemlessly use EC2
or ECS IAM credentials. It also recognizes the standard AWS credential files
and environment variables.

To run `s3cache` with the bucket configuration in the previous section:

	$ CAS_BUCKET=s3cache.example.com \
	    CAS_PREFIX=/cache/cas/ \
	    AC_BUCKET=s3cache.example.com \
	    AC_PREFIX=/cache/ac/ \
        ./s3cache
    {"address":":http","level":"info","message":"start http server"}

The `s3cache` will run in the foreground until stopped.

[api]: https://github.com/bazelbuild/bazel/blob/master/src/main/java/com/google/devtools/build/lib/remote/README.md "Bazel Cache API"
